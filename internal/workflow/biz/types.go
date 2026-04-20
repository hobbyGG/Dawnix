package biz

import (
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

// 十字链表
type RuntimeGraph struct {
	Nodes map[string]Node                // nodeID->NodeConfig的映射
	Next  map[string][]*domain.EdgeModel // nodeID->Edge 以node为起点的所有边
	Prev  map[string][]*domain.EdgeModel // nodeID->Edge 以node为终点的所有边

	StartNode Node
}

// FormDataItem is a generic form item used by both definition and instance payloads.
type FormDataItem struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
	Type  string          `json:"type"`
}

func DecodeFormDataItems(payload []byte) ([]FormDataItem, error) {
	if len(payload) == 0 {
		return []FormDataItem{}, nil
	}

	var items []FormDataItem
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, fmt.Errorf("unmarshal form data items failed: %w", err)
	}
	return items, nil
}

func FormDataItemsToMap(items []FormDataItem) (map[string]interface{}, error) {
	values := make(map[string]interface{}, len(items))
	for _, item := range items {
		if item.Key == "" {
			continue
		}

		var val interface{}
		if len(item.Value) > 0 {
			if err := json.Unmarshal(item.Value, &val); err != nil {
				return nil, fmt.Errorf("unmarshal form item value for key %s failed: %w", item.Key, err)
			}
		}
		values[item.Key] = val
	}
	return values, nil
}

func MergeFormDataItems(base []FormDataItem, incoming []FormDataItem) []FormDataItem {
	if len(base) == 0 {
		return append([]FormDataItem{}, incoming...)
	}

	merged := append([]FormDataItem{}, base...)
	indexByKey := make(map[string]int, len(merged))
	for i, item := range merged {
		if item.Key == "" {
			continue
		}
		indexByKey[item.Key] = i
	}

	for _, item := range incoming {
		if item.Key == "" {
			continue
		}
		if idx, ok := indexByKey[item.Key]; ok {
			merged[idx] = item
			continue
		}
		indexByKey[item.Key] = len(merged)
		merged = append(merged, item)
	}

	return merged
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
			Condition:  edgeModel.Condition,
			IsDefault:  edgeModel.IsDefault,
		}
		schedulerGraph.Next[edge.SourceNode] = append(schedulerGraph.Next[edge.SourceNode], &edge)
		schedulerGraph.Prev[edge.TargetNode] = append(schedulerGraph.Prev[edge.TargetNode], &edge)
	}
	return schedulerGraph, nil
}
