package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/domain"
	"go.uber.org/zap"
)

type ProcessDefinitionService struct {
	repo                biz.ProcessDefinitionRepo
	logger              *zap.Logger
	emailServiceEnabled bool
}

func NewProcessDefinitionService(repo biz.ProcessDefinitionRepo, logger *zap.Logger, emailServiceEnabled bool) *ProcessDefinitionService {
	return &ProcessDefinitionService{
		repo:                repo,
		logger:              logger,
		emailServiceEnabled: emailServiceEnabled,
	}
}

func (s *ProcessDefinitionService) CreateProcessDefinition(c context.Context, params *biz.ProcessDefinitionCreateParams) (int64, error) {
	model, err := s.validateAndBuildModel(params)
	if err != nil {
		return 0, err
	}
	id, err := s.repo.Create(c, model)
	if err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}
	return id, nil
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

func (s *ProcessDefinitionService) UpdateProcessDefinition(ctx context.Context, id int64, params *biz.ProcessDefinitionCreateParams) error {
	if id <= 0 {
		return fmt.Errorf("id is invalid")
	}
	existModel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get processDefinition by id failed: %w", err)
	}

	model, err := s.validateAndBuildModel(params)
	if err != nil {
		return err
	}

	model.ID = existModel.ID
	model.Version = existModel.Version
	model.IsActive = existModel.IsActive
	model.CreatedAt = existModel.CreatedAt
	model.CreatedBy = existModel.CreatedBy
	if model.Code == "" {
		model.Code = existModel.Code
	}

	err = s.repo.Update(ctx, model)
	if err != nil {
		return fmt.Errorf("fail to update processDefinition: %w", err)
	}
	return nil
}

func (s *ProcessDefinitionService) validateAndBuildModel(params *biz.ProcessDefinitionCreateParams) (*domain.ProcessDefinition, error) {
	if params == nil {
		return nil, fmt.Errorf("params is nil")
	}
	if params.Structure == nil {
		return nil, fmt.Errorf("structure is nil")
	}

	if err := validateFormDefinition(params.FormDefinition); err != nil {
		return nil, fmt.Errorf("invalid form_definition: %w", err)
	}
	if err := validateGraphStructure(params.Structure); err != nil {
		return nil, fmt.Errorf("invalid structure graph: %w", err)
	}
	if err := validateXORRoutingRules(params.Structure); err != nil {
		return nil, fmt.Errorf("invalid structure routing rules: %w", err)
	}
	if err := validateInclusiveRoutingRules(params.Structure); err != nil {
		return nil, fmt.Errorf("invalid structure routing rules: %w", err)
	}

	for _, node := range params.Structure.Nodes {
		if node.Type == domain.NodeTypeEmailService {
			if !s.emailServiceEnabled {
				return nil, fmt.Errorf("email service node is disabled")
			}
			var emailParams domain.EmailNodeParams
			if err := json.Unmarshal(node.Properties, &emailParams); err != nil {
				return nil, fmt.Errorf("fail to unmarshal email service properties: %w", err)
			}
			if err := validateEmailNodeParams(emailParams); err != nil {
				return nil, fmt.Errorf("invalid email service properties: %w", err)
			}
		}
	}

	model, err := paramsToProcessDef(params)
	if err != nil {
		return nil, fmt.Errorf("convert request to model failed: %w", err)
	}
	return model, nil
}

func validateFormDefinition(items []biz.FormDataItem) error {
	if err := biz.ValidateFormDataItems(items, biz.FormValidationModeDefinition); err != nil {
		return err
	}
	return nil
}

func validateEmailNodeParams(params domain.EmailNodeParams) error {
	if params.To == "" {
		return fmt.Errorf("to is required")
	}
	if _, err := mail.ParseAddress(params.To); err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}
	if params.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if params.Body == "" {
		return fmt.Errorf("body is required")
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