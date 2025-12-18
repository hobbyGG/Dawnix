package biz

import "github.com/hobbyGG/Dawnix/internal/biz/model"

// biz层静态图结构, 与model对齐
type WorkflowGraph struct {
	Nodes []WorkflowNode
	Edges []WorkflowEdge
}

type WorkflowNode struct {
	ID   string
	Type string
	Name string
}

type WorkflowEdge struct {
	ID         string
	SourceNode string
	TargetNode string
}

// 十字链表
type RuntimeGraph struct {
	Nodes map[string]*RuntimeNode  // nodeID->NodeConfig的映射
	Next  map[string][]RuntimeEdge // nodeID->Edge 以node为起点的所有变
	Prev  map[string][]RuntimeEdge // nodeID->Edge 以node为终点的所有边

	StartNode *RuntimeNode
}

type RuntimeNode struct {
	ID   string
	Type string
	Name string
}

type RuntimeEdge struct {
	ID         string
	SourceNode string
	TargetNode string
}

func NewSchedulerGraph(graphModel *model.WorkflowGraph) *RuntimeGraph {
	schedulerGraph := &RuntimeGraph{
		Nodes: make(map[string]*RuntimeNode),
		Next:  make(map[string][]RuntimeEdge),
		Prev:  make(map[string][]RuntimeEdge),
	}
	for _, nodeModel := range graphModel.Nodes {
		node := &RuntimeNode{
			ID:   nodeModel.ID,
			Type: nodeModel.Type,
			Name: nodeModel.Name,
		}
		schedulerGraph.Nodes[node.ID] = node
		if node.Type == model.NodeTypeStart {
			schedulerGraph.StartNode = node
		}
	}
	// 建立Next与Prev
	for _, edgeModel := range graphModel.Edges {
		edge := RuntimeEdge{
			ID:         edgeModel.ID,
			SourceNode: edgeModel.SourceNode,
			TargetNode: edgeModel.TargetNode,
		}
		schedulerGraph.Next[edge.SourceNode] = append(schedulerGraph.Next[edge.SourceNode], edge)
		schedulerGraph.Prev[edge.TargetNode] = append(schedulerGraph.Prev[edge.TargetNode], edge)
	}
	return schedulerGraph
}
