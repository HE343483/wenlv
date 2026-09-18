package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"wenlv-backend/model"
	"wenlv-backend/repository"
)

// ============ 用户偏好记忆模块(带权重衰减、合并与 TOP-K 召回) ============

const (
	// TripExplicitInitWeight 用户手动添加偏好时的初始权重(高于 LLM 自动提取)
	TripExplicitInitWeight = 8.5
	tripDaySeconds         = 86400.0
)

// TripMemoryService 用户偏好记忆服务。
type TripMemoryService struct {
	settings *TripSettings
	repo     *repository.UserMemoryRepo
	llm      *TripLLM
}

// NewTripMemoryService 构造记忆服务。
func NewTripMemoryService(settings *TripSettings, repo *repository.UserMemoryRepo, llm *TripLLM) *TripMemoryService {
	return &TripMemoryService{settings: settings, repo: repo, llm: llm}
}

// Enabled 记忆模块是否开启。
func (s *TripMemoryService) Enabled() bool {
	return s.settings.MemoryEnabled() && s.repo != nil
}

// AddMemory 新增记忆;相同文本合并权重;低于准入门槛直接拒绝。
func (s *TripMemoryService) AddMemory(ctx context.Context, userID, content, source string, initWeight float64) error {
	if !s.Enabled() || strings.TrimSpace(userID) == "" || strings.TrimSpace(content) == "" {
		return nil
	}
	if initWeight < s.settings.MemoryParams().MinInitWeight {
		return nil
	}
	items, err := s.repo.ListAll(userID)
	if err != nil {
		return err
	}
	trimmed := strings.TrimSpace(content)
	for i := range items {
		if strings.TrimSpace(items[i].Content) == trimmed {
			newWeight := items[i].Weight + initWeight*0.4
			if newWeight > 10.0 {
				newWeight = 10.0
			}
			items[i].Weight = newWeight
			items[i].LastAccessTime = float64(time.Now().Unix())
			return s.repo.Save(&items[i])
		}
	}
	now := float64(time.Now().Unix())
	return s.repo.Save(&model.UserMemory{
		MemoryID:       fmt.Sprintf("%d-%s", time.Now().UnixNano(), userID),
		UserID:         userID,
		Content:        trimmed,
		Source:         source,
		Weight:         initWeight,
		CreateTime:     now,
		LastAccessTime: now,
	})
}

// RecallUserMemory 召回记忆:惰性遗忘衰减 + 权重降序 + TOP-K 截断。
func (s *TripMemoryService) RecallUserMemory(ctx context.Context, userID string) ([]model.UserMemory, error) {
	if !s.Enabled() || strings.TrimSpace(userID) == "" {
		return nil, nil
	}
	items, err := s.repo.ListAll(userID)
	if err != nil {
		return nil, err
	}
	params := s.settings.MemoryParams()
	now := float64(time.Now().Unix())
	valid := make([]model.UserMemory, 0, len(items))
	for _, item := range items {
		deltaDays := (now - item.LastAccessTime) / tripDaySeconds
		item.Weight = item.Weight * pow(params.DecayFactor, deltaDays)
		item.LastAccessTime = now
		if item.Weight >= params.WeightThreshold {
			item.Weight = roundWeight(item.Weight)
			valid = append(valid, item)
			_ = s.repo.Save(&item)
		} else {
			_, _ = s.repo.Delete(userID, item.MemoryID)
		}
	}
	sort.SliceStable(valid, func(i, j int) bool { return valid[i].Weight > valid[j].Weight })
	if len(valid) > params.MaxRecall {
		valid = valid[:params.MaxRecall]
	}
	return valid, nil
}

// DeleteMemory 删除单条记忆。
func (s *TripMemoryService) DeleteMemory(ctx context.Context, userID, memoryID string) (bool, error) {
	if !s.Enabled() {
		return false, errors.New("记忆模块未开启")
	}
	return s.repo.Delete(userID, memoryID)
}

