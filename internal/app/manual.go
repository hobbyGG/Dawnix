package app

import (
	"fmt"
	"time"

	"github.com/hobbyGG/Dawnix/internal/api"
	"github.com/hobbyGG/Dawnix/internal/api/handler"
	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/biz/model"
	"github.com/hobbyGG/Dawnix/internal/data"
	"github.com/hobbyGG/Dawnix/internal/service"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewAppManual(logger *zap.Logger) (*App, error) {
	// 数据库初始化化
	db, err := gorm.Open(postgres.Open("host=localhost user=root password=123 dbname=root port=5432 sslmode=disable TimeZone=Asia/Shanghai"), &gorm.Config{
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true, TablePrefix: "dawnix_"},
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to initialize postgresSQL: %w", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	db.AutoMigrate(
		&model.ProcessDefinition{},
		&model.ProcessTask{},
		&model.ProcessInstance{},
	)

	dataObj, cleanup, err := data.NewData(db)
	if err != nil {
		return nil, fmt.Errorf("fail to initialize data layer: %w", err)
	}

	helloRepo := data.NewHelloRepo(logger)
	processDefinitionRepo := data.NewProcessDefinitionRepo(dataObj)
	instanceRepo := data.NewInstanceRepo(dataObj)
	TaskCmdRepo := data.NewCommandTaskRepo(dataObj)
	TaskQueryRepo := data.NewQueryTaskRepo(dataObj)

	txManager := data.NewTransactionManager(db)
	scheduler := biz.NewScheduler(txManager, processDefinitionRepo, instanceRepo, TaskCmdRepo)

	helloSvc := service.NewHelloService(helloRepo, logger)
	helloHandler := handler.NewHelloHandler(helloSvc, logger)

	processDefinisionSvc := service.NewProcessDefinitionService(processDefinitionRepo, logger)
	processDefinitionHandler := handler.NewProcessDefinitionHandler(processDefinisionSvc, logger)

	instanceSvc := service.NewInstanceService(instanceRepo, scheduler, logger)
	instanceHandler := handler.NewInstanceHandler(instanceSvc, logger)

	taskSvc := service.NewTaskService(TaskCmdRepo, TaskQueryRepo, logger)
	taskHandler := handler.NewTaskHandler(taskSvc, logger)

	r := api.NewRouter(helloHandler, processDefinitionHandler, instanceHandler, taskHandler)

	app := &App{
		Server:  r,
		Cleanup: cleanup,
	}
	return app, nil
}
