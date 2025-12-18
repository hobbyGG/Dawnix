package service

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"go.uber.org/zap"
)

type InstanceService struct {
	instanceRepo biz.InstanceRepo
	scheduler    biz.InstanceScheduler
	logger       *zap.Logger
}

func NewInstanceService(instanceRepo biz.InstanceRepo, scheduler biz.InstanceScheduler, logger *zap.Logger) *InstanceService {
	return &InstanceService{instanceRepo: instanceRepo, scheduler: scheduler, logger: logger}
}

func (s *InstanceService) CreateInstance(ctx context.Context, cmd biz.StartProcessInstanceCmd) (int64, error) {
	// 业务校验
	// 对cmd的一些字段进行业务上的校验

	// 启动一个流程实例
	id, err := s.scheduler.StartProcessInstance(ctx, cmd)
	if err != nil {
		return 0, fmt.Errorf("scheduler start process instance failed: %w", err)
	}
	return id, nil
}

func (s *InstanceService) ListInstances(ctx context.Context, params *biz.ListInstancesParams) ([]model.ProcessInstance, error) {
	instances, err := s.instanceRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list instances failed: %w", err)
	}
	return instances, nil
}

func (s *InstanceService) GetInstanceDetail(ctx context.Context, id int64) (*model.ProcessInstance, error) {
	instance, err := s.instanceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get instance by id failed: %w", err)
	}
	return instance, nil
}

func (s *InstanceService) DeleteInstance(ctx context.Context, id int64) error {
	return nil
}
