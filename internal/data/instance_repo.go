package data

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	dataModel "github.com/hobbyGG/Dawnix/internal/data/model"
	domain "github.com/hobbyGG/Dawnix/internal/domain"
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

func (repo *InstanceRepo) Create(ctx context.Context, inst *domain.ProcessInstance) (int64, error) {
	poInst := processInstanceToPO(inst)
	if err := repo.db.DB(ctx).WithContext(ctx).Create(poInst).Error; err != nil {
		return 0, err
	}
	inst.ID = poInst.ID
	return poInst.ID, nil
}

func (repo *InstanceRepo) List(ctx context.Context, params *biz.ListInstancesParams) ([]domain.ProcessInstance, error) {
	var instances []dataModel.ProcessInstance
	query := repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.ProcessInstance{})
	if err := query.Offset(params.Page - 1).Limit(params.Size).Find(&instances).Error; err != nil {
		return nil, err
	}

	result := make([]domain.ProcessInstance, 0, len(instances))
	for i := range instances {
		if item := instances[i].ToDomain(); item != nil {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (repo *InstanceRepo) GetByID(ctx context.Context, id int64) (*domain.ProcessInstance, error) {
	var instance dataModel.ProcessInstance
	if err := repo.db.DB(ctx).WithContext(ctx).First(&instance, id).Error; err != nil {
		return nil, err
	}
	return instance.ToDomain(), nil
}

func (repo *InstanceRepo) GetWithExecutionsByID(ctx context.Context, id int64) (*domain.ProcessInstance, []domain.Execution, error) {
	var instance dataModel.ProcessInstance
	if err := repo.db.DB(ctx).WithContext(ctx).Where("id = ?", id).First(&instance).Error; err != nil {
		return nil, nil, err
	}

	var executions []dataModel.Execution
	if err := repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.Execution{}).Where("inst_id = ? and is_active = ?", id, true).Find(&executions).Error; err != nil {
		return nil, nil, err
	}

	result := make([]domain.Execution, 0, len(executions))
	for i := range executions {
		if item := executions[i].ToDomain(); item != nil {
			result = append(result, *item)
		}
	}
	return instance.ToDomain(), result, nil
}

func (repo *InstanceRepo) Delete(ctx context.Context, id int64) error {
	return repo.db.DB(ctx).WithContext(ctx).Delete(&dataModel.ProcessInstance{}, id).Error
}

func (repo *InstanceRepo) Update(ctx context.Context, inst *domain.ProcessInstance) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Save(processInstanceToPO(inst)).Error; err != nil {
		return err
	}
	return nil
}

func (repo *InstanceRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.ProcessInstance{}).Where("id = ?", id).Update("status", status).Error
}

func processInstanceToPO(src *domain.ProcessInstance) *dataModel.ProcessInstance {
	if src == nil {
		return nil
	}
	return &dataModel.ProcessInstance{
		BaseModel: dataModel.BaseModel{
			ID:        src.ID,
			CreatedAt: src.CreatedAt,
			UpdatedAt: src.UpdatedAt,
			CreatedBy: src.CreatedBy,
			UpdatedBy: src.UpdatedBy,
		},
		DefinitionID:      src.DefinitionID,
		ProcessCode:       src.ProcessCode,
		SnapshotStructure: src.SnapshotStructure,
		ParentID:          src.ParentID,
		ParentNodeID:      src.ParentNodeID,
		FormData:          src.FormData,
		Status:            src.Status,
		SubmitterID:       src.SubmitterID,
		FinishedAt:        src.FinishedAt,
	}
}
