package biz

import (
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/domain"
)

// 十字链表
type RuntimeGraph struct {
	Nodes map[string]Node                // nodeID->NodeConfig的映射
	Next  map[string][]*domain.EdgeModel // nodeID->Edge 以node为起点的所有边
	Prev  map[string][]*domain.EdgeModel // nodeID->Edge 以node为终点的所有边

	StartNode Node
}

func NewSchedulerRuntimeGraph(graphModel *domain.GraphModel, registry NodeRegistry) (*RuntimeGraph, error) {
	if graphModel == nil {
		return nil, fmt.Errorf("graph model is nil")
	}
	if registry == nil {
		return nil, fmt.Errorf("node registry is nil")
	}
	schedulerGraph := &RuntimeGraph{
		Nodes: make(map[string]Node),
		Next:  make(map[string][]*domain.EdgeModel),
		Prev:  make(map[string][]*domain.EdgeModel),
	}
	for _, nodeModel := range graphModel.Nodes {
		node, err := buildNodeFromRegistry(registry, &nodeModel)
		if err != nil {
			return nil, fmt.Errorf("build runtime node failed: %w", err)
		}
		schedulerGraph.Nodes[node.ID()] = node
		if node.Type() == domain.NodeTypeStart {
			schedulerGraph.StartNode = node
		}
	}
	if schedulerGraph.StartNode == nil {
		return nil, fmt.Errorf("start node not found")
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
	return schedulerGraph, nil
}
