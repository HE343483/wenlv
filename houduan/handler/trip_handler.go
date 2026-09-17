package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"wenlv-backend/model"
	"wenlv-backend/service"
)

// TripHandler AI 行程规划接口(移植自 TripStar,响应结构保持与前端一致)。
type TripHandler struct {
	planner *service.TripPlanner
	chat    *service.TripChatService
	tasks   *service.TripTaskStore
	settings *service.TripSettings
}

// NewTripHandler 构造行程规划处理器。
func NewTripHandler(planner *service.TripPlanner, chat *service.TripChatService,
	tasks *service.TripTaskStore, settings *service.TripSettings) *TripHandler {
	return &TripHandler{planner: planner, chat: chat, tasks: tasks, settings: settings}
}

var tripUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	// 与主站 CORS 策略保持一致,允许跨域前端订阅任务进度
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Plan 提交旅行规划任务(立即返回 task_id)。
func (h *TripHandler) Plan(c *gin.Context) {
	var req model.TripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误:城市/日期/天数等必填字段不完整"})
		return
	}
	req.Normalize()
	if len(req.Cities) == 0 || req.StartDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误:至少需要一个城市与开始日期"})
		return
	}

	taskID := service.NewTaskID()
	h.tasks.Create(taskID, &req)
	h.tasks.Update(taskID, func(t *service.TripTask) {
		t.Status = service.TripTaskProcessing
		t.Stage = "submitted"
		t.Progress = 5
		t.Message = "任务已提交,正在初始化流程..."
	})

	cityDisplay := req.City
	if len(req.Cities) > 1 {
		names := make([]string, 0, len(req.Cities))
		for _, cs := range req.Cities {
			names = append(names, cs.City)
		}
		cityDisplay = joinArrow(names)
	}
	fmt.Printf("\n📥 收到旅行规划请求 (task_id=%s): %s\n", taskID, cityDisplay)

	// 后台执行规划,通过 WebSocket / 轮询推送进度。
	// 这里必须脱离请求上下文:HTTP 响应返回后请求上下文会被取消,而规划任务需要继续执行。
	planCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	go func() {
		defer cancel()
		h.planner.RunPlanning(planCtx, taskID, &req)
	}()

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"plan_id": taskID,
		"status":  "processing",
		"ws_url":  "/api/trip/ws/" + taskID,
		"message": "任务已提交,可通过 WebSocket /api/trip/ws/" + taskID + " 实时订阅状态",
	})
}

// Status 查询任务状态(兼容轮询客户端)。
func (h *TripHandler) Status(c *gin.Context) {
	taskID := c.Param("taskId")
	event := h.tasks.Snapshot(taskID)
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "任务不存在"})
		return
	}
	switch event.Status {
	case service.TripTaskCompleted:
		c.JSON(http.StatusOK, gin.H{
			"task_id": taskID,
			"plan_id": event.PlanID,
			"status":  event.Status,
			"result":  event.Result,
		})
	case service.TripTaskFailed:
		c.JSON(http.StatusOK, gin.H{
			"task_id":         taskID,
			"plan_id":         event.PlanID,
			"status":          event.Status,
			"error":           event.Error,
			"request_payload": event.RequestPayload,
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"task_id":       taskID,
			"plan_id":       event.PlanID,
			"status":        event.Status,
			"stage":         event.Stage,
			"progress":      event.Progress,
			"progress_text": event.Message,
		})
	}
}

// History 最近历史计划。
func (h *TripHandler) History(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 10)
	c.JSON(http.StatusOK, gin.H{"items": h.tasks.History(limit)})
}

// PlanDetail 回看历史计划:从 trip_plans 表返回完整行程与知识图谱。
func (h *TripHandler) PlanDetail(c *gin.Context) {
	planID := c.Param("planId")
	record, err := h.tasks.GetPlanRecord(planID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "计划不存在或已删除"})
		return
	}
	result := &model.TripPlanResponse{Success: true, PlanID: record.PlanID}
	if record.PlanJSON != "" {
		var plan model.TripPlan
		if err := json.Unmarshal([]byte(record.PlanJSON), &plan); err == nil {
			result.Data = &plan
		}
	}
	if record.GraphJSON != "" && record.GraphJSON != "null" {
		var graph model.KnowledgeGraphData
		if err := json.Unmarshal([]byte(record.GraphJSON), &graph); err == nil {
			result.GraphData = &graph
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"task_id": record.TaskID,
		"plan_id": record.PlanID,
		"status":  "completed",
		"result":  result,
	})
}

// DeleteHistory 删除一条落库的历史计划。
func (h *TripHandler) DeleteHistory(c *gin.Context) {
	planID := c.Param("planId")
	if err := h.tasks.DeletePlan(planID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "删除失败:" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已删除"})
}

// Health 健康检查。
func (h *TripHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":         "healthy",
		"service":        "trip-planner",
		"map_provider":   service.CurrentMapProvider(h.settings),
		"llm_configured": h.settings.Snapshot().OpenAIAPIKey != "",
	})
}

// WS WebSocket 订阅任务状态。
func (h *TripHandler) WS(c *gin.Context) {
	taskID := c.Param("taskId")
	conn, err := tripUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 先订阅再取快照(与原版一致):若任务在两步之间结束,最终事件仍会进入订阅通道,避免前端永久等待
	events, cancel := h.tasks.Subscribe(taskID)
	defer cancel()

	snapshot := h.tasks.Snapshot(taskID)
	if snapshot == nil {
		_ = conn.WriteJSON(gin.H{
			"task_id": taskID, "plan_id": taskID, "status": "failed",
			"stage": "failed", "progress": 100,
			"message": "任务不存在", "error": "任务不存在",
		})
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(1008, "任务不存在"), time.Now().Add(time.Second))
		return
	}
	// 推送快照,保证后连接的前端也能同步当前状态
	if err := conn.WriteJSON(snapshot); err != nil {
		return
	}
	if service.IsTaskFinal(snapshot.Status) {
		return
	}

	// 客户端关闭连接时退出阻塞
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := conn.WriteJSON(event); err != nil {
				return
			}
			if service.IsTaskFinal(event.Status) {
				return
			}
		}
	}
}

// Ask 行程智能问答。
func (h *TripHandler) Ask(c *gin.Context) {
	var req model.TripChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "参数错误:message 与 trip_plan 必填"})
		return
	}
	reply, err := h.chat.ChatWithTripContext(c.Request.Context(), req.Message, req.TripPlan, req.History)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "AI问答服务异常: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.TripChatResponse{Success: true, Reply: reply})
}

func joinArrow(items []string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += " → "
		}
		out += item
	}
	return out
}