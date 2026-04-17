package agent

import (
	"context"
	"github.com/cloudwego/eino/compose"
)

type Orchestrator struct {
	Nodes *AgentNodes
}

func NewOrchestrator(nodes *AgentNodes) *Orchestrator {
	return &Orchestrator{Nodes: nodes}
}

func (o *Orchestrator) Build(ctx context.Context) (compose.Runnable[*AgentState, *AgentState], error) {
	g := compose.NewGraph[*AgentState, *AgentState]()

	g.AddLambdaNode("LoadContext", compose.InvokableLambda(o.Nodes.LoadContextNode))
	g.AddLambdaNode("FetchWeather", compose.InvokableLambda(o.Nodes.FetchWeatherNode))
	g.AddLambdaNode("Analyst", compose.InvokableLambda(o.Nodes.AnalystNode)) // 分析师节点
	g.AddLambdaNode("Coach", compose.InvokableLambda(o.Nodes.CoachNode))     // 教练节点
	g.AddLambdaNode("SaveHistory", compose.InvokableLambda(o.Nodes.SaveHistoryNode))
	// 连线
	g.AddEdge(compose.START, "LoadContext")
	g.AddEdge("LoadContext", "FetchWeather")
	g.AddEdge("FetchWeather", "Analyst") // 先进分析师
	g.AddEdge("Analyst", "Coach")        // 分析师出报告后，交给教练
	g.AddEdge("Coach", compose.END)
	g.AddEdge("SaveHistory", compose.END)
	return g.Compile(ctx)
}
