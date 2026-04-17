package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"E-plan/internal/domain"

	// 引入 eino 的包
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// LLMClient 包装了 Eino 的 ChatModel，并提供了业务降级（Heuristic）能力
type LLMClient struct {
	chatModel model.ToolCallingChatModel
}

// NewLLMClient 构造函数
// 你可以在 main.go 里初始化好 Eino 的 OpenAI/豆包/Gemini 模型传进来。
// 如果传入的 chatModel 为 nil，则默认全走启发式兜底。
func NewLLMClient(chatModel model.ToolCallingChatModel) *LLMClient {
	return &LLMClient{
		chatModel: chatModel,
	}
}

// Generate 调用大模型获取响应。如果模型不可用或调用报错，降级为本地规则生成。
func (c *LLMClient) Generate(ctx context.Context, systemPrompt string, userMessage string, user *domain.User, summary *domain.DailySummary, forecast *domain.WeatherCondition) (string, error) {
	// 1. 如果没有配置大模型（初始化时传了 nil），直接走降级逻辑
	if c.chatModel == nil {
		return heuristicLLM(systemPrompt, userMessage, user, summary, forecast), nil
	}

	// 2. 将字符串组装成 Eino 标准的 Message 切片
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMessage),
	}

	// 3. 直接调用 Eino 模型 (自带超时、重试等配置)
	// 如果需要在生成时强制输出 JSON，可以通过 options 传入参数 (具体取决于使用的 Eino Provider 插件)
	resp, err := c.chatModel.Generate(ctx, messages)

	if err != nil {
		// 4. 如果大模型网络超时或出错，触发降级保护，依然让系统可用
		// fmt.Printf("[Warning] LLM调用失败, 触发降级: %v\n", err)
		return heuristicLLM(systemPrompt, userMessage, user, summary, forecast), nil
	}

	// 5. 直接返回大模型的文本内容
	return resp.Content, nil
}

func heuristicLLM(systemPrompt string, userMessage string, user *domain.User, summary *domain.DailySummary, forecast *domain.WeatherCondition) string {
	// Minimal, stable heuristic output that matches ParseLLMOutput expectations.
	// This keeps /v1/agent/daily usable without an API key.
	intensity := "low"
	fatigue := "low"

	totalMins := 0
	totalLoad := 0
	feelingSum := 0
	feelingN := 0
	for _, s := range summary.Sessions {
		totalMins += s.Base.DurationSecs / 60
		if s.Advanced != nil {
			totalLoad += s.Advanced.TrainingLoad
		}
		if s.Base.FeelingScore > 0 {
			feelingSum += s.Base.FeelingScore
			feelingN++
		}
	}

	if totalLoad >= 250 || totalMins >= 90 {
		intensity = "high"
	} else if totalLoad >= 120 || totalMins >= 45 {
		intensity = "medium"
	}

	if summary.SleepHours > 0 && summary.SleepHours < 6 {
		fatigue = "high"
	} else if summary.SleepHours > 0 && summary.SleepHours < 7 {
		fatigue = "medium"
	}

	if feelingN > 0 {
		avgFeeling := float64(feelingSum) / float64(feelingN)
		if avgFeeling <= 4 {
			fatigue = "high"
		} else if avgFeeling <= 6 && fatigue == "low" {
			fatigue = "medium"
		}
	}

	score := 20
	switch intensity {
	case "low":
		score += 10
	case "medium":
		score += 0
	case "high":
		score -= 10
	}
	switch fatigue {
	case "low":
		score += 10
	case "medium":
		score -= 5
	case "high":
		score -= 20
	}
	if score > 100 {
		score = 100
	}
	if score < -100 {
		score = -100
	}

	bodyStatus := "Recovered"
	if fatigue == "high" {
		bodyStatus = "Fatigued"
	} else if fatigue == "medium" {
		bodyStatus = "Slightly tired"
	}

	highlights := []string{
		fmt.Sprintf("Training time: %d min", totalMins),
		fmt.Sprintf("Estimated intensity: %s", intensity),
	}
	if totalLoad > 0 {
		highlights = append(highlights, fmt.Sprintf("Training load: %d", totalLoad))
	}
	warnings := []string{}
	if fatigue == "high" {
		warnings = append(warnings, "High fatigue signal: prioritize recovery and sleep.")
	}
	if forecast != nil && strings.Contains(strings.ToLower(forecast.Condition), "rain") {
		warnings = append(warnings, "Rain forecast: consider indoor alternatives or adjust route.")
	}

	// Plan suggestion.
	activityType := string(domain.ActivityGeneral)
	title := "Easy recovery session"
	instructions := "10 min warm-up + 20-30 min easy effort + 10 min cool-down. Keep it comfortable."
	target := domain.PlanTargets{DurationSecs: 40 * 60}

	if user != nil {
		switch user.Focus {
		case domain.FocusRunning:
			activityType = string(domain.ActivityRunning)
			title = "Easy run (recovery)"
			target = domain.PlanTargets{DistanceKm: 5, TargetHRZone: "Zone 2"}
		case domain.FocusFitness:
			activityType = string(domain.ActivityStrength)
			title = "Full-body technique day"
			instructions = "Keep weights light to moderate. Focus on form. Stop 2-3 reps before failure."
			target = domain.PlanTargets{TotalVolumeKg: 0}
		case domain.FocusCycling:
			activityType = string(domain.ActivityCycling)
			title = "Easy spin (recovery)"
			target = domain.PlanTargets{DurationSecs: 45 * 60, TargetHRZone: "Zone 2"}
		}
	}

	if fatigue == "low" && intensity != "high" {
		title = "Build session"
		instructions = "Warm-up 10 min. Main: 3 x 6 min steady (moderate-hard) with 3 min easy. Cool-down 10 min."
		if activityType == string(domain.ActivityRunning) {
			target = domain.PlanTargets{DurationSecs: 55 * 60}
		} else if activityType == string(domain.ActivityCycling) {
			target = domain.PlanTargets{DurationSecs: 60 * 60}
		}
	}
	if fatigue == "high" {
		title = "Rest or mobility"
		activityType = string(domain.ActivityGeneral)
		instructions = "Optional: 20-30 min easy walk + 10 min mobility. If soreness is high, take full rest."
		target = domain.PlanTargets{DurationSecs: 30 * 60}
	}

	out := map[string]any{
		"report": map[string]any{
			"body_status": bodyStatus,
			"summary":     fmt.Sprintf("Score %d. Intensity=%s, fatigue=%s. %s", score, intensity, fatigue, pickSummary(user, intensity, fatigue)),
			"highlights":  highlights,
			"warnings":    warnings,
		},
		"plan": map[string]any{
			"activity_type":    activityType,
			"title":            title,
			"instructions":     instructions,
			"target_metrics":   target,
			"reasoning":        "Heuristic fallback (no LLM API key configured).",
			"forecast_weather": map[string]any{}, // ignored by parser/server-fill; kept for compatibility
		},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func pickSummary(user *domain.User, intensity string, fatigue string) string {
	if user == nil {
		return "Adjust tomorrow based on recovery and consistency."
	}
	if fatigue == "high" {
		return "Tomorrow should be recovery-first to avoid injury risk."
	}
	if intensity == "high" {
		return "Tomorrow should be easy to consolidate gains."
	}
	return "Tomorrow can progress slightly while staying controlled."
}
