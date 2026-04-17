package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/domain"
	"go.uber.org/zap"
)

type ProcessDefinitionService struct {
	repo   biz.ProcessDefinitionRepo
	logger *zap.Logger
}

func NewProcessDefinitionService(repo biz.ProcessDefinitionRepo, logger *zap.Logger) *ProcessDefinitionService {
	return &ProcessDefinitionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ProcessDefinitionService) CreateProcessDefinition(c context.Context, params *biz.ProcessDefinitionCreateParams) (int64, error) {
	// 这里实现创建流程模板的业务逻辑
	// 业务校验：流程是否已经存在，唯一字段是否存在冲突
	if params == nil {
		return 0, fmt.Errorf("params is nil")
	}
	if params.Structure == nil {
		return 0, fmt.Errorf("structure is nil")
	}

	if err := validateFormDefinition(params.FormDefinition); err != nil {
		return 0, fmt.Errorf("invalid form_definition: %w", err)
	}
	if err := validateGraphStructure(params.Structure); err != nil {
		return 0, fmt.Errorf("invalid structure graph: %w", err)
	}
	if err := validateXORRoutingRules(params.Structure); err != nil {
		return 0, fmt.Errorf("invalid structure routing rules: %w", err)
	}
	if err := validateInclusiveRoutingRules(params.Structure); err != nil {
		return 0, fmt.Errorf("invalid structure routing rules: %w", err)
	}

	for _, node := range params.Structure.Nodes {
		if node.Type == domain.NodeTypeEmailService {
			// 验证参数
			var emailParams domain.EmailNodeParams
			if err := json.Unmarshal(node.Properties, &emailParams); err != nil {
				return 0, fmt.Errorf("fail to unmarshal email service properties: %w", err)
			}
		}
	}

	model, err := paramsToProcessDef(params)
	if err != nil {
		return 0, fmt.Errorf("convert request to model failed: %w", err)
	}
	id, err := s.repo.Create(c, model)
	if err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}
	return id, nil
}

func validateFormDefinition(items []biz.FormDataItem) error {
	for i, item := range items {
		if item.Key == "" {
			return fmt.Errorf("item[%d].key is required", i)
		}
		if item.Type == "" {
			return fmt.Errorf("item[%d].type is required", i)
		}
		if len(item.Value) == 0 {
			return fmt.Errorf("item[%d].value is required", i)
		}
		if !json.Valid(item.Value) {
			return fmt.Errorf("item[%d].value must be valid json", i)
		}
	}
	return nil
}

func validateGraphStructure(graph *domain.GraphModel) error {
	if graph == nil {
		return fmt.Errorf("graph is nil")
	}
	if len(graph.Nodes) == 0 {
		return fmt.Errorf("graph must contain at least one node")
	}

	nodeByID := make(map[string]domain.NodeModel, len(graph.Nodes))
	startCount := 0
	endCount := 0
	for i, node := range graph.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node[%d].id is required", i)
		}
		if node.Type == "" {
			return fmt.Errorf("node[%d].type is required", i)
		}
		if _, exists := nodeByID[node.ID]; exists {
			return fmt.Errorf("duplicate node id: %s", node.ID)
		}
		nodeByID[node.ID] = node

		switch node.Type {
		case domain.NodeTypeStart:
			startCount++
		case domain.NodeTypeEnd:
			endCount++
		}
	}
	if startCount != 1 {
		return fmt.Errorf("graph must contain exactly one start node, got %d", startCount)
	}
	if endCount == 0 {
		return fmt.Errorf("graph must contain at least one end node")
	}

	inDegree := make(map[string]int, len(nodeByID))
	outDegree := make(map[string]int, len(nodeByID))
	edgeIDs := make(map[string]struct{}, len(graph.Edges))
	for i, edge := range graph.Edges {
		if edge.ID == "" {
			return fmt.Errorf("edge[%d].id is required", i)
		}
		if _, exists := edgeIDs[edge.ID]; exists {
			return fmt.Errorf("duplicate edge id: %s", edge.ID)
		}
		edgeIDs[edge.ID] = struct{}{}

		if edge.SourceNode == "" || edge.TargetNode == "" {
			return fmt.Errorf("edge %s source and target are required", edge.ID)
		}
		if _, exists := nodeByID[edge.SourceNode]; !exists {
			return fmt.Errorf("edge %s source node not found: %s", edge.ID, edge.SourceNode)
		}
		if _, exists := nodeByID[edge.TargetNode]; !exists {
			return fmt.Errorf("edge %s target node not found: %s", edge.ID, edge.TargetNode)
		}

		outDegree[edge.SourceNode]++
		inDegree[edge.TargetNode]++
	}

	for _, node := range graph.Nodes {
		if node.Type == domain.NodeTypeStart && inDegree[node.ID] > 0 {
			return fmt.Errorf("start node %s cannot have incoming edges", node.ID)
		}
		if node.Type != domain.NodeTypeEnd && outDegree[node.ID] == 0 {
			return fmt.Errorf("node %s has no outgoing edges", node.ID)
		}
		if node.Type == domain.NodeTypeEnd && outDegree[node.ID] > 0 {
			return fmt.Errorf("end node %s cannot have outgoing edges", node.ID)
		}
	}

	return nil
}

