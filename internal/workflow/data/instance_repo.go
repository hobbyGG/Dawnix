package data

import (
	"context"
	"time"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	dataModel "github.com/hobbyGG/Dawnix/internal/workflow/data/model"
	domain "github.com/hobbyGG/Dawnix/internal/workflow/domain"
	"gorm.io/gorm"
)

type InstanceRepo struct {
	// gorm连接db
	db *Data
}

// 构造方法返回biz层
func NewInstanceRepo(db *Data) biz.InstanceRepo {
	return &InstanceRepo{
		db: db,
	}
}

func (repo *InstanceRepo) Create(ctx context.Context, inst *domain.ProcessInstance) (int64, error) {
	poInst := processInstanceToModel(inst)
	if err := repo.db.DB(ctx).WithContext(ctx).Create(poInst).Error; err != nil {
		return 0, err
	}
	inst.ID = poInst.ID
	return poInst.ID, nil
}

func (repo *InstanceRepo) List(ctx context.Context, params *biz.ListInstancesParams) ([]domain.ProcessInstance, int64, error) {
	var instances []dataModel.ProcessInstance
	query := repo.db.DB(ctx).WithContext(ctx).
		Table("process_instances as i").
		Select("i.*")
	query = applyInstanceListFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.Size
	if err := query.Order("i.created_at DESC").Offset(offset).Limit(params.Size).Find(&instances).Error; err != nil {
		return nil, 0, err
	}

	result := make([]domain.ProcessInstance, 0, len(instances))
	for i := range instances {
		if item := instances[i].ToDomain(); item != nil {
			result = append(result, *item)
		}
	}
	return result, total, nil
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
	if err := repo.db.DB(ctx).WithContext(ctx).Save(processInstanceToModel(inst)).Error; err != nil {
		return err
	}
	return nil
}

func (repo *InstanceRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if isFinishedInstanceStatus(status) {
		updates["finished_at"] = time.Now()
	}
	return repo.db.DB(ctx).WithContext(ctx).Model(&dataModel.ProcessInstance{}).Where("id = ?", id).Updates(updates).Error
}

func applyInstanceListFilters(query *gorm.DB, params *biz.ListInstancesParams) *gorm.DB {
	if params == nil {
		return query
	}

	if params.ProcessCode != "" {
		query = query.Where("i.process_code = ?", params.ProcessCode)
	}
	if params.SubmitterID != "" {
		query = query.Where("i.submitter_id = ?", params.SubmitterID)
	}
	if !params.CreatedAtFrom.IsZero() {
		query = query.Where("i.created_at >= ?", params.CreatedAtFrom)
	}
	if !params.CreatedAtTo.IsZero() {
		query = query.Where("i.created_at <= ?", params.CreatedAtTo)
	}

	switch params.State {
	case biz.ListInstancesStateFinished:
		query = query.Where("i.status IN ?", []string{
			domain.InstanceStatusApproved,
			domain.InstanceStatusRejected,
			domain.InstanceStatusCanceled,
		})
	case biz.ListInstancesStateUnfinished:
		query = query.Where("i.status IN ?", []string{
			domain.InstanceStatusPending,
			domain.InstanceStatusSuspended,
		})
	}
	return query
}

func isFinishedInstanceStatus(status string) bool {
	switch status {
	case domain.InstanceStatusApproved, domain.InstanceStatusRejected, domain.InstanceStatusCanceled:
		return true
	default:
		return false
	}
}

func processInstanceToModel(src *domain.ProcessInstance) *dataModel.ProcessInstance {
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
