package data

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	dataModel "github.com/hobbyGG/Dawnix/internal/workflow/data/model"
	domain "github.com/hobbyGG/Dawnix/internal/workflow/domain"
)

type TaskRepo struct {
	db *Data
}

func NewTaskRepo(db *Data) biz.TaskRepo {
	return &TaskRepo{
		db: db,
	}
}

var _ biz.TaskRepo = (*TaskRepo)(nil)

func (repo *TaskRepo) Create(ctx context.Context, task *domain.ProcessTask) error {
	poTask := processTaskToPO(task)
	if err := repo.db.DB(ctx).WithContext(ctx).Create(poTask).Error; err != nil {
		return err
	}
	task.ID = poTask.ID
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, taskID int64) (*domain.ProcessTask, error) {
	var task dataModel.ProcessTask
	err := r.db.DB(ctx).WithContext(ctx).First(&task, taskID).Error
	return task.ToDomain(), err
}

func (r *TaskRepo) GetDetailView(ctx context.Context, taskID int64) (*domain.TaskView, error) {
	var result domain.TaskView

	err := r.db.DB(ctx).WithContext(ctx).Table("process_tasks as t").
		Select(`
			t.id as id,
			t.node_id as node_name,
			t.status,
			t.assignee,
			t.candidates,
			t.action,
			t.comment,
			t.form_data,  -- 详情页需要这个大 JSON
			t.created_at as create_time,
			t.finished_at,

			d.name as process_title,
			d.code as process_code,

			i.id as instance_id,
			i.submitter_id as submitter_name
		`).
		Joins("LEFT JOIN process_instances i ON t.instance_id = i.id").
		Joins("LEFT JOIN process_definitions d ON i.definition_id = d.id").
		Where("t.id = ?", taskID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *TaskRepo) ListWithFilter(ctx context.Context, params *biz.ListTasksParams) ([]*domain.TaskView, int64, error) {
	var results []*domain.TaskView
	var total int64

	// 1. 基础查询：关联表
	// 注意：这里不需要 Select，Select 留到最后 Fetch 数据时再做，避免 Count 时查询多余字段
	db := r.db.DB(ctx).WithContext(ctx).Table("process_tasks as t").
		Joins("LEFT JOIN process_instances i ON t.instance_id = i.id").
		Joins("LEFT JOIN process_definitions d ON i.definition_id = d.id")

	// 2. 状态过滤
	if params.Status != "" {
		db = db.Where("t.status = ?", params.Status)
	}

	// 3. 身份匹配 (核心修正)
	if params.UserID != "" {
		// 构造 JSON 数组字符串 ["user:123"]
		userJSON := fmt.Sprintf("[%q]", params.UserID)

		// 逻辑：Assignee 是我 OR (Assignee 为空/NULL AND 我在候选人数组里)
		// 注意：candidates 列是 json 类型，需转为 jsonb 后再使用 @>。
		db = db.Where(
			"t.assignee = ? OR ((t.assignee = '' OR t.assignee IS NULL) AND t.candidates::jsonb @> ?::jsonb)",
			params.UserID,
			userJSON,
		)
	}

	// 4. 获取总数 (Count)
	// 注意：Count 必须在 Limit/Offset 之前
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. 执行分页与字段投影
	// 这里的 Select 只影响最终的 Scan，不影响上面的 Count
	selectSQL := `
        t.id,
        t.node_id as task_name,
        t.status,
        d.name as process_title,
        i.submitter_id as submitter_name,
        t.created_at as arrived_at
    `

	offset := (params.Page - 1) * params.Size
	err := db.Select(selectSQL).
		Order("t.created_at DESC").
		Offset(offset).Limit(params.Size).
		Scan(&results).Error

	return results, total, err
}

func (repo *TaskRepo) Update(ctx context.Context, task *domain.ProcessTask) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Save(processTaskToPO(task)).Error; err != nil {
		return err
	}
	return nil
}

func processTaskToPO(src *domain.ProcessTask) *dataModel.ProcessTask {
	if src == nil {
		return nil
	}
	return &dataModel.ProcessTask{
		BaseModel: dataModel.BaseModel{
			ID:        src.ID,
			CreatedAt: src.CreatedAt,
			UpdatedAt: src.UpdatedAt,
			CreatedBy: src.CreatedBy,
			UpdatedBy: src.UpdatedBy,
		},
		InstanceID:  src.InstanceID,
		ExecutionID: src.ExecutionID,
		NodeID:      src.NodeID,
		Type:        src.Type,
		Assignee:    src.Assignee,
		Candidates:  src.Candidates,
		Status:      src.Status,
		Action:      src.Action,
		Comment:     src.Comment,
		FormData:    src.FormData,
	}
}
