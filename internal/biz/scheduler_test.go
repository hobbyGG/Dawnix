package biz_test

import (
	"context"
	"os"
	"testing"

	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"github.com/hobbyGG/Dawnix/internal/data"
	"github.com/hobbyGG/Dawnix/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type schedulerTestEnv struct {
	db           *gorm.DB
	definition   biz.ProcessDefinitionRepo
	instance     biz.InstanceRepo
	execution    biz.ExecutionRepo
	task         biz.TaskCommandRepo
	scheduler    *biz.Scheduler
	defaultGraph *model.GraphModel
}

func setupSchedulerTestEnv(t *testing.T) *schedulerTestEnv {
	t.Helper()

	dsn := os.Getenv("DAWNIX_TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=root password=123 dbname=root port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true, TablePrefix: "dawnix_"},
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Skipf("skip scheduler tests because db is unavailable: %v", err)
	}

	if err := db.AutoMigrate(
		&model.ProcessDefinition{},
		&model.ProcessInstance{},
		&model.ProcessTask{},
		&model.Execution{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db failed: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	dataObj, _, err := data.NewData(db)
	if err != nil {
		t.Fatalf("new data object failed: %v", err)
	}

	definitionRepo := data.NewProcessDefinitionRepo(dataObj)
	instanceRepo := data.NewInstanceRepo(dataObj)
	executionRepo := data.NewExecutionRepo(dataObj)
	taskRepo := data.NewCommandTaskRepo(dataObj)

	scheduler := biz.NewScheduler(
		data.NewTransactionManager(db),
		definitionRepo,
		instanceRepo,
		executionRepo,
		taskRepo,
		nil,
	)

	return &schedulerTestEnv{
		db:           db,
		definition:   definitionRepo,
		instance:     instanceRepo,
		execution:    executionRepo,
		task:         taskRepo,
		scheduler:    scheduler,
		defaultGraph: util.RandomGraphStartUserEnd(),
	}
}

func truncateSchedulerTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	err := db.Exec("TRUNCATE TABLE process_tasks, dawnix_execution, process_instances, process_definitions RESTART IDENTITY CASCADE").Error
	if err != nil {
		t.Fatalf("truncate tables failed: %v", err)
	}
}

func findNodeIDByType(graph *model.GraphModel, nodeType string) string {
	for _, node := range graph.Nodes {
		if node.Type == nodeType {
			return node.ID
		}
	}
	return ""
}

func TestScheduler_StartProcessInstance(t *testing.T) {
	env := setupSchedulerTestEnv(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		prepare func(t *testing.T, env *schedulerTestEnv) *biz.StartProcessInstanceCmd
		assert  func(t *testing.T, env *schedulerTestEnv, instID int64, err error)
	}{
		{
			name: "success_create_instance_and_execution",
			prepare: func(t *testing.T, env *schedulerTestEnv) *biz.StartProcessInstanceCmd {
				def := util.RandomProcessDefinition(env.defaultGraph)
				_, err := env.definition.Create(ctx, def)
				if err != nil {
					t.Fatalf("create definition failed: %v", err)
				}
				return &biz.StartProcessInstanceCmd{
					ProcessCode: def.Code,
					SubmitterID: util.RandomString("submitter"),
					Variables: map[string]interface{}{
						"amount": 88,
					},
				}
			},
			assert: func(t *testing.T, env *schedulerTestEnv, instID int64, err error) {
				if err != nil {
					t.Fatalf("start process instance failed: %v", err)
				}
				if instID <= 0 {
					t.Fatalf("invalid inst id: %d", instID)
				}

				inst, getErr := env.instance.GetByID(ctx, instID)
				if getErr != nil {
					t.Fatalf("get instance failed: %v", getErr)
				}
				if inst.Status != model.InstanceStatusPending {
					t.Fatalf("expected instance status %s, got %s", model.InstanceStatusPending, inst.Status)
				}

				var exec model.Execution
				if queryErr := env.db.WithContext(ctx).Where("inst_id = ?", instID).First(&exec).Error; queryErr != nil {
					t.Fatalf("query execution failed: %v", queryErr)
				}
				startNodeID := findNodeIDByType(env.defaultGraph, model.NodeTypeStart)
				if exec.NodeID != startNodeID {
					t.Fatalf("expected execution node %s, got %s", startNodeID, exec.NodeID)
				}
			},
		},
		{
			name: "fail_when_definition_not_found",
			prepare: func(t *testing.T, env *schedulerTestEnv) *biz.StartProcessInstanceCmd {
				return &biz.StartProcessInstanceCmd{
					ProcessCode: util.RandomString("missing_code"),
					SubmitterID: util.RandomString("submitter"),
					Variables:   map[string]interface{}{"k": "v"},
				}
			},
			assert: func(t *testing.T, env *schedulerTestEnv, instID int64, err error) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateSchedulerTables(t, env.db)
			cmd := tt.prepare(t, env)
			instID, err := env.scheduler.StartProcessInstance(ctx, cmd)
			tt.assert(t, env, instID, err)
		})
	}
}

