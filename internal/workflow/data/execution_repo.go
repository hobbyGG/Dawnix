package data

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	dataModel "github.com/hobbyGG/Dawnix/internal/workflow/data/model"
	domain "github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type ExecutionRepo struct {
	db *Data
}

func NewExecutionRepo(db *Data) biz.ExecutionRepo {
	return &ExecutionRepo{db: db}
}

var _ biz.ExecutionRepo = (*ExecutionRepo)(nil)

func (repo *ExecutionRepo) Create(ctx context.Context, exec *domain.Execution) error {
	poExec := executionToPO(exec)
	if err := repo.db.DB(ctx).WithContext(ctx).Create(poExec).Error; err != nil {
		return fmt.Errorf("create execution: %w", err)
	}
	exec.ID = poExec.ID
	return nil
}

func (repo *ExecutionRepo) CreateBatch(ctx context.Context, execs []domain.Execution) error {
	for i := range execs {
		poExec := executionToPO(&execs[i])
		if err := repo.db.DB(ctx).WithContext(ctx).Create(poExec).Error; err != nil {
			return fmt.Errorf("create execution batch item: %w", err)
		}
		execs[i].ID = poExec.ID
	}
	return nil
}

func (repo *ExecutionRepo) GetByID(ctx context.Context, id int64) (*domain.Execution, error) {
	var exec dataModel.Execution
	if err := repo.db.DB(ctx).WithContext(ctx).First(&exec, id).Error; err != nil {
		return nil, fmt.Errorf("get execution by id %d: %w", id, err)
	}
	return exec.ToDomain(), nil
}

func (repo *ExecutionRepo) GetActiveNums(ctx context.Context, instID int64) (int, error) {
	var total int64
	err := repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.Execution{}).Where("inst_id = ? and is_active = ?", instID, true).Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("count active executions by inst_id %d: %w", instID, err)
	}
	return int(total), nil
}

func (repo *ExecutionRepo) GetActiveNumsByParentID(ctx context.Context, parentID int64) (int, error) {
	var total int64
	err := repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.Execution{}).Where("parent_id = ? and is_active = ?", parentID, true).Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("count active executions by parent_id %d: %w", parentID, err)
	}
	return int(total), nil
}

func (repo *ExecutionRepo) Update(ctx context.Context, exec *domain.Execution) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Save(executionToPO(exec)).Error; err != nil {
		return fmt.Errorf("update execution id %d: %w", exec.ID, err)
	}
	return nil
}

func (repo *ExecutionRepo) UpdateNodeID(ctx context.Context, execID int64, nodeID string) error {
	return nil
}

func (repo *ExecutionRepo) DeleteByID(ctx context.Context, id int64) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Delete(&dataModel.Execution{}, id).Error; err != nil {
		return fmt.Errorf("delete execution id %d: %w", id, err)
	}
	return nil
}

func executionToPO(src *domain.Execution) *dataModel.Execution {
	if src == nil {
		return nil
	}
	return &dataModel.Execution{
		BaseModel: dataModel.BaseModel{
			ID:        src.ID,
			CreatedAt: src.CreatedAt,
			UpdatedAt: src.UpdatedAt,
			CreatedBy: src.CreatedBy,
			UpdatedBy: src.UpdatedBy,
		},
		InstID:   src.InstID,
		ParentID: src.ParentID,
		NodeID:   src.NodeID,
		IsActive: src.IsActive,
	}
}
