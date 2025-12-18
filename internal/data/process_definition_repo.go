package data

import (
	"context"
	"errors"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type ProcessDefinitionRepo struct {
	// gorm连接db
	db *Data
}

func NewProcessDefinitionRepo(db *Data) biz.ProcessDefinitionRepo {
	return &ProcessDefinitionRepo{
		db: db,
	}
}

var _ biz.ProcessDefinitionRepo = (*ProcessDefinitionRepo)(nil)

func (repo *ProcessDefinitionRepo) Create(ctx context.Context, model *model.ProcessDefinition) (int64, error) {
	// 这里实现创建流程模板的数据库操作
	if model == nil {
		return -1, errors.New("repo: creation model cannot be nil")
	}
	if err := repo.db.DB(ctx).WithContext(ctx).Create(model).Error; err != nil {
		return -1, err
	}
	return model.ID, nil
}

func (repo *ProcessDefinitionRepo) List(ctx context.Context, params *biz.ProcessDefinitionListParams) ([]*model.ProcessDefinition, error) {
	// 这里实现获取流程模板列表的数据库操作
	if params == nil {
		return nil, nil
	}

	var pdList []*model.ProcessDefinition
	query := repo.db.DB(ctx).WithContext(ctx).Model(&model.ProcessDefinition{})

	res := query.Offset(params.Page - 1).Limit(params.Size).Find(&pdList)
	if err := res.Error; err != nil {
		return nil, err
	}
	return pdList, nil
}

func (repo *ProcessDefinitionRepo) GetByID(ctx context.Context, id int64) (*model.ProcessDefinition, error) {
	// 这里实现获取流程模板详情的数据库操作
	var pd model.ProcessDefinition
	res := repo.db.DB(ctx).WithContext(ctx).First(&pd, id)
	if err := res.Error; err != nil {
		return nil, err
	}
	return &pd, nil
}

// GetByCode 根据业务code获取流程模板详情，返回最新version
func (repo *ProcessDefinitionRepo) GetByCode(ctx context.Context, code string) (*model.ProcessDefinition, error) {
	var pd model.ProcessDefinition
	res := repo.db.DB(ctx).WithContext(ctx).Where("code = ?", code).Order("version DESC").First(&pd)
	if err := res.Error; err != nil {
		return nil, err
	}
	return &pd, nil
}

func (repo *ProcessDefinitionRepo) DeleteByID(ctx context.Context, id int64) error {
	// 这里实现删除流程模板的数据库操作
	res := repo.db.DB(ctx).WithContext(ctx).Delete(&model.ProcessDefinition{}, id)
	return res.Error
}

func (repo *ProcessDefinitionRepo) Update(ctx context.Context, model *model.ProcessDefinition) error {
	// 这里实现更新流程模板的数据库操作

	return nil
}
