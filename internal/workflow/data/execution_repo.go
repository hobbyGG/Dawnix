package data

import (
	"context"

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
		return err
	}
	exec.ID = poExec.ID
	return nil
}

func (repo *ExecutionRepo) CreateBatch(ctx context.Context, execs []domain.Execution) error {
	for i := range execs {
		poExec := executionToPO(&execs[i])
		if err := repo.db.DB(ctx).WithContext(ctx).Create(poExec).Error; err != nil {
			return err
		}
		execs[i].ID = poExec.ID
	}
	return nil
}

func (repo *ExecutionRepo) GetByID(ctx context.Context, id int64) (*domain.Execution, error) {
	var exec dataModel.Execution
	if err := repo.db.DB(ctx).WithContext(ctx).First(&exec, id).Error; err != nil {
		return nil, err
	}
	return exec.ToDomain(), nil
}

func (repo *ExecutionRepo) GetActiveNums(ctx context.Context, instID int64) (int, error) {
	var execs []dataModel.Execution
	err := repo.db.sqlDB.WithContext(ctx).Model(&dataModel.Execution{}).Where("inst_id = ? and is_active = ?", instID, true).Find(&execs).Error
	if err != nil {
		return 0, err
	}
	return len(execs), nil
}

func (repo *ExecutionRepo) GetActiveNumsByParentID(ctx context.Context, parentID int64) (int, error) {
	var execs []dataModel.Execution
	err := repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.Execution{}).Where("parent_id = ? and is_active = ?", parentID, true).Find(&execs).Error
	if err != nil {
		return 0, err
	}
	return len(execs), nil
}

func (repo *ExecutionRepo) Update(ctx context.Context, exec *domain.Execution) error {
	return repo.db.DB(ctx).WithContext(ctx).Save(executionToPO(exec)).Error
}

func (repo *ExecutionRepo) UpdateNodeID(ctx context.Context, execID int64, nodeID string) error {
	return nil
}

func (repo *ExecutionRepo) DeleteByID(ctx context.Context, id int64) error {
	return repo.db.DB(ctx).WithContext(ctx).Delete(&dataModel.Execution{}, id).Error
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
