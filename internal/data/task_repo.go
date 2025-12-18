package data

import (
	"context"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
)

type TaskRepo struct {
	db *Data
}

func NewCommandTaskRepo(db *Data) biz.TaskCommandRepo {
	return &TaskRepo{
		db: db,
	}
}

func NewQueryTaskRepo(db *Data) biz.TaskQueryRepo {
	return &TaskRepo{
		db: db,
	}
}

var _ biz.TaskCommandRepo = (*TaskRepo)(nil)
var _ biz.TaskQueryRepo = (*TaskRepo)(nil)

func (repo *TaskRepo) Create(ctx context.Context, task *model.ProcessTask) error {
	if err := repo.db.DB(ctx).WithContext(ctx).Create(task).Error; err != nil {
		return err
	}
	return nil
}

func (r *TaskRepo) GetDetailView(ctx context.Context, taskID int64) (*model.TaskView, error) {
	var result model.TaskView

	// 这里的 SQL 逻辑是 CQRS 的精髓：Data 层负责解决数据的复杂性
	err := r.db.DB(ctx).WithContext(ctx).Table("process_tasks as t").
		Select(`
			t.id as task_id,
			t.node_id as node_name,
			t.status,
			t.assignee,
			t.candidates,
			t.action,
			t.comment,
			t.variables,  -- 详情页需要这个大 JSON
			t.created_at as create_time,
			t.finished_at,

			d.name as process_title,
			d.code as process_code,

			i.id as instance_id,
			i.submitter_id as submitter_name
		`).
		// 连表 Instance
		Joins("LEFT JOIN process_instances i ON t.instance_id = i.id").
		// 连表 Definition
		Joins("LEFT JOIN process_definitions d ON i.definition_id = d.id").
		Where("t.id = ?", taskID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (repo *TaskRepo) ListPending(ctx context.Context, params *biz.ListTasksParams) ([]*model.TaskView, error) {
	var tasks []*model.TaskView
	query := repo.db.DB(ctx).WithContext(ctx).Table("process_tasks AS t").Select("t.id AS task_id, t.node_id AS task_name, d.title AS process_title, u.name AS submitter_name, t.created_at AS arrived_at").
		Joins("JOIN process_instances AS i ON t.instance_id = i.id").
		Joins("JOIN process_definitions AS d ON i.definition_id = d.id").
		Joins("JOIN users AS u ON i.submitter_id = u.id").
		Where("t.status = ?", model.TaskStatusPending).
		Offset((params.Page - 1) * params.Size).
		Limit(params.Size)
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}
