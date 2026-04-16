package biz

import (
	"github.com/hobbyGG/Dawnix/internal/domain"
)

// 十字链表
type RuntimeGraph struct {
	Nodes map[string]*domain.NodeModel   // nodeID->NodeConfig的映射
	Next  map[string][]*domain.EdgeModel // nodeID->Edge 以node为起点的所有边
	Prev  map[string][]*domain.EdgeModel // nodeID->Edge 以node为终点的所有边

	StartNode *domain.NodeModel
}

func NewSchedulerRuntimeGraph(graphModel *domain.GraphModel) *RuntimeGraph {
	schedulerGraph := &RuntimeGraph{
		Nodes: make(map[string]*domain.NodeModel),
		Next:  make(map[string][]*domain.EdgeModel),
		Prev:  make(map[string][]*domain.EdgeModel),
	}
	for _, nodeModel := range graphModel.Nodes {
		node := &domain.NodeModel{
			ID:         nodeModel.ID,
			Type:       nodeModel.Type,
			Name:       nodeModel.Name,
			Candidates: nodeModel.Candidates,
			Properties: nodeModel.Properties,
		}
		schedulerGraph.Nodes[node.ID] = node
		if node.Type == domain.NodeTypeStart {
			schedulerGraph.StartNode = node
		}
	}
	// 建立Next与Prev
	for _, edgeModel := range graphModel.Edges {
		edge := domain.EdgeModel{
			ID:         edgeModel.ID,
			SourceNode: edgeModel.SourceNode,
			TargetNode: edgeModel.TargetNode,
		}
		schedulerGraph.Next[edge.SourceNode] = append(schedulerGraph.Next[edge.SourceNode], &edge)
		schedulerGraph.Prev[edge.TargetNode] = append(schedulerGraph.Prev[edge.TargetNode], &edge)
	}
	return schedulerGraph
}
