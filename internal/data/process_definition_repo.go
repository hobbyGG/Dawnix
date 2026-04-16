package data

import (
	"context"
	"errors"

	"github.com/hobbyGG/Dawnix/internal/biz"
	dataModel "github.com/hobbyGG/Dawnix/internal/data/model"
	domain "github.com/hobbyGG/Dawnix/internal/domain"
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

func (repo *ProcessDefinitionRepo) Create(ctx context.Context, pd *domain.ProcessDefinition) (int64, error) {
	// 这里实现创建流程模板的数据库操作
	if pd == nil {
		return -1, errors.New("repo: creation model cannot be nil")
	}
	poPD := processDefinitionToPO(pd)
	if err := repo.db.DB(ctx).WithContext(ctx).Create(poPD).Error; err != nil {
		return -1, err
	}
	pd.ID = poPD.ID
	return poPD.ID, nil
}

func (repo *ProcessDefinitionRepo) List(ctx context.Context, params *biz.ProcessDefinitionListParams) ([]domain.ProcessDefinition, error) {
	// 这里实现获取流程模板列表的数据库操作
	if params == nil {
		return nil, nil
	}

	var pdList []dataModel.ProcessDefinition
	query := repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.ProcessDefinition{})

	res := query.Offset(params.Page - 1).Limit(params.Size).Find(&pdList)
	if err := res.Error; err != nil {
		return nil, err
	}
	result := make([]domain.ProcessDefinition, 0, len(pdList))
	for i := range pdList {
		if item := pdList[i].ToDomain(); item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (repo *ProcessDefinitionRepo) GetByID(ctx context.Context, id int64) (*domain.ProcessDefinition, error) {
	// 这里实现获取流程模板详情的数据库操作
	var pd dataModel.ProcessDefinition
	res := repo.db.DB(ctx).WithContext(ctx).First(&pd, id)
	if err := res.Error; err != nil {
		return nil, err
	}
	return pd.ToDomain(), nil
}

// GetByCode 根据业务code获取流程模板详情，返回最新version
func (repo *ProcessDefinitionRepo) GetByCode(ctx context.Context, code string) (*domain.ProcessDefinition, error) {
	var pd dataModel.ProcessDefinition
	res := repo.db.DB(ctx).WithContext(ctx).Where("code = ?", code).Order("version DESC").First(&pd)
	if err := res.Error; err != nil {
		return nil, err
	}
	return pd.ToDomain(), nil
}

func (repo *ProcessDefinitionRepo) DeleteByID(ctx context.Context, id int64) error {
	// 这里实现删除流程模板的数据库操作
	res := repo.db.DB(ctx).WithContext(ctx).Delete(&dataModel.ProcessDefinition{}, id)
	return res.Error
}

func (repo *ProcessDefinitionRepo) Update(ctx context.Context, pd *domain.ProcessDefinition) error {
	// 这里实现更新流程模板的数据库操作
	return repo.db.DB(ctx).WithContext(ctx).Save(processDefinitionToPO(pd)).Error
}

func processDefinitionToPO(src *domain.ProcessDefinition) *dataModel.ProcessDefinition {
	if src == nil {
		return nil
	}
	return &dataModel.ProcessDefinition{
		BaseModel: dataModel.BaseModel{
			ID:        src.ID,
			CreatedAt: src.CreatedAt,
			UpdatedAt: src.UpdatedAt,
			CreatedBy: src.CreatedBy,
			UpdatedBy: src.UpdatedBy,
		},
		Code:      src.Code,
		Version:   src.Version,
		Name:      src.Name,
		Structure: src.Structure,
		Config:    src.Config,
		IsActive:  src.IsActive,
	}
}
