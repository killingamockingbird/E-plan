package agent

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"time"

	"E-plan/internal/prompt"
	"E-plan/internal/repository"
	"E-plan/internal/tools"
)

// AgentNodes 包含所有图节点的实现，并注入所需的外部依赖
type AgentNodes struct {
	UserRepo    repository.UserRepository
	DataRepo    repository.SportsDataRepository
	PlanRepo    repository.PlanRepository
	HistoryRepo repository.HistoryRepository
	PromptMgr   *prompt.Manager
	WeatherAPI  *tools.WeatherClient

	// 🚀 恢复：双 Agent 架构
	AnalystAgent compose.Runnable[[]*schema.Message, *schema.Message] // 负责深度分析的 ADK 专家
	CoachClient  *LLMClient                                           // 负责写计划的主教练 (使用你带兜底逻辑的封装)
}

// 节点 1：加载用户画像和历史记忆 (Context Loading)
func (n *AgentNodes) LoadContextNode(ctx context.Context, state *AgentState) (*AgentState, error) {
	// 1. 查用户信息
	user, err := n.UserRepo.GetUser(ctx, state.UserID)
	if err != nil {
		return state, err
	}
	state.UserProfile = user

	// 2. 查过去 7 天数据作为 Agent 的“长期记忆”
	pastData, err := n.DataRepo.GetRecentSummaries(ctx, state.UserID, time.Now().AddDate(0, 0, -7))
	if err == nil {
		state.PastWeekData = pastData
	}

	// 💡 读取历史对话
	history, err := n.HistoryRepo.LoadHistory(ctx, state.UserID)
	if err != nil {
		return nil, err
	}
	state.HistoryMessages = history

	return state, nil
}

// 节点 2：工具调用 - 查天气 (Tool Calling)
func (n *AgentNodes) FetchWeatherNode(ctx context.Context, state *AgentState) (*AgentState, error) {
	// 优化：优先从用户画像中提取城市，找不到再默认用“北京”
	location := "北京"
	if state.UserProfile != nil && state.UserProfile.City != "" { // 假设你的 User 结构体里有 City 字段
		location = state.UserProfile.City
	}

	weather, err := n.WeatherAPI.GetTomorrowWeather(ctx, location)
	if err == nil {
		state.Forecast = weather
	}
	return state, nil
}

// 节点 3：动态渲染 Prompt
func (n *AgentNodes) BuildPromptNode(ctx context.Context, state *AgentState) (*AgentState, error) {
	// 将加载到的所有上下文传递给 Prompt Manager 进行模板渲染
	sysPrompt, err := n.PromptMgr.BuildSystemPrompt(state.UserProfile, state.PastWeekData, state.Forecast)
	if err != nil {
		return state, err
	}
	state.SystemPrompt = sysPrompt
	return state, nil
}

func (n *AgentNodes) AnalystNode(ctx context.Context, state *AgentState) (*AgentState, error) {
	if state.TodayData == nil {
		return state, fmt.Errorf("TodayData is missing for UserID: %s", state.UserID)
	}

	// 1. 组装给分析师的 Prompt
	sysMsg := schema.SystemMessage("你是专业数据分析师，请结合工具深度分析，并输出 JSON 格式的 AnalysisReport。")
	userMsg := schema.UserMessage(state.TodayData.ToAgentContext())
	// 💡 把历史记录拼接在前面，把今天的最新输入放在后面
	inputMessages := append(state.HistoryMessages, sysMsg, userMsg)
	// 2. 调用 ADK 分析师 Agent
	resultMsg, err := n.AnalystAgent.Invoke(ctx, inputMessages)
	if err != nil {
		return state, err
	}

	state.HistoryMessages = append(state.HistoryMessages, userMsg)
	state.HistoryMessages = append(state.HistoryMessages, resultMsg)
	// 3. 解析并落盘分析报告
	report, err := ParseAnalysisReportJSON(resultMsg.Content)
	if err == nil && report != nil {
		report.UserID = state.UserID
		state.FinalReport = report

		// 保存报告到数据库
		if err := n.PlanRepo.SaveReport(ctx, report); err != nil {
			fmt.Printf("[Error] 保存分析报告失败: %v\n", err)
		}
	} else {
		// 如果解析失败，为了不阻塞教练，可以给一个兜底的空报告
		fmt.Printf("[Warning] 解析分析报告失败: %v\n", err)
	}

	return state, nil
}

// 节点 4：主教练根据报告制定计划 (CoachNode)
func (n *AgentNodes) CoachNode(ctx context.Context, state *AgentState) (*AgentState, error) {
	// 1. 提取分析师的报告和天气作为主教练的上下文
	reportSummary := "暂无详细分析"
	if state.FinalReport != nil {
		reportSummary = state.FinalReport.Summary
	}
	weatherCondition := "未知"
	if state.Forecast != nil {
		weatherCondition = state.Forecast.Condition
	}

	// 2. 使用你的 PromptManager 生成给教练的 System Prompt
	// 这里可以把天气和历史记录传给教练
	sysPrompt, err := n.PromptMgr.BuildSystemPrompt(state.UserProfile, state.PastWeekData, state.Forecast)
	if err != nil {
		return state, err
	}

	// 3. 将分析师的结论作为 User Message 喂给主教练
	contextInfo := fmt.Sprintf("【数据分析师的报告总结】\n%s\n\n【明天天气】\n%s\n\n请严格输出 JSON 格式的 TrainingPlan。", reportSummary, weatherCondition)

	// 4. 调用主教练大模型 (带兜底逻辑)
	responseJSON, err := n.CoachClient.Generate(
		ctx, sysPrompt, contextInfo,
		state.UserProfile, state.TodayData, state.Forecast,
	)
	if err != nil {
		return state, err
	}
	state.RawLLMOutput = responseJSON

	// 5. 解析并落盘训练计划
	plan, err := ParseTrainingPlanJSON(responseJSON)
	if err == nil && plan != nil {
		plan.UserID = state.UserID
		plan.TargetDate = time.Now().AddDate(0, 0, 1)
		state.FinalPlan = plan

		if err := n.PlanRepo.SavePlan(ctx, plan); err != nil {
			fmt.Printf("[Error] 保存训练计划失败: %v\n", err)
		}
	}

	return state, nil
}
func (n *AgentNodes) SaveHistoryNode(ctx context.Context, state *AgentState) (*AgentState, error) {
	// 假设你把分析师和教练的对话都追加到了 state.HistoryMessages 里
	// 💡 落盘持久化
	err := n.HistoryRepo.SaveHistory(ctx, state.SessionID, state.UserID, state.HistoryMessages)
	if err != nil {
		fmt.Printf("保存历史会话失败: %v\n", err)
	}
	return state, nil
}
