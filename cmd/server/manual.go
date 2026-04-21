package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hobbyGG/Dawnix/api"
	authAPI "github.com/hobbyGG/Dawnix/api/auth"
	"github.com/hobbyGG/Dawnix/api/workflow"
	authData "github.com/hobbyGG/Dawnix/internal/auth/data"
	authModel "github.com/hobbyGG/Dawnix/internal/auth/data/model"
	authService "github.com/hobbyGG/Dawnix/internal/auth/service"
	"github.com/hobbyGG/Dawnix/internal/workflow/biz"
	"github.com/hobbyGG/Dawnix/internal/workflow/conf"
	"github.com/hobbyGG/Dawnix/internal/workflow/data"
	dataModel "github.com/hobbyGG/Dawnix/internal/workflow/data/model"
	"github.com/hobbyGG/Dawnix/internal/workflow/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type App struct {
	Server   *gin.Engine
	HTTPAddr string
	Cleanup  func()
}

func (a *App) Run() error {
	if a.HTTPAddr == "" {
		a.HTTPAddr = ":8080"
	}
	return a.Server.Run(a.HTTPAddr)
}

func NewAppManual(logger *zap.Logger, cfg *conf.Bootstrap) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	// 数据库初始化化
	db, err := gorm.Open(postgres.Open(cfg.Data.Database.DSN), &gorm.Config{
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true, TablePrefix: "dawnix_"},
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to initialize postgresSQL: %w", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(cfg.Data.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Data.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Data.Database.ConnMaxLifetime) * time.Minute)
	db.AutoMigrate(
		&dataModel.ProcessDefinition{},
		&dataModel.ProcessTask{},
		&dataModel.ProcessInstance{},
		&dataModel.Execution{},
		&authModel.User{},
		&authModel.AuthIdentity{},
	)

	// rdb 初始化
	redOpts := &redis.Options{
		Addr: cfg.Data.Redis.Addr,
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
	taskRepo := data.NewTaskRepo(dataObj)

	// MQ 初始化（基础设施实现位于 data 层）
	mq := data.NewRedisMQ(rdb)
	nodeRegistry := biz.NewDefaultNodeRegistry(biz.NodeDeps{
		TaskRepo:      taskRepo,
		InstanceRepo:  instanceRepo,
		ExecutionRepo: executionRepo,
		ServiceTaskMQ: data.NewServiceTaskMQ(mq),
	}, biz.NodeFeatures{
		EmailServiceEnabled: cfg.Biz.Features.EmailService.Enabled,
	})

	// scheduler初始化
	txManager := data.NewTransactionManager(db)
	scheduler := biz.NewScheduler(
		txManager,
		processDefinitionRepo,
		instanceRepo,
		executionRepo,
		taskRepo,
		nodeRegistry,
	)

	processDefinitionSvc := service.NewProcessDefinitionService(processDefinitionRepo, logger, cfg.Biz.Features.EmailService.Enabled)
	processDefinitionHandler := workflow.NewProcessDefinitionHandler(processDefinitionSvc, logger)
	enumHandler := workflow.NewEnumHandler(cfg.Biz.Features.EmailService.Enabled)

	instanceSvc := service.NewInstanceService(instanceRepo, scheduler, logger)
	instanceHandler := workflow.NewInstanceHandler(instanceSvc, logger)

	taskSvc := service.NewTaskService(taskRepo, scheduler, logger, cfg.Biz.Task.DefaultScope)
	taskHandler := workflow.NewTaskHandler(taskSvc, logger)
	authRepo := authData.NewRepo(dataObj)
	authSvc, err := authService.NewService(authRepo, authService.Config{
		JWTSecret:        cfg.Auth.JWT.Secret,
		JWTIssuer:        cfg.Auth.JWT.Issuer,
		JWTExpireMinutes: cfg.Auth.JWT.ExpireMinutes,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("fail to initialize auth service: %w", err)
	}
	authHandler := authAPI.NewHandler(authSvc, logger)
	authMiddleware := authService.JWTMiddleware(authSvc)

	// 允许跨域中间件配置
	config := cors.DefaultConfig()
	config.AllowOrigins = cfg.Server.CORS.AllowOrigins
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}

	r := api.NewRouter(
		[]gin.HandlerFunc{cors.New(config), authMiddleware},
		authHandler,
		processDefinitionHandler,
		enumHandler,
		instanceHandler,
		taskHandler,
	)

	app := &App{
		Server:   r,
		HTTPAddr: cfg.Server.HTTP.Addr,
		Cleanup:  cleanup,
	}
	return app, nil
}
