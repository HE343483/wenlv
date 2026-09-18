package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"wenlv-backend/model"
)

// ============ 大模型 JSON 输出容错解析 ============
// 移植自原项目 trip_planner_agent 的多层容错策略:
// 基础清理 → 修复未转义引号 → 截断修复 → 正则暴力提取 → 最后交由 LLM 修补。

var (
	codeFenceStartRe = regexp.MustCompile("(?s)^```(?:json)?\\s*")
	codeFenceEndRe   = regexp.MustCompile("```\\s*$")
	jsLineCommentRe  = regexp.MustCompile(`//[^\n]*`)
	jsBlockCommentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	controlCharRe    = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f]`)
	trailingCommaRe  = regexp.MustCompile(`,\s*([\]}])`)
	arithExprRe      = regexp.MustCompile(`:\s*(\d+(?:\s*[+\-*/]\s*\d+)+(?:\s*=\s*\d+)?)`)
	jsonObjectRe     = regexp.MustCompile(`(?s)\{.*\}`)
	trailingJunkRe   = regexp.MustCompile(`,\s*$`)
)

var cjkPunctReplacer = strings.NewReplacer(
	"\u201c", "'", "\u201d", "'", "\u2018", "'", "\u2019", "'",
	"\uff1a", ":", "\uff0c", ",",
)

// sanitizeJSONString 清理大模型输出中常见的 JSON 格式污染。
func sanitizeJSONString(s string) string {
	s = strings.TrimSpace(s)
	s = codeFenceStartRe.ReplaceAllString(s, "")
	s = codeFenceEndRe.ReplaceAllString(s, "")
	s = jsLineCommentRe.ReplaceAllString(s, "")
	s = jsBlockCommentRe.ReplaceAllString(s, "")
	s = controlCharRe.ReplaceAllString(s, "")
	s = trailingCommaRe.ReplaceAllString(s, "$1")
	s = cjkPunctReplacer.Replace(s)
	s = arithExprRe.ReplaceAllStringFunc(s, fixArithmeticExpr)
	return s
}

// fixArithmeticExpr 将 budget 等数值字段中的算术表达式替换为最终结果。
// 例如 "total_attractions": 30+54+120+120=324 → "total_attractions": 324
func fixArithmeticExpr(match string) string {
	sub := arithExprRe.FindStringSubmatch(match)
	if len(sub) < 2 {
		return match
	}
	expr := strings.TrimSpace(sub[1])
	if idx := strings.LastIndex(expr, "="); idx >= 0 {
		return strings.Replace(match, sub[1], strings.TrimSpace(expr[idx+1:]), 1)
	}
	if value, ok := evaluateArithmetic(expr); ok {
		return strings.Replace(match, sub[1], strconv.FormatFloat(value, 'f', -1, 64), 1)
	}
	return match
}

// evaluateArithmetic 计算形如 30+54*2-10 的简单算术表达式(支持 + - * / 优先级)。
func evaluateArithmetic(expr string) (float64, bool) {
	tokens := make([]string, 0, 16)
	current := strings.Builder{}
	for _, r := range expr {
		switch r {
		case '+', '-', '*', '/':
			if current.Len() == 0 {
				return 0, false
			}
			tokens = append(tokens, current.String(), string(r))
			current.Reset()
		case ' ', '\t':
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() == 0 {
		return 0, false
	}
	tokens = append(tokens, current.String())

	// 第一遍:处理乘除
	stack := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if (token == "*" || token == "/") && len(stack) > 0 && i+1 < len(tokens) {
			left, err1 := strconv.ParseFloat(stack[len(stack)-1], 64)
			right, err2 := strconv.ParseFloat(tokens[i+1], 64)
			if err1 != nil || err2 != nil {
				return 0, false
			}
			var value float64
			if token == "*" {
				value = left * right
			} else {
				if right == 0 {
					return 0, false
				}
				value = left / right
			}
			stack[len(stack)-1] = strconv.FormatFloat(value, 'f', -1, 64)
			i++
			continue
		}
		stack = append(stack, token)
	}

	// 第二遍:处理加减
	if len(stack) == 0 {
		return 0, false
	}
	total, err := strconv.ParseFloat(stack[0], 64)
	if err != nil {
		return 0, false
	}
	for i := 1; i+1 < len(stack); i += 2 {
		op := stack[i]
		value, err := strconv.ParseFloat(stack[i+1], 64)
		if err != nil {
			return 0, false
		}
		switch op {
		case "+":
			total += value
		case "-":
			total -= value
		default:
			return 0, false
		}
	}
	return total, true
}

// fixUnescapedQuotes 修复 JSON 字符串值内部未转义的双引号。
// 例如 "description": "这是"好的"景点" → "description": "这是'好的'景点"
func fixUnescapedQuotes(s string) string {
	var result strings.Builder
	result.Grow(len(s) + 16)
	inString := false
	escapeNext := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escapeNext {
			result.WriteByte(ch)
			escapeNext = false
			continue
		}
		if ch == '\\' && inString {
			escapeNext = true
			result.WriteByte(ch)
			continue
		}
		if ch == '"' {
			if !inString {
				inString = true
				result.WriteByte(ch)
				continue
			}
			rest := strings.TrimLeft(s[i+1:], " \t\r\n")
			if rest == "" {
				inString = false
				result.WriteByte(ch)
				continue
			}
			switch rest[0] {
			case ',', '}', ']', ':':
				inString = false
				result.WriteByte(ch)
			default:
				result.WriteByte('\'')
			}
			continue
		}
		result.WriteByte(ch)
	}
	return result.String()
}

// repairTruncatedJSON 修复被 max_tokens 截断的不完整 JSON。
func repairTruncatedJSON(s string) string {
	s = strings.TrimRight(s, " \t\r\n")
	if s == "" {
		return s
	}

	// Step 1: 关闭未终止的字符串
	inStr := false
	escape := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escape {
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inStr = !inStr
		}
	}
	if inStr {
		s = strings.TrimRight(s, "\\")
		s += "\""
	}

	// Step 2: 反复去除尾部非法字符
	validTail := "}]\"" + "0123456789" + "els"
	for i := 0; i < 10; i++ {
		stripped := strings.TrimRight(s, " \t\r\n")
		if stripped == "" {
			break
		}
		if strings.ContainsRune(validTail, rune(stripped[len(stripped)-1])) {
			break
		}
		s = stripped[:len(stripped)-1]
	}
	s = trailingJunkRe.ReplaceAllString(s, "")

	// Step 3: 依据未闭合的括号栈补齐
	stack := make([]byte, 0, 8)
	inStr2 := false
	esc2 := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if esc2 {
			esc2 = false
			continue
		}
		if ch == '\\' && inStr2 {
			esc2 = true
			continue
		}
		if ch == '"' {
			inStr2 = !inStr2
			continue
		}
		if inStr2 {
			continue
		}
		switch ch {
		case '{', '[':
			stack = append(stack, ch)
		case '}':
			if len(stack) > 0 && stack[len(stack)-1] == '{' {
				stack = stack[:len(stack)-1]
			}
		case ']':
			if len(stack) > 0 && stack[len(stack)-1] == '[' {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(stack) > 0 {
		var closing strings.Builder
		closing.WriteString("\n")
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i] == '[' {
				closing.WriteByte(']')
			} else {
				closing.WriteByte('}')
			}
		}
		s += closing.String()
	}
	return s
}

// llmRepairJSON 使用 LLM 修复无法本地修复的 JSON(最后手段)。
func llmRepairJSON(ctx context.Context, llm *TripLLM, broken string) string {
	tail := broken
	if runes := []rune(broken); len(runes) > 2000 {
		tail = string(runes[len(runes)-2000:])
	}
	head := broken
	if runes := []rune(broken); len(runes) > 500 {
		head = string(runes[:500])
	}
	prompt := fmt.Sprintf(`以下是一段被截断的旅行计划 JSON,请你补全它使其成为合法的 JSON。
