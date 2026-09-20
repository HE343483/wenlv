package service

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wenlv-backend/logger"
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

	// 可选:行程完成后后台把景点图片上传 OSS 并把直链写回 plan_json
	imageEnricher *TripImageEnricher
	enriching     sync.Map // plan_id -> true,防同一计划重复触发上传
}

// AttachDB 接入 MySQL,行程完成后自动落库,历史接口优先读库。
func (s *TripTaskStore) AttachDB(db *gorm.DB) { s.db = db }

// AttachImageEnricher 接入 OSS 图片上传器(nil 跳过)。
func (s *TripTaskStore) AttachImageEnricher(e *TripImageEnricher) { s.imageEnricher = e }

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
		logger.Infof("已加载 %d 个持久化旅行任务", loaded)
	}
}

// Create 创建任务。
func (s *TripTaskStore) Create(taskID string, payload *model.TripRequest) *TripTask {
	task := &TripTask{
		TaskID: taskID, PlanID: taskID,
		Status: TripTaskProcessing, Stage: "submitted", Progress: 0,
		Message:        "任务已提交,等待执行...",
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

// savePlanRecord 任务终态落库到 trip_plans:成功存完整行程,失败存错误信息(按 plan_id 幂等覆盖)。
func (s *TripTaskStore) savePlanRecord(task *TripTask) {
	if s.db == nil {
		return
	}
	// 仅终态写入,处理中的任务不落库
	if task.Status != TripTaskCompleted && task.Status != TripTaskFailed {
		return
	}
	if task.Status == TripTaskCompleted && (task.Result == nil || task.Result.Data == nil) {
		return
	}
	// 城市与日期:优先取行程结果,失败时回退到原始请求
	city, citiesCSV, startDate, endDate := "", "", "", ""
	travelDays := 0
	overallSuggestions := ""
	planLanguage := ""
	planJSON, graphJSON := []byte("null"), []byte("null")
	if task.Status == TripTaskCompleted {
		plan := task.Result.Data
		var err error
		if planJSON, err = json.Marshal(plan); err != nil {
			planJSON = []byte("null")
		}
		if graphJSON, err = json.Marshal(task.Result.GraphData); err != nil {
			graphJSON = []byte("null")
		}
		city = plan.City
		citiesCSV = strings.Join(plan.Cities, ",")
		startDate, endDate = plan.StartDate, plan.EndDate
		travelDays = len(plan.Days)
		overallSuggestions = plan.OverallSuggestions
		planLanguage = plan.Language
	}
	if req := task.RequestPayload; req != nil {
		if city == "" && len(req.Cities) > 0 {
			names := make([]string, 0, len(req.Cities))
			for _, cs := range req.Cities {
				names = append(names, cs.City)
			}
			city = names[0]
			if citiesCSV == "" {
				citiesCSV = strings.Join(names, ",")
			}
		}
		if startDate == "" {
			startDate = req.StartDate
		}
		if endDate == "" {
			endDate = req.EndDate
		}
		if travelDays == 0 {
			travelDays = req.TravelDays
		}
	}
	reqJSON := "null"
	if task.RequestPayload != nil {
		if b, err := json.Marshal(task.RequestPayload); err == nil {
			reqJSON = string(b)
		}
	}
	status := task.Status
	if status == "" {
		status = TripTaskCompleted
	}
	// 计划语言:优先取行程结果里记录的生成语言,失败任务回退原始请求语言
	if planLanguage == "" && task.RequestPayload != nil {
		planLanguage = task.RequestPayload.Language
	}
	planLanguage = model.NormalizeLang(planLanguage)
	record := model.TripPlanRecord{
		PlanID:             firstNonEmpty(task.PlanID, task.TaskID),
		TaskID:             task.TaskID,
		UserID:             reqUserID(task.RequestPayload),
		City:               city,
		Cities:             citiesCSV,
		StartDate:          startDate,
		EndDate:            endDate,
		TravelDays:         travelDays,
		OverallSuggestions: overallSuggestions,
		Language:           planLanguage,
		PlanJSON:           string(planJSON),
		GraphJSON:          string(graphJSON),
		RequestJSON:        reqJSON,
		Status:             status,
		ErrorMessage:       task.Error,
	}
	if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&record).Error; err != nil {
		logger.Errorf("行程计划落库失败 plan_id=%s: %v", record.PlanID, err)
		return
	}
	if status == TripTaskFailed {
		logger.Warnf("失败任务已记录 plan_id=%s city=%s err=%s", record.PlanID, record.City, task.Error)
	} else {
		logger.Infof("行程计划已落库 plan_id=%s city=%s days=%d", record.PlanID, record.City, record.TravelDays)
		s.maybeEnrichImages(task)
	}
}

// reqUserID 提取请求里的用户标识(前端匿名 ID 或登录用户 ID)。
func reqUserID(req *model.TripRequest) string {
	if req == nil {
		return ""
	}
	return strings.TrimSpace(req.UserID)
}

// maybeEnrichImages completed 行程首次落库后,后台把景点图片上传 OSS 并把直链二次落库。
// 不阻塞任务返回:用户当次看到的实时页面仍走代理,历史详情随后改用 OSS 直链。
func (s *TripTaskStore) maybeEnrichImages(task *TripTask) {
	if s.imageEnricher == nil || task.Result == nil || task.Result.Data == nil {
		return
	}
	planID := firstNonEmpty(task.PlanID, task.TaskID)
	if _, loaded := s.enriching.LoadOrStore(planID, true); loaded {
		return
	}
	go func() {
		defer s.enriching.Delete(planID)
		// 深拷贝后再抓图上传,避免与轮询/WS 广播并发读写同一结构
		raw, err := json.Marshal(task.Result.Data)
		if err != nil {
			return
		}
		copied := &model.TripPlan{}
		if err := json.Unmarshal(raw, copied); err != nil {
			return
		}
		ok, fail := s.imageEnricher.Enrich(context.Background(), copied)
		if ok == 0 {
			if fail > 0 {
				logger.Warnf("[行程图片OSS] plan_id=%s 图片上传全部失败,历史详情继续走代理", planID)
			}
			return
		}
		// OSS 链接写回内存任务并二次落库(savePlanRecord 按 plan_id 幂等覆盖)
		s.mu.Lock()
		if t, exists := s.tasks[task.TaskID]; exists && t.Result != nil && t.Result.Data != nil {
			applyPlanImages(t.Result.Data, copied)
		}
		s.mu.Unlock()
		s.savePlanRecord(task)
		logger.Infof("[行程图片OSS] plan_id=%s 成功 %d 张,失败 %d 张,历史详情已改用 OSS 直链", planID, ok, fail)
	}()
}

// applyPlanImages 把上传补齐的 image_url 按景点名同步回原 plan。
func applyPlanImages(dst, src *model.TripPlan) {
	urlByName := make(map[string]string)
	for i := range src.Days {
		for j := range src.Days[i].Attractions {
			a := &src.Days[i].Attractions[j]
			if a.Name != "" && a.ImageURL != "" {
				urlByName[a.Name] = a.ImageURL
			}
		}
	}
	for i := range dst.Days {
		for j := range dst.Days[i].Attractions {
			if url, ok := urlByName[dst.Days[i].Attractions[j].Name]; ok {
				dst.Days[i].Attractions[j].ImageURL = url
			}
		}
	}
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
		logger.Warnf("持久化任务 %s 失败: %v", task.TaskID, err)
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
				status := r.Status
				if status == "" {
					status = TripTaskCompleted
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
					Language:           r.Language,
					Status:             status,
					ErrorMessage:       r.ErrorMessage,
				})
			}
			if len(items) > 0 {
				return items
			}
		}
	}
	return s.historyFromMemory(limit)
}

// memoryPlanLanguage 内存历史回退时提取计划语言:优先行程结果,再回退原始请求。
func memoryPlanLanguage(task *TripTask, _ string) string {
	if task.Result != nil && task.Result.Data != nil && strings.TrimSpace(task.Result.Data.Language) != "" {
		return model.NormalizeLang(task.Result.Data.Language)
	}
	if task.RequestPayload != nil {
		return model.NormalizeLang(task.RequestPayload.Language)
	}
	return ""
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
		// 失败任务也进入历史(带错误信息);处理中的任务跳过
		if task.Status != TripTaskCompleted && task.Status != TripTaskFailed {
			continue
		}
		isFailed := task.Status == TripTaskFailed
		if !isFailed && (task.Result == nil || task.Result.Data == nil) {
			continue
		}
		req := task.RequestPayload
		city, cities := "", []string(nil)
		startDate, endDate, travelDays := "", "", 0
		overallSuggestions := ""
		if !isFailed {
			plan := task.Result.Data
			city = plan.City
			cities = plan.Cities
			startDate, endDate = plan.StartDate, plan.EndDate
			travelDays = len(plan.Days)
			overallSuggestions = plan.OverallSuggestions
		}
		if city == "" && req != nil {
			city = req.City
		}
		if len(cities) == 0 && req != nil && len(req.Cities) > 0 {
			cities = make([]string, 0, len(req.Cities))
			for _, cs := range req.Cities {
				cities = append(cities, cs.City)
			}
		}
		if req != nil {
			if startDate == "" {
				startDate = req.StartDate
			}
			if endDate == "" {
				endDate = req.EndDate
			}
			if travelDays == 0 {
				travelDays = req.TravelDays
			}
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
			OverallSuggestions: overallSuggestions,
			Language:           memoryPlanLanguage(task, overallSuggestions),
			Status:             task.Status,
			ErrorMessage:       task.Error,
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

// SaveStoryCard 将故事卡片文案按语言持久化到 trip_plans.story_card_json,
// 避免每次打开弹窗都重复调用 LLM;按 plan_id+language 覆盖更新。
func (s *TripTaskStore) SaveStoryCard(planID string, lang string, content *TripStoryCardContent) error {
	if s.db == nil {
		return gorm.ErrInvalidDB
	}
	var record model.TripPlanRecord
	if err := s.db.Where("plan_id = ?", planID).First(&record).Error; err != nil {
		return err
	}
	cards := map[string]*TripStoryCardContent{}
	if strings.TrimSpace(record.StoryCardJSON) != "" {
		if err := json.Unmarshal([]byte(record.StoryCardJSON), &cards); err != nil {
			cards = map[string]*TripStoryCardContent{}
		}
	}
	cards[lang] = content
	buf, err := json.Marshal(cards)
	if err != nil {
		return err
	}
	return s.db.Model(&model.TripPlanRecord{}).Where("plan_id = ?", planID).
		Update("story_card_json", string(buf)).Error
}

// GetStoryCard 读取已缓存的故事卡片文案(未缓存返回 nil,不视为错误)。
func (s *TripTaskStore) GetStoryCard(planID string, lang string) (*TripStoryCardContent, error) {
	record, err := s.GetPlanRecord(planID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.StoryCardJSON) == "" {
		return nil, nil
	}
	var cards map[string]*TripStoryCardContent
	if err := json.Unmarshal([]byte(record.StoryCardJSON), &cards); err != nil {
		return nil, nil
	}
	return cards[lang], nil
}

// DeletePlan 删除一条落库的历史计划。
func (s *TripTaskStore) DeletePlan(planID string) error {
	if s.db == nil {
		return gorm.ErrInvalidDB
	}
	return s.db.Where("plan_id = ?", planID).Delete(&model.TripPlanRecord{}).Error
}
