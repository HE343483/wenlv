package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wenlv-backend/model"
)

// ============ 行程规划异步任务管理 ============
// 与原项目一致:提交任务立即返回 task_id,后台执行并通过 WebSocket / 轮询推送进度;
// 任务状态持久化到 <dataDir>/trip_tasks/*.json,服务重启后可查看历史计划。

const (
	TripTaskProcessing = "processing"
	TripTaskCompleted  = "completed"
	TripTaskFailed     = "failed"
)

// TripTask 单个行程规划任务。
type TripTask struct {
	TaskID         string
	PlanID         string
	Status         string
	Stage          string
	Progress       int
	Message        string
	Result         *model.TripPlanResponse
	Error          string
	RequestPayload *model.TripRequest
	UpdatedAt      time.Time

	subscribers []chan *model.TripTaskEvent
}

type tripTaskFile struct {
	TaskID         string                  `json:"task_id"`
	PlanID         string                  `json:"plan_id"`
	Status         string                  `json:"status"`
	Stage          string                  `json:"stage"`
	Progress       int                     `json:"progress"`
	Message        string                  `json:"message"`
	Result         *model.TripPlanResponse `json:"result"`
	Error          string                  `json:"error"`
	RequestPayload *model.TripRequest      `json:"request_payload"`
}

// TripTaskStore 任务存储(内存 + MySQL 持久化 + 磁盘兜底)。
type TripTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*TripTask
	dir   string
	db    *gorm.DB // 可选:接入后完成任务落库 trip_plans,历史/回走数据库
}

// AttachDB 接入 MySQL,行程完成后自动落库,历史接口优先读库。
func (s *TripTaskStore) AttachDB(db *gorm.DB) { s.db = db }

// NewTripTaskStore 构造任务存储并预加载历史任务。
func NewTripTaskStore(dataDir string) *TripTaskStore {
	s := &TripTaskStore{
		tasks: map[string]*TripTask{},
		dir:   filepath.Join(dataDir, "trip_tasks"),
	}
	s.loadAll()
	return s
}

// NewTaskID 生成 8 位任务 ID(与原项目 uuid 前 8 位风格一致)。
func NewTaskID() string {
	const chars = "0123456789abcdef"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func (s *TripTaskStore) filePath(taskID string) string {
	return filepath.Join(s.dir, taskID+".json")
}

func (s *TripTaskStore) loadAll() {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}
		var payload tripTaskFile
		if err := json.Unmarshal(raw, &payload); err != nil {
			continue
		}
		taskID := payload.TaskID
		if taskID == "" {
			taskID = strings.TrimSuffix(entry.Name(), ".json")
		}
		task := &TripTask{
			TaskID: taskID, PlanID: firstNonEmpty(payload.PlanID, taskID),
			Status: payload.Status, Stage: payload.Stage, Progress: payload.Progress,
			Message: payload.Message, Result: payload.Result, Error: payload.Error,
			RequestPayload: payload.RequestPayload,
		}
		if info, err := entry.Info(); err == nil {
			task.UpdatedAt = info.ModTime()
		}
		// 服务重启后处理中任务无法恢复执行,标记为失败,避免前端无限等待
		if task.Status != TripTaskCompleted && task.Status != TripTaskFailed {
			task.Status = TripTaskFailed
			task.Stage = TripTaskFailed
			task.Progress = 100
			task.Error = "服务已重启,未完成的旅行规划任务无法恢复,请重新生成。"
			task.Message = task.Error
		}
		s.tasks[taskID] = task
		loaded++
	}
	if loaded > 0 {
		fmt.Printf("📦 已加载 %d 个持久化旅行任务\n", loaded)
	}
}

// Create 创建任务。
func (s *TripTaskStore) Create(taskID string, payload *model.TripRequest) *TripTask {
	task := &TripTask{
		TaskID: taskID, PlanID: taskID,
		Status: TripTaskProcessing, Stage: "submitted", Progress: 0,
		Message: "任务已提交,等待执行...",
		RequestPayload: payload, UpdatedAt: time.Now(),
	}
	s.mu.Lock()
	s.tasks[taskID] = task
	s.mu.Unlock()
	s.persist(task)
	return task
}

// Get 优先内存,其次磁盘。
func (s *TripTaskStore) Get(taskID string) *TripTask {
	s.mu.RLock()
	task, ok := s.tasks[taskID]
	s.mu.RUnlock()
	if ok {
		return task
	}
	return nil
}

// Update 更新任务状态:持久化并广播事件。
func (s *TripTaskStore) Update(taskID string, mutate func(*TripTask)) {
	s.mu.Lock()
	task, ok := s.tasks[taskID]
	if !ok {
		s.mu.Unlock()
		return
	}
	mutate(task)
	task.UpdatedAt = time.Now()
	event := buildTaskEvent(task, true)
	s.mu.Unlock()

	s.persist(task)
	s.savePlanRecord(task)
	s.broadcast(task, event)
}

// savePlanRecord 规划成功后将完整行程落库到 trip_plans(按 plan_id 幂等覆盖)。
func (s *TripTaskStore) savePlanRecord(task *TripTask) {
	if s.db == nil || task.Result == nil || task.Result.Data == nil {
		return
	}
	plan := task.Result.Data
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return
	}
	graphJSON, err := json.Marshal(task.Result.GraphData)
	if err != nil {
		graphJSON = []byte("null")
	}
	reqJSON := "null"
	if task.RequestPayload != nil {
		if b, err := json.Marshal(task.RequestPayload); err == nil {
			reqJSON = string(b)
		}
	}
	record := model.TripPlanRecord{
		PlanID:             firstNonEmpty(task.PlanID, task.TaskID),
		TaskID:             task.TaskID,
		UserID:             reqUserID(task.RequestPayload),
		City:               plan.City,
		Cities:             strings.Join(plan.Cities, ","),
		StartDate:          plan.StartDate,
		EndDate:            plan.EndDate,
		TravelDays:         len(plan.Days),
		OverallSuggestions: plan.OverallSuggestions,
		PlanJSON:           string(planJSON),
		GraphJSON:          string(graphJSON),
		RequestJSON:        reqJSON,
	}
	if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&record).Error; err != nil {
		fmt.Printf("⚠️  行程计划落库失败 plan_id=%s: %v\n", record.PlanID, err)
		return
	}
	fmt.Printf("💾 行程计划已落库 plan_id=%s city=%s days=%d\n", record.PlanID, record.City, record.TravelDays)
}

// reqUserID 提取请求里的用户标识(前端匿名 ID 或登录用户 ID)。
func reqUserID(req *model.TripRequest) string {
	if req == nil {
		return ""
	}
	return strings.TrimSpace(req.UserID)
}

func (s *TripTaskStore) persist(task *TripTask) {
	payload := tripTaskFile{
		TaskID: task.TaskID, PlanID: task.PlanID, Status: task.Status,
		Stage: task.Stage, Progress: task.Progress, Message: task.Message,
		Result: task.Result, Error: task.Error, RequestPayload: task.RequestPayload,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return
	}
	target := s.filePath(task.TaskID)
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		fmt.Printf("⚠️  持久化任务 %s 失败: %v\n", task.TaskID, err)
		return
	}
	_ = os.Rename(tmp, target)
}

func buildTaskEvent(task *TripTask, includeResult bool) *model.TripTaskEvent {
	event := &model.TripTaskEvent{
		TaskID: task.TaskID, PlanID: task.PlanID, Status: task.Status,
		Stage: task.Stage, Progress: task.Progress, Message: task.Message,
	}
	if task.Error != "" {
		event.Error = task.Error
	}
	if task.Status == TripTaskFailed && task.RequestPayload != nil {
		event.RequestPayload = task.RequestPayload
	}
	if includeResult && task.Result != nil {
		event.Result = task.Result
	}
	return event
}

func (s *TripTaskStore) broadcast(task *TripTask, event *model.TripTaskEvent) {
	s.mu.RLock()
	subs := append([]chan *model.TripTaskEvent{}, task.subscribers...)
	s.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- event:
		default: // 订阅者消费不及时则跳过该事件,避免阻塞任务
		}
	}
}

// Snapshot 返回任务当前事件(供 WebSocket 建连时先行推送)。
func (s *TripTaskStore) Snapshot(taskID string) *model.TripTaskEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return nil
	}
	return buildTaskEvent(task, true)
}

// Subscribe 订阅任务事件,返回事件通道与取消订阅函数。
func (s *TripTaskStore) Subscribe(taskID string) (<-chan *model.TripTaskEvent, func()) {
	ch := make(chan *model.TripTaskEvent, 16)
	s.mu.Lock()
	task, ok := s.tasks[taskID]
	if ok {
		task.subscribers = append(task.subscribers, ch)
	}
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		if t, ok := s.tasks[taskID]; ok {
			kept := t.subscribers[:0]
			for _, sub := range t.subscribers {
				if sub != ch {
					kept = append(kept, sub)
				}
			}
			t.subscribers = kept
		}
		s.mu.Unlock()
	}
	return ch, cancel
}

// IsFinal 任务是否已结束。
func IsTaskFinal(status string) bool {
	return status == TripTaskCompleted || status == TripTaskFailed
}

// History 返回最近完成的历史计划摘要:优先读 MySQL(重启不丢、查询快),无库或库为空时回退内存。
func (s *TripTaskStore) History(limit int) []model.TripHistoryItem {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if s.db != nil {
		var records []model.TripPlanRecord
		if err := s.db.Order("updated_at DESC").Limit(limit).Find(&records).Error; err == nil {
			items := make([]model.TripHistoryItem, 0, len(records))
			for _, r := range records {
				cities := splitCSV(r.Cities)
				displayCity := r.City
				if displayCity == "" && len(cities) > 0 {
					displayCity = cities[0]
				}
				if len(cities) > 1 {
					displayCity = strings.Join(cities, " → ")
				}
				items = append(items, model.TripHistoryItem{
					PlanID:             r.PlanID,
					TaskID:             r.TaskID,
					City:               displayCity,
					Cities:             cities,
					StartDate:          r.StartDate,
					EndDate:            r.EndDate,
					TravelDays:         r.TravelDays,
					UpdatedAt:          r.UpdatedAt.Format("2006-01-02T15:04:05"),
					OverallSuggestions: r.OverallSuggestions,
				})
			}
			if len(items) > 0 {
				return items
			}
		}
	}
	return s.historyFromMemory(limit)
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// historyFromMemory 内存任务兜底逻辑(兼容未落库的旧磁盘任务)。
func (s *TripTaskStore) historyFromMemory(limit int) []model.TripHistoryItem {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	s.mu.RLock()
	tasks := make([]*TripTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	s.mu.RUnlock()
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt) })

	items := make([]model.TripHistoryItem, 0, limit)
	for _, task := range tasks {
		if task.Status != TripTaskCompleted || task.Result == nil || task.Result.Data == nil {
			continue
		}
		plan := task.Result.Data
		req := task.RequestPayload
		city := plan.City
		if city == "" && req != nil {
			city = req.City
		}
		cities := plan.Cities
		startDate := plan.StartDate
		endDate := plan.EndDate
		if req != nil {
			if startDate == "" {
				startDate = req.StartDate
			}
			if endDate == "" {
				endDate = req.EndDate
			}
		}
		travelDays := len(plan.Days)
		if req != nil && req.TravelDays > 0 {
			travelDays = req.TravelDays
		}
		if city == "" && len(cities) == 0 {
			continue
		}
		displayCity := city
		if len(cities) > 1 {
			displayCity = strings.Join(cities, " → ")
		}
		items = append(items, model.TripHistoryItem{
			PlanID:             firstNonEmpty(task.PlanID, task.TaskID),
			TaskID:             task.TaskID,
			City:               displayCity,
			Cities:             cities,
			StartDate:          startDate,
			EndDate:            endDate,
			TravelDays:         travelDays,
			UpdatedAt:          task.UpdatedAt.Format("2006-01-02T15:04:05"),
			OverallSuggestions: plan.OverallSuggestions,
		})
		if len(items) >= limit {
			break
		}
	}
	return items
}

// GetPlanRecord 按 plan_id 读取落库的完整行程(回看历史计划)。
func (s *TripTaskStore) GetPlanRecord(planID string) (*model.TripPlanRecord, error) {
	if s.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var record model.TripPlanRecord
	if err := s.db.Where("plan_id = ?", planID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// DeletePlan 删除一条落库的历史计划。
func (s *TripTaskStore) DeletePlan(planID string) error {
	if s.db == nil {
		return gorm.ErrInvalidDB
	}
	return s.db.Where("plan_id = ?", planID).Delete(&model.TripPlanRecord{}).Error
}