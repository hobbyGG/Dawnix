package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/api"
	"github.com/hobbyGG/Dawnix/internal/biz"
	"github.com/hobbyGG/Dawnix/internal/data"
	dataModel "github.com/hobbyGG/Dawnix/internal/data/model"
	"github.com/hobbyGG/Dawnix/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type App struct {
	Server  *gin.Engine
	Cleanup func()
}

func (a *App) Run() error {
	// 默认8080端口
	return a.Server.Run(":8080")
}

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
		&dataModel.ProcessDefinition{},
		&dataModel.ProcessTask{},
		&dataModel.ProcessInstance{},
		&dataModel.Execution{},
	)

	// rdb 初始化
	redOpts := &redis.Options{
		Addr: "127.0.0.1:16379",
	}
	rdb := redis.NewClient(redOpts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("fail to initialize redis: %w", err)
	}

	dataObj, cleanup, err := data.NewData(db)
	if err != nil {
		return nil, fmt.Errorf("fail to initialize data layer: %w", err)
	}

	// data注入repo
	processDefinitionRepo := data.NewProcessDefinitionRepo(dataObj)
	instanceRepo := data.NewInstanceRepo(dataObj)
	executionRepo := data.NewExecutionRepo(dataObj)
	taskCmdRepo := data.NewCommandTaskRepo(dataObj)
	taskQueryRepo := data.NewQueryTaskRepo(dataObj)

	// MQ 初始化
	mq := biz.NewRedisMQ(rdb)

	// scheduler初始化
	txManager := data.NewTransactionManager(db)
	scheduler := biz.NewScheduler(
		txManager,
		processDefinitionRepo,
		instanceRepo,
		executionRepo,
		taskCmdRepo,
		biz.NewServiceTaskMQImpl(mq),
	)

	processDefinitionSvc := service.NewProcessDefinitionService(processDefinitionRepo, logger)
	processDefinitionHandler := api.NewProcessDefinitionHandler(processDefinitionSvc, logger)

	instanceSvc := service.NewInstanceService(instanceRepo, scheduler, logger)
	instanceHandler := api.NewInstanceHandler(instanceSvc, logger)

	taskSvc := service.NewTaskService(taskCmdRepo, taskQueryRepo, scheduler, logger)
	taskHandler := api.NewTaskHandler(taskSvc, logger)

	r := api.NewRouter(processDefinitionHandler, instanceHandler, taskHandler)

	// 允许跨域中间件配置
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"} // 前端地址
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	app := &App{
		Server:  r,
		Cleanup: cleanup,
	}
	return app, nil
}