// ClearMemory 清空用户全部记忆。
func (s *TripMemoryService) ClearMemory(ctx context.Context, userID string) error {
	if !s.Enabled() {
		return errors.New("记忆模块未开启")
	}
	return s.repo.Clear(userID)
}

// BuildPromptSnippet 组装注入规划 Prompt 的偏好文本(TOP-K + 单条长度截断)。
func (s *TripMemoryService) BuildPromptSnippet(ctx context.Context, userID string) string {
	memories, err := s.RecallUserMemory(ctx, userID)
	if err != nil || len(memories) == 0 {
		return ""
	}
	lines := make([]string, 0, len(memories))
	maxLen := s.settings.MemoryParams().MaxSingleContent
	for _, m := range memories {
		content := strings.TrimSpace(m.Content)
		runes := []rune(content)
		if len(runes) > maxLen {
			content = string(runes[:maxLen])
		}
		lines = append(lines, "- "+content)
	}
	return "【用户历史旅行偏好】\n" + strings.Join(lines, "\n")
}

// ============ 偏好提取(行程生成后调用 LLM 打分入库) ============

const tripExtractPrompt = `你需要从用户出行需求和最终生成的旅行行程中,提炼用户稳定旅行偏好。
规则:
1. 只提取长期稳定偏好,临时目的地、单次短期安排不要提取
2. 每条偏好附带重要性分数(0~10),分数越高代表用户长期习惯越强
3. 输出JSON数组,结构示例:
[
    {"content": "偏爱小众自然景点", "score": 7.2},
    {"content": "住宿偏好民宿", "score": 6.5}
]
4. 没有稳定偏好返回空数组[]
只返回纯JSON,禁止额外解释文字。

用户原始需求:
%s
生成行程内容:
%s
`

// SavePreferencesAfterTrip 行程生成成功后异步提取偏好并写入记忆库(失败不影响主流程)。
func (s *TripMemoryService) SavePreferencesAfterTrip(ctx context.Context, userID string, req *model.TripRequest, plan *model.TripPlan) {
	if !s.Enabled() || strings.TrimSpace(userID) == "" {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("⚠️  偏好记忆提取异常: %v\n", r)
		}
	}()

	userQuery := fmt.Sprintf("城市: %s; 偏好: %s; 额外要求: %s",
		req.City, strings.Join(req.Preferences, ", "), req.FreeTextInput)
	if len(req.Preferences) == 0 {
		userQuery = fmt.Sprintf("城市: %s; 偏好: 无; 额外要求: %s", req.City, firstNonEmpty(req.FreeTextInput, "无"))
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return
	}
	content := string(planJSON)
	runes := []rune(content)
	if len(runes) > 6000 {
		content = string(runes[:6000])
	}

	prompt := fmt.Sprintf(tripExtractPrompt, userQuery, content)
	reply, err := s.llm.Chat(ctx, NewLLMMessages("你是一个旅行偏好分析助手,只输出JSON。", prompt), 0.2, 800)
	if err != nil {
		fmt.Printf("⚠️  偏好记忆提取失败: %v\n", err)
		return
	}
	jsonText := extractJSONArray(reply)
	if jsonText == "" {
		return
	}
	var parsed []struct {
		Content string  `json:"content"`
		Score   float64 `json:"score"`
	}
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		fmt.Printf("⚠️  偏好记忆 JSON 解析失败: %v\n", err)
		return
	}
	added := 0
	minWeight := s.settings.MemoryParams().MinInitWeight
	for _, item := range parsed {
		text := strings.TrimSpace(item.Content)
		if text == "" || item.Score < minWeight {
			continue
		}
		if err := s.AddMemory(ctx, userID, text, "implicit", item.Score); err == nil {
			added++
		}
	}
	if added > 0 {
		fmt.Printf("🧠 用户 %s 偏好记忆已更新,新增/合并 %d 条\n", truncateText(userID, 8), added)
	}
}

func pow(base, exp float64) float64 {
	return math.Pow(base, exp)
}

func roundWeight(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}