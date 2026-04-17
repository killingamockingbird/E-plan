package agent

import (
	"E-plan/internal/domain"
	"github.com/cloudwego/eino/schema"
)

// AgentState 定义了在 Graph 中流转的全局数据总线
type AgentState struct {
	// 1. 初始输入 (由 Handler 传入)
	UserID    string
	SessionID string
	TodayData *domain.DailySummary

	// 2. 过程加载数据 (由前面的 Node 填充，供后面的 Node 使用)
	UserProfile  *domain.User
	PastWeekData []*domain.DailySummary
	Forecast     *domain.WeatherCondition // 可能为空（如果路由判断不需要查天气）

	// 3. Prompt 渲染结果
	SystemPrompt string

	// 4. 大模型输出与解析结果
	RawLLMOutput string
	FinalPlan    *domain.TrainingPlan
	FinalReport  *domain.AnalysisReport

	HistoryMessages []*schema.Message
}
