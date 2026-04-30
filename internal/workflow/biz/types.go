package biz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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
	ID    string          `json:"id"`
	Label string          `json:"label"`
	Value json.RawMessage `json:"value"`
	Type  string          `json:"type"`
}

type FormValidationMode string

const (
	FormValidationModeDefinition FormValidationMode = "definition"
	FormValidationModeRuntime    FormValidationMode = "runtime"
)

var formTypeAliasToCanonical = map[string]string{
	"text_single_line": domain.FormTypeTextSingleLine,
	"single_line_text": domain.FormTypeTextSingleLine,
	"text":             domain.FormTypeTextSingleLine,
	"string":           domain.FormTypeTextSingleLine,

	"number": domain.FormTypeNumber,

	"single_select": domain.FormTypeSingleSelect,
	"select":        domain.FormTypeSingleSelect,
	"dropdown":      domain.FormTypeSingleSelect,
	"single_choice": domain.FormTypeSingleSelect,

	"date": domain.FormTypeDate,
}

func DecodeFormDataItems(payload []byte) ([]FormDataItem, error) {
	if len(payload) == 0 || bytes.Equal(payload, []byte("null")) || bytes.Equal(payload, []byte("{}")) {
		return []FormDataItem{}, nil
	}

	var items []FormDataItem
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, fmt.Errorf("unmarshal form data items failed: %w", err)
	}
	return items, nil
}

// func ValidateFormDataItems(items []FormDataItem, mode FormValidationMode) error {
// 	seenIdentity := make(map[string]struct{}, len(items))
// 	for i, item := range items {
// 		ref := formItemRef(item, i)

// 		if item.ID == "" {
// 			return fmt.Errorf("%s id is required", ref)
// 		}
// 		if item.Label == "" {
// 			return fmt.Errorf("%s label is required", ref)
// 		}
// 		if item.Type == "" {
// 			return fmt.Errorf("%s type is required", ref)
// 		}

// 		canonicalType, err := normalizeFormType(item.Type)
// 		if err != nil {
// 			return fmt.Errorf("%s type is invalid: %w", ref, err)
// 		}

// 		identity := formItemIdentity(item)
// 		if _, exists := seenIdentity[identity]; exists {
// 			return fmt.Errorf("%s duplicated identity: %s", ref, identity)
// 		}
// 		seenIdentity[identity] = struct{}{}

// 		if len(item.Value) == 0 {
// 			if mode == FormValidationModeRuntime {
// 				return fmt.Errorf("%s value is required", ref)
// 			}
// 			continue
// 		}
// 		if !json.Valid(item.Value) {
// 			return fmt.Errorf("%s value must be valid json", ref)
// 		}

// 		if err := ValidateFormValueByType(canonicalType, item.Value); err != nil {
// 			return fmt.Errorf("%s value is invalid for type %s: %w", ref, canonicalType, err)
// 		}
// 	}
// 	return nil
// }

// func ValidateFormValueByType(typ string, raw json.RawMessage) error {
// 	switch typ {
// 	case domain.FormTypeTextSingleLine:
// 		var value string
// 		if err := json.Unmarshal(raw, &value); err != nil {
// 			return fmt.Errorf("expect string: %w", err)
// 		}
// 		return nil
// 	case domain.FormTypeNumber:
// 		var value float64
// 		if err := json.Unmarshal(raw, &value); err != nil {
// 			return fmt.Errorf("expect number: %w", err)
// 		}
// 		return nil
// 	case domain.FormTypeSingleSelect:
// 		var value interface{}
// 		if err := json.Unmarshal(raw, &value); err != nil {
// 			return fmt.Errorf("expect json primitive: %w", err)
// 		}
// 		switch value.(type) {
// 		case string, float64, bool:
// 			return nil
// 		default:
// 			return fmt.Errorf("expect string/number/boolean")
// 		}
// 	case domain.FormTypeDate:
// 		var value string
// 		if err := json.Unmarshal(raw, &value); err != nil {
// 			return fmt.Errorf("expect RFC3339 datetime string: %w", err)
// 		}
// 		if _, err := time.Parse(time.RFC3339, value); err != nil {
// 			return fmt.Errorf("expect RFC3339 datetime string: %w", err)
// 		}
// 		return nil
// 	default:
// 		return fmt.Errorf("unsupported form type: %s", typ)
// 	}
// }

func FormDataItemsToMap(items []FormDataItem) (map[string]interface{}, error) {
	values := make(map[string]interface{}, len(items))
	for _, item := range items {
		if item.Label == "" {
			continue
		}

		var val interface{}
		if len(item.Value) > 0 {
			if err := json.Unmarshal(item.Value, &val); err != nil {
				return nil, fmt.Errorf("unmarshal form item value for label %s failed: %w", item.Label, err)
			}
		}
		values[item.Label] = val
	}
	return values, nil
}

func MergeFormDataItems(base []FormDataItem, incoming []FormDataItem) []FormDataItem {
	if len(base) == 0 {
		return append([]FormDataItem{}, incoming...)
	}

	merged := append([]FormDataItem{}, base...)
	indexByIdentity := make(map[string]int, len(merged))
	for i, item := range merged {
		identity := formItemIdentity(item)
		if identity == "" {
			continue
		}
		indexByIdentity[identity] = i
	}

	for _, item := range incoming {
		identity := formItemIdentity(item)
		if identity == "" {
			continue
		}
		if idx, ok := indexByIdentity[identity]; ok {
			merged[idx] = item
			continue
		}
		indexByIdentity[identity] = len(merged)
		merged = append(merged, item)
	}

	return merged
}

func formItemIdentity(item FormDataItem) string {
	if item.ID != "" {
		return item.ID
	}
	return item.Label
}

func formItemRef(item FormDataItem, idx int) string {
	if item.ID != "" {
		return fmt.Sprintf("item[%d](id=%s)", idx, item.ID)
	}
	if item.Label != "" {
		return fmt.Sprintf("item[%d](label=%s)", idx, item.Label)
	}
	return fmt.Sprintf("item[%d]", idx)
}

func normalizeFormType(rawType string) (string, error) {
	normalized := strings.ToLower(rawType)
	if normalized == "" {
		return "", fmt.Errorf("type is empty")
	}
	if canonical, ok := formTypeAliasToCanonical[normalized]; ok {
		return canonical, nil
	}
	switch normalized {
	case "member", "contact", "member_picker", "user", "user_picker", "attachment", "file", "image":
		return "", fmt.Errorf("form type %s is not supported in current version", rawType)
	default:
		return "", fmt.Errorf("unsupported form type: %s", rawType)
	}
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
