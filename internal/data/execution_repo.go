package data

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type ExecutionRepo struct {
	db *Data
}

func NewExecutionRepo(db *Data) biz.ExecutionRepo {
	return &ExecutionRepo{db: db}
}

var _ biz.ExecutionRepo = (*ExecutionRepo)(nil)

func (repo *ExecutionRepo) Create(ctx context.Context, exec *model.Execution) error {
	return repo.db.DB(ctx).WithContext(ctx).Create(exec).Error
}

func (repo *ExecutionRepo) CreateBatch(ctx context.Context, execs []model.Execution) error {
	return repo.db.DB(ctx).WithContext(ctx).CreateInBatches(execs, len(execs)).Error
}

func (repo *ExecutionRepo) GetByID(ctx context.Context, id int64) (*model.Execution, error) {
	var exec model.Execution
	if err := repo.db.DB(ctx).WithContext(ctx).First(&exec, id).Error; err != nil {
		return nil, err
	}
	return &exec, nil
}

func (repo *ExecutionRepo) GetActiveNums(ctx context.Context, instID int64) (int, error) {
	var execs []model.Execution
	err := repo.db.sqlDB.WithContext(ctx).Model(&model.Execution{}).Where("inst_id = ? and is_active = ?", instID, true).Find(&execs).Error
	if err != nil {
		return 0, err
	}
	return len(execs), nil
}

func (repo *ExecutionRepo) GetActiveNumsByParentID(ctx context.Context, parentID int64) (int, error) {
	var execs []model.Execution
	err := repo.db.DB(ctx).WithContext(ctx).Model(&model.Execution{}).Where("parent_id = ? and is_active = ?", parentID, true).Find(&execs).Error
	if err != nil {
		return 0, err
	}
	return len(execs), nil
}

func (repo *ExecutionRepo) Update(ctx context.Context, exec *model.Execution) error {
	return repo.db.DB(ctx).WithContext(ctx).Save(exec).Error
}

func (repo *ExecutionRepo) UpdateNodeID(ctx context.Context, execID int64, nodeID string) error {
	return nil
}

func (repo *ExecutionRepo) DeleteByID(ctx context.Context, id int64) error {
	return repo.db.DB(ctx).WithContext(ctx).Delete(&model.Execution{}, id).Error
}
