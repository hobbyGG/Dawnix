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

func (repo *ExecutionRepo) GetByID(ctx context.Context, id int64) (*model.Execution, error) {
	var exec model.Execution
	if err := repo.db.DB(ctx).WithContext(ctx).First(&exec, id).Error; err != nil {
		return nil, err
	}
	return &exec, nil
}

func (repo *ExecutionRepo) Update(ctx context.Context, exec *model.Execution) error {
	return repo.db.DB(ctx).WithContext(ctx).Save(exec).Error
}

func (repo *ExecutionRepo) DeleteByID(ctx context.Context, id int64) error {
	return repo.db.DB(ctx).WithContext(ctx).Delete(&model.Execution{}, id).Error
}
