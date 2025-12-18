package service

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
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

func (s *ProcessDefinitionService) ListProcessDefinitions(ctx context.Context, params *biz.ProcessDefinitionListParams) ([]*model.ProcessDefinition, error) {
	// 这里实现获取流程模板列表的业务逻辑
	// 业务校验：分页参数是否合法
	pdList, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fail to get processDefinition list: %w", err)
	}
	return pdList, nil
}

func (s *ProcessDefinitionService) GetProcessDefinitionDetail(ctx context.Context, id int64) (*model.ProcessDefinition, error) {
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

func (s *ProcessDefinitionService) UpdateProcessDefinition(ctx context.Context, model *model.ProcessDefinition) error {
	// 这里实现更新流程模板的业务逻辑
	// 业务校验：流程模板是否存在，唯一字段是否冲突等
	err := s.repo.Update(ctx, model)
	if err != nil {
		return fmt.Errorf("fail to update processDefinition: %w", err)
	}
	return nil
}
