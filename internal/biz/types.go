package biz

import (
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

// 十字链表
type RuntimeGraph struct {
	Nodes map[string]*model.NodeModel   // nodeID->NodeConfig的映射
	Next  map[string][]*model.EdgeModel // nodeID->Edge 以node为起点的所有边
	Prev  map[string][]*model.EdgeModel // nodeID->Edge 以node为终点的所有边

	StartNode *model.NodeModel
}

func NewSchedulerRuntimeGraph(graphModel *model.GraphModel) *RuntimeGraph {
	schedulerGraph := &RuntimeGraph{
		Nodes: make(map[string]*model.NodeModel),
		Next:  make(map[string][]*model.EdgeModel),
		Prev:  make(map[string][]*model.EdgeModel),
	}
	for _, nodeModel := range graphModel.Nodes {
		node := &model.NodeModel{
			ID:         nodeModel.ID,
			Type:       nodeModel.Type,
			Name:       nodeModel.Name,
			Candidates: nodeModel.Candidates,
		}
		schedulerGraph.Nodes[node.ID] = node
		if node.Type == model.NodeTypeStart {
			schedulerGraph.StartNode = node
		}
	}
	// 建立Next与Prev
	for _, edgeModel := range graphModel.Edges {
		edge := model.EdgeModel{
			ID:         edgeModel.ID,
			SourceNode: edgeModel.SourceNode,
			TargetNode: edgeModel.TargetNode,
		}
		schedulerGraph.Next[edge.SourceNode] = append(schedulerGraph.Next[edge.SourceNode], &edge)
		schedulerGraph.Prev[edge.TargetNode] = append(schedulerGraph.Prev[edge.TargetNode], &edge)
	}
	return schedulerGraph
}