func TestScheduler_CompleteTask(t *testing.T) {
	env := setupSchedulerTestEnv(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		prepare func(t *testing.T, env *schedulerTestEnv) *model.ProcessTask
		assert  func(t *testing.T, env *schedulerTestEnv, task *model.ProcessTask, err error)
	}{
		{
			name: "success_complete_task_and_finish_instance",
			prepare: func(t *testing.T, env *schedulerTestEnv) *model.ProcessTask {
				def := util.RandomProcessDefinition(env.defaultGraph)
				_, err := env.definition.Create(ctx, def)
				if err != nil {
					t.Fatalf("create definition failed: %v", err)
				}

				inst := util.RandomProcessInstance(def.ID, def.Code, env.defaultGraph)
				instID, err := env.instance.Create(ctx, inst)
				if err != nil {
					t.Fatalf("create instance failed: %v", err)
				}

				userNodeID := findNodeIDByType(env.defaultGraph, model.NodeTypeUserTask)
				exec := &model.Execution{InstID: instID, NodeID: userNodeID, IsActive: true}
				if err = env.execution.Create(ctx, exec); err != nil {
					t.Fatalf("create execution failed: %v", err)
				}

				task := util.RandomProcessTask(instID, exec.ID, userNodeID)
				if err = env.task.Create(ctx, task); err != nil {
					t.Fatalf("create task failed: %v", err)
				}
				return task
			},
			assert: func(t *testing.T, env *schedulerTestEnv, task *model.ProcessTask, err error) {
				if err != nil {
					t.Fatalf("complete task failed: %v", err)
				}

				storedTask, taskErr := env.task.GetByID(ctx, task.ID)
				if taskErr != nil {
					t.Fatalf("get task failed: %v", taskErr)
				}
				if storedTask.Status != model.TaskStatusApproved {
					t.Fatalf("expected task status %s, got %s", model.TaskStatusApproved, storedTask.Status)
				}

				inst, instErr := env.instance.GetByID(ctx, task.InstanceID)
				if instErr != nil {
					t.Fatalf("get instance failed: %v", instErr)
				}
				if inst.Status != model.InstanceStatusApproved {
					t.Fatalf("expected instance status %s, got %s", model.InstanceStatusApproved, inst.Status)
				}

				_, execErr := env.execution.GetByID(ctx, task.ExecutionID)
				if execErr == nil {
					t.Fatalf("expected execution deleted, but still exists")
				}
			},
		},
		{
			name: "fail_when_execution_not_found",
			prepare: func(t *testing.T, env *schedulerTestEnv) *model.ProcessTask {
				def := util.RandomProcessDefinition(env.defaultGraph)
				_, err := env.definition.Create(ctx, def)
				if err != nil {
					t.Fatalf("create definition failed: %v", err)
				}

				inst := util.RandomProcessInstance(def.ID, def.Code, env.defaultGraph)
				instID, err := env.instance.Create(ctx, inst)
				if err != nil {
					t.Fatalf("create instance failed: %v", err)
				}

				userNodeID := findNodeIDByType(env.defaultGraph, model.NodeTypeUserTask)
				task := util.RandomProcessTask(instID, 99999999, userNodeID)
				if err = env.task.Create(ctx, task); err != nil {
					t.Fatalf("create task failed: %v", err)
				}
				return task
			},
			assert: func(t *testing.T, env *schedulerTestEnv, task *model.ProcessTask, err error) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				storedTask, taskErr := env.task.GetByID(ctx, task.ID)
				if taskErr != nil {
					t.Fatalf("get task failed: %v", taskErr)
				}
				if storedTask.Status != model.TaskStatusPending {
					t.Fatalf("expected task status %s, got %s", model.TaskStatusPending, storedTask.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateSchedulerTables(t, env.db)
			task := tt.prepare(t, env)
			err := env.scheduler.CompleteTask(ctx, task)
			tt.assert(t, env, task, err)
		})
	}
}