只输出修复后的完整 JSON,不要输出任何解释文字。

开头部分:
%s

...(中间省略)...

尾部被截断部分:
%s
`, head, tail)

	reply, err := llm.Chat(ctx, UserMessage(prompt), 0.0, 8000)
	if err != nil {
		fmt.Printf("⚠️  LLM 修复 JSON 失败: %v\n", err)
		return broken
	}
	return extractJSONObjectText(reply, broken)
}

func extractJSONObjectText(reply, fallback string) string {
	if idx := strings.Index(reply, "```json"); idx >= 0 {
		start := idx + len("```json")
		if end := strings.Index(reply[start:], "```"); end > 0 {
			return strings.TrimSpace(reply[start : start+end])
		}
	}
	if idx := strings.Index(reply, "```"); idx >= 0 {
		start := idx + 3
		if end := strings.Index(reply[start:], "```"); end > 0 {
			return strings.TrimSpace(reply[start : start+end])
		}
	}
	if m := jsonObjectRe.FindString(reply); m != "" {
		return m
	}
	if strings.TrimSpace(reply) == "" {
		return fallback
	}
	return reply
}

// ParseTripPlan 解析大模型返回的行程 JSON,带多层容错。
func ParseTripPlan(ctx context.Context, llm *TripLLM, response string, req *model.TripRequest) (*model.TripPlan, error) {
	jsonStr := extractJSONFromResponse(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("行程 JSON 解析失败: 响应中未找到JSON数据")
	}

	jsonStr = sanitizeJSONString(jsonStr)

	type attempt struct {
		name      string
		candidate string
	}
	attempts := []attempt{{"基础清理", jsonStr}}

	fixedQuotes := fixUnescapedQuotes(jsonStr)
	attempts = append(attempts, attempt{"修复未转义引号", fixedQuotes})

	repaired := repairTruncatedJSON(jsonStr)
	if repaired != jsonStr {
		attempts = append(attempts, attempt{"截断修复", repaired})
		if combined := fixUnescapedQuotes(repaired); combined != repaired {
			attempts = append(attempts, attempt{"截断+引号修复", combined})
		}
	}

	if m := jsonObjectRe.FindString(jsonStr); m != "" {
		brutal := fixUnescapedQuotes(sanitizeJSONString(m))
		attempts = append(attempts, attempt{"正则提取", brutal})
		if brutalRepaired := repairTruncatedJSON(brutal); brutalRepaired != brutal {
			attempts = append(attempts, attempt{"正则+截断修复", brutalRepaired})
		}
	}

	var lastErr error
	for _, a := range attempts {
		var plan model.TripPlan
		if err := json.Unmarshal([]byte(a.candidate), &plan); err == nil {
			if a.name != "基础清理" {
				fmt.Printf("✅ 行程 JSON 通过「%s」成功解析\n", a.name)
			}
			return &plan, nil
		} else {
			lastErr = err
			if a.name == "基础清理" {
				fmt.Printf("⚠️  首次行程 JSON 解析失败: %v\n", err)
			} else {
				fmt.Printf("⚠️  「%s」仍失败: %v\n", a.name, err)
			}
		}
	}

	// 最终手段:LLM 修复
	fmt.Println("🔧 所有本地修复均失败,尝试使用 LLM 修复行程 JSON...")
	llmFixed := sanitizeJSONString(llmRepairJSON(ctx, llm, jsonStr))
	var plan model.TripPlan
	if err := json.Unmarshal([]byte(llmFixed), &plan); err == nil {
		fmt.Println("✅ 行程 JSON 通过 LLM 修复成功解析")
		return &plan, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("未知解析错误")
	}
	return nil, fmt.Errorf("行程 JSON 解析失败: %v", lastErr)
}

// extractJSONFromResponse 从模型响应中截取 JSON 文本(兼容代码块与截断场景)。
func extractJSONFromResponse(response string) string {
	if idx := strings.Index(response, "```json"); idx >= 0 {
		start := idx + len("```json")
		if end := strings.Index(response[start:], "```"); end >= 0 {
			return strings.TrimSpace(response[start : start+end])
		}
		return strings.TrimSpace(response[start:])
	}
	if idx := strings.Index(response, "```"); idx >= 0 {
		start := idx + 3
		if end := strings.Index(response[start:], "```"); end >= 0 {
			return strings.TrimSpace(response[start : start+end])
		}
		return strings.TrimSpace(response[start:])
	}
	if start := strings.Index(response, "{"); start >= 0 {
		if end := strings.LastIndex(response, "}"); end > start {
			return response[start : end+1]
		}
		return response[start:]
	}
	return ""
}