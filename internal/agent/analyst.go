package agent

import (
	"E-plan/internal/tools"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const maxSteps = 5

// analystState 是分析师子图内部流转的状态
type analystState struct {
	messages []*schema.Message // 消息历史（包含 LLM 回复和工具执行结果）
	finalOut *schema.Message   // 最终输出给主图的消息
	step     int
}

// 定义大模型可见的参数结构
type WeatherArgs struct {
	Location string `json:"location"`
}

// 定义返回给大模型的结果结构
type WeatherResult struct {
	City        string `json:"city"`
	Condition   string `json:"condition"`
	Temperature string `json:"temperature"`
}

// BuildAnalystAgent 构建分析师子图 (标准的 Tool Call Loop 编排)
func BuildAnalystAgent(ctx context.Context, chatModel model.ToolCallingChatModel, weatherClient *tools.WeatherClient) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
	weatherTool := utils.NewTool(&schema.ToolInfo{
		Name: "get_weather_info",
		Desc: "获取特定城市的天气预报",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"location": {
					Type:     schema.String,
					Desc:     "城市名称，例如北京",
					Required: true,
				},
			},
		),
	}, func(ctx context.Context, args *WeatherArgs) (*WeatherResult, error) {
		// 🚀 在这里调用 WeatherClient 的方法
		domainWeather, err := weatherClient.GetTomorrowWeather(ctx, args.Location)
		if err != nil {
			return nil, err
		}

		// 转换为工具返回的格式
		return &WeatherResult{
			City:        domainWeather.City,
			Condition:   domainWeather.Condition,
			Temperature: domainWeather.Temperature,
		}, nil
	})

	// 收集所有工具
	expertTools := []tool.BaseTool{weatherTool}

	// 提取 ToolInfo 用于告知 LLM
	var toolSpecs []*schema.ToolInfo
	for _, t := range expertTools {
		info, _ := t.Info(ctx)
		toolSpecs = append(toolSpecs, info)
	}
	// 💡 1. 核心修复：直接定义子图的整体出入参，消除外部 Adapter
	g := compose.NewGraph[[]*schema.Message, *schema.Message]()

	// --- 节点 0：初始化状态 (InitNode) ---
	// 将外部传入的 Message 切片包装成内部流转的 analystState
	g.AddLambdaNode("InitNode", compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*analystState, error) {
		return &analystState{
			messages: append([]*schema.Message{}, input...),
		}, nil
	}))

	// --- 节点 1：LLM 推理节点 ---
	g.AddLambdaNode("LLMNode", compose.InvokableLambda(func(ctx context.Context, state *analystState) (*analystState, error) {
		if state.step > maxSteps {
			return nil, fmt.Errorf("超过最大推理步数，可能出现死循环")
		}

		resp, err := chatModel.Generate(ctx, state.messages)
		if err != nil {
			return nil, err
		}

		state.messages = append(state.messages, resp)
		state.step++

		// ✅ 只有非 tool_call 才认为是最终输出
		if len(resp.ToolCalls) == 0 {
			state.finalOut = resp
		}

		return state, nil
	}))

	// --- 节点 2：工具执行节点 ---
	g.AddLambdaNode("ToolExecuteNode", compose.InvokableLambda(func(ctx context.Context, state *analystState) (*analystState, error) {
		lastMsg := state.messages[len(state.messages)-1]

		for _, tc := range lastMsg.ToolCalls {
			var result string

			// --- 💡 逻辑分发：根据注册的名字路由 ---
			switch tc.Function.Name {
			case "get_weather_info":
				// 1. 解析参数 (LLM 传过来的是 JSON 字符串)
				// 你可以使用 json.Unmarshal 解析 tc.Function.Arguments
				result = `{"temp": "22C", "condition": "Sunny"}` // 这里接你真实的 WeatherClient

			case "query_user_health_data":
				result = `{"average_heart_rate": 75}` // 这里接你的 Repository

			default:
				result = fmt.Sprintf("错误：工具 %s 未定义", tc.Function.Name)
			}

			// 将结果包装成消息
			state.messages = append(state.messages, &schema.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID, // 必须匹配，否则 LLM 会报错
			})
		}
		return state, nil
	}))

	// --- 节点 3：提取最终输出 (OutputNode) ---
	g.AddLambdaNode("OutputNode", compose.InvokableLambda(func(ctx context.Context, state *analystState) (*schema.Message, error) {
		return state.finalOut, nil
	}))

	// --- 💡 2. 核心修复：路由分支的正确写法 ---
	g.AddBranch("CheckToolBranch", compose.NewGraphBranch(func(ctx context.Context, state *analystState) (string, error) {
		lastMsg := state.messages[len(state.messages)-1]

		if len(lastMsg.ToolCalls) > 0 {
			// 需要调工具，返回对应的节点名
			return "ToolExecuteNode", nil
		}

		// 不需要调工具，去往输出提取节点
		return "OutputNode", nil
	}, map[string]bool{
		// 必须在这里穷举分支函数所有可能返回的节点名
		"ToolExecuteNode": true,
		"OutputNode":      true,
	}))
	// --- 💡 3. 核心修复：Eino 连线机制 ---
	g.AddEdge(compose.START, "InitNode")
	g.AddEdge("InitNode", "LLMNode")

	// LLM 思考完后，送入分支判断器
	g.AddEdge("LLMNode", "CheckToolBranch")

	// 在 Eino 中，分支的可能去向必须通过 AddEdge 显式连线声明，不需要 Map
	g.AddEdge("CheckToolBranch", "ToolExecuteNode")
	g.AddEdge("CheckToolBranch", "OutputNode")

	// 工具执行完后，必须连回 LLM 形成闭环 (ReAct 循环)
	g.AddEdge("ToolExecuteNode", "LLMNode")

	// 输出提取完毕后，安全结束图
	g.AddEdge("OutputNode", compose.END)

	// 编译子图。因为 Graph 的出入参类型与函数声明完全一致，直接返回即可
	compiledGraph, err := g.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译分析师子图失败: %w", err)
	}

	return compiledGraph, nil
}
