package app

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
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

	// data注入repo
	helloRepo := data.NewHelloRepo(logger)
	processDefinitionRepo := data.NewProcessDefinitionRepo(dataObj)
	instanceRepo := data.NewInstanceRepo(dataObj)
	TaskCmdRepo := data.NewCommandTaskRepo(dataObj)
	TaskQueryRepo := data.NewQueryTaskRepo(dataObj)

	// executor注入
	executor := make(map[string]biz.NodeBehaviour)
	executor[model.NodeTypeUserTask] = biz.NewUserNodeBehaviour(TaskCmdRepo)
	navigator := biz.NewNavigator()
	// 更多executor的注入......
	// scheduler初始化
	txManager := data.NewTransactionManager(db)
	scheduler := biz.NewScheduler(&biz.SchedulerDependencies{
		TxManager:      txManager,
		DefinitionRepo: processDefinitionRepo,
		InstanceRepo:   instanceRepo,
		TaskCmdRepo:    TaskCmdRepo,
		Navigator:      navigator,
		NodeExecutor:   executor,
	})

	helloSvc := service.NewHelloService(helloRepo, logger)
	helloHandler := handler.NewHelloHandler(helloSvc, logger)

	processDefinisionSvc := service.NewProcessDefinitionService(processDefinitionRepo, logger)
	processDefinitionHandler := handler.NewProcessDefinitionHandler(processDefinisionSvc, logger)

	instanceSvc := service.NewInstanceService(instanceRepo, scheduler, logger)
	instanceHandler := handler.NewInstanceHandler(instanceSvc, logger)

	taskSvc := service.NewTaskService(TaskCmdRepo, TaskQueryRepo, scheduler, logger)
	taskHandler := handler.NewTaskHandler(taskSvc, logger)

	r := api.NewRouter(helloHandler, processDefinitionHandler, instanceHandler, taskHandler)

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
