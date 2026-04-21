package data

import (
	"context"
	"fmt"

	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	dataModel "github.com/hobbyGG/Dawnix/internal/workflow/data/model"
	domain "github.com/hobbyGG/Dawnix/internal/workflow/domain"
	"github.com/lib/pq"
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
		return fmt.Errorf("create process task: %w", err)
	}
	task.ID = poTask.ID
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, taskID int64) (*domain.ProcessTask, error) {
	var task dataModel.ProcessTask
	err := r.db.DB(ctx).WithContext(ctx).First(&task, taskID).Error
	if err != nil {
		return nil, fmt.Errorf("get process task by id %d: %w", taskID, err)
	}
	return task.ToDomain(), nil
}

func (r *TaskRepo) GetDetailView(ctx context.Context, taskID int64) (*domain.TaskDetailView, error) {
	var task dataModel.ProcessTask
	if err := r.db.DB(ctx).WithContext(ctx).First(&task, taskID).Error; err != nil {
		return nil, fmt.Errorf("get process task for detail by id %d: %w", taskID, err)
	}

	type taskDetailJoin struct {
		ProcessTitle string
		SubmitterID  string
	}
	var join taskDetailJoin

	err := r.db.DB(ctx).WithContext(ctx).Table("process_tasks as t").
		Select(`
			d.name as process_title,
			i.submitter_id as submitter_id
		`).
		Joins("LEFT JOIN process_instances i ON t.instance_id = i.id").
		Joins("LEFT JOIN process_definitions d ON i.definition_id = d.id").
		Where("t.id = ?", taskID).
		Scan(&join).Error

	if err != nil {
		return nil, fmt.Errorf("get task detail view by id %d: %w", taskID, err)
	}
	return &domain.TaskDetailView{
		ID:           task.ID,
		InstanceID:   task.InstanceID,
		ExecutionID:  task.ExecutionID,
		NodeID:       task.NodeID,
		Type:         task.Type,
		Assignee:     task.Assignee,
		Candidates:   []string(task.Candidates),
		Status:       task.Status,
		Action:       task.Action,
		Comment:      task.Comment,
		FormData:     task.FormData,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		CreatedBy:    task.CreatedBy,
		UpdatedBy:    task.UpdatedBy,
		ProcessTitle: join.ProcessTitle,
		SubmitterID:  join.SubmitterID,
	}, nil
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
		// 逻辑：Assignee 是我 OR (Assignee 为空/NULL AND 我在候选人数组里)
		db = db.Where(
			"t.assignee = ? OR ((t.assignee = '' OR t.assignee IS NULL) AND t.candidates @> ?)",
			params.UserID,
			pq.StringArray{params.UserID},
		)
	}

	// 4. 获取总数 (Count)
	// 注意：Count 必须在 Limit/Offset 之前
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count task views: %w", err)
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

	if err != nil {
		return nil, 0, fmt.Errorf("list task views: %w", err)
	}
	return results, total, nil
}

func (repo *TaskRepo) Update(ctx context.Context, task *domain.ProcessTask) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Save(processTaskToPO(task)).Error; err != nil {
		return fmt.Errorf("update process task id %d: %w", task.ID, err)
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
		Candidates:  pq.StringArray(src.Candidates),
		Status:      src.Status,
		Action:      src.Action,
		Comment:     src.Comment,
		FormData:    src.FormData,
	}
}
