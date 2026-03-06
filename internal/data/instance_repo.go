package data

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type InstanceRepo struct {
	// gorm连接db
	db *Data
}

func NewInstanceRepo(db *Data) biz.InstanceRepo {
	return &InstanceRepo{
		db: db,
	}
}

func (repo *InstanceRepo) Create(ctx context.Context, model *model.ProcessInstance) (int64, error) {
	if err := repo.db.DB(ctx).WithContext(ctx).Create(model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}
func (repo *InstanceRepo) List(ctx context.Context, params *biz.ListInstancesParams) ([]model.ProcessInstance, error) {
	var instances []model.ProcessInstance
	query := repo.db.DB(ctx).WithContext(ctx).Model(&model.ProcessInstance{})
	// 可以根据params添加过滤条件，例如分页等
	if err := query.Offset(params.Page - 1).Limit(params.Size).Find(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}
func (repo *InstanceRepo) GetByID(ctx context.Context, id int64) (*model.ProcessInstance, error) {
	query := repo.db.DB(ctx).WithContext(ctx).Model(&model.ProcessInstance{})
	var instance model.ProcessInstance
	if err := query.Where("id = ?", id).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (repo *InstanceRepo) GetWithExecutionsByID(ctx context.Context, id int64) (*model.ProcessInstance, []model.Execution, error) {
	query := repo.db.DB(ctx).WithContext(ctx).Model(&model.ProcessInstance{})
	var instance model.ProcessInstance
	query = query.Where("id = ?", id).First(&instance)
	if query.Error != nil {
		return nil, nil, query.Error
	}

	var executions []model.Execution
	if err := repo.db.DB(ctx).WithContext(ctx).Model(&model.Execution{}).Where("inst_id = ? and is_active = ?", id, true).Find(&executions).Error; err != nil {
		return nil, nil, err
	}

	return &instance, executions, nil
}

func (repo *InstanceRepo) Delete(ctx context.Context, id int64) error {
	res := repo.db.DB(ctx).WithContext(ctx).Delete(&model.ProcessInstance{}, id)
	return res.Error
}

func (repo *InstanceRepo) Update(ctx context.Context, model *model.ProcessInstance) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	return nil
}

func (repo *InstanceRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Model(&model.ProcessInstance{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}
	return nil
}