func validateXORRoutingRules(graph *domain.GraphModel) error {
	if graph == nil {
		return fmt.Errorf("graph is nil")
	}

	edgesBySource := graph.EdgesBySource()

	for _, node := range graph.Nodes {
		if node.Type != domain.NodeTypeXORGateway {
			continue
		}

		edges := edgesBySource[node.ID]
		if len(edges) == 0 {
			return fmt.Errorf("xor node %s has no outgoing edges", node.ID)
		}

		defaultCount := 0
		for _, edge := range edges {
			condition := edge.Condition
			if edge.IsDefault {
				defaultCount++
				if condition != "" {
					return fmt.Errorf("xor node %s edge %s: default edge cannot define condition", node.ID, edge.ID)
				}
				continue
			}

			if condition == "" {
				return fmt.Errorf("xor node %s edge %s: condition is required", node.ID, edge.ID)
			}
		}

		if defaultCount > 1 {
			return fmt.Errorf("xor node %s has %d default edges, only one is allowed", node.ID, defaultCount)
		}
	}

	return nil
}

func validateInclusiveRoutingRules(graph *domain.GraphModel) error {
	if graph == nil {
		return fmt.Errorf("graph is nil")
	}

	edgesBySource := graph.EdgesBySource()

	for _, node := range graph.Nodes {
		if node.Type != domain.NodeTypeInclusiveGateway {
			continue
		}

		edges := edgesBySource[node.ID]
		if len(edges) == 0 {
			return fmt.Errorf("inclusive node %s has no outgoing edges", node.ID)
		}

		defaultCount := 0
		for _, edge := range edges {
			if edge.IsDefault {
				defaultCount++
				if edge.Condition != "" {
					return fmt.Errorf("inclusive node %s edge %s: default edge cannot define condition", node.ID, edge.ID)
				}
			}
		}

		if defaultCount > 1 {
			return fmt.Errorf("inclusive node %s has %d default edges, only one is allowed", node.ID, defaultCount)
		}
	}

	return nil
}

func (s *ProcessDefinitionService) ListProcessDefinitions(ctx context.Context, params *biz.ProcessDefinitionListParams) ([]domain.ProcessDefinition, error) {
	// 这里实现获取流程模板列表的业务逻辑
	// 业务校验：分页参数是否合法
	pdList, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fail to get processDefinition list: %w", err)
	}
	return pdList, nil
}

func (s *ProcessDefinitionService) GetProcessDefinitionDetail(ctx context.Context, id int64) (*domain.ProcessDefinition, error) {
	// 这里实现获取流程模板详情的业务逻辑
	pdDetail, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fail to get processDefinition detail: %w", err)
	}
	return pdDetail, nil
}

func (s *ProcessDefinitionService) DeleteProcessDefinition(ctx context.Context, id int64) error {
	// 这里实现删除流程模板的业务逻辑
	// 业务校验：流程模板是否存在，是否允许删除等

	s.repo.DeleteByID(ctx, id)
	return nil
}

func (s *ProcessDefinitionService) UpdateProcessDefinition(ctx context.Context, model *domain.ProcessDefinition) error {
	// 这里实现更新流程模板的业务逻辑
	// 业务校验：流程模板是否存在，唯一字段是否冲突等
	err := s.repo.Update(ctx, model)
	if err != nil {
		return fmt.Errorf("fail to update processDefinition: %w", err)
	}
	return nil
}
