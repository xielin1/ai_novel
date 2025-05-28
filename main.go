package main

import (
	"embed"
	"gin-template/common"
	"gin-template/middleware"
	"gin-template/model"
	"gin-template/repository"
	"gin-template/router"
	"gin-template/service"
	task2 "gin-template/service/task"
	"gin-template/task"
	"gin-template/util"
	"log"
	"os"
	"strconv"
	"time"

	_ "gin-template/docs" // 导入 swagg
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Gin Template API
// @version         1.0
// @description     Gin Template API 服务
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

//go:embed web/build
var buildFS embed.FS

//go:embed web/build/index.html
var indexPage []byte

func main() {
	common.SetupGinLog()
	common.SysLog("Gin Template " + common.Version + " started")
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化 UUID 生成器
	util.NewHybridGenerator(1)

	// Initialize SQL Database
	err := model.InitDB()
	if err != nil {
		common.FatalLog(err)
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.FatalLog(err)
		}
	}()
	controllers, err1 := InitializeAllController(model.DB)
	if err1 != nil {
		common.FatalLog(err1)
	}
	InitTokenService()

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		common.FatalLog(err)
	}

	// Initialize options
	model.InitOptionMap()

	// Initialize HTTP server
	server := gin.Default()
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.CORS())

	// Initialize session store
	if common.RedisEnabled {
		opt := common.ParseRedisOption()
		store, _ := redis.NewStore(opt.MinIdleConns, opt.Network, opt.Addr, opt.Password, []byte(common.SessionSecret))
		server.Use(sessions.Sessions("session", store))
	} else {
		store := cookie.NewStore([]byte(common.SessionSecret))
		server.Use(sessions.Sessions("session", store))
	}

	// 注册 Swagger 路由
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//创建补偿表和补偿调度器
	compensation, err2 := task.NewDBCompensation(model.DB)
	// 创建调度器（每30秒执行一次）todo 优化成配置,增加开关配置
	if err2 != nil {
		common.FatalLog(err)
	}
	scheduler := task.NewDBScheduler(compensation, 30*time.Second)
	InitTask(scheduler)
	scheduler.Start()

	router.SetRouter(server, buildFS, indexPage, controllers)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	err = server.Run(":" + port)
	if err != nil {
		log.Println(err)
	}
}

// InitTask 注册task
func InitTask(scheduler *task.DBScheduler) {
	scheduler.RegisterHandler(task.UserTokenInitCompensationTask, task2.CompensationUserTokensInit)
	scheduler.RegisterHandler(task.TokenCreditCompensationTask, task2.CompensationTokenCredit)
	scheduler.RegisterHandler(task.TokenDebitCompensationTask, task2.CompensationTokenDebit)
}

func InitTokenService() {
	tokenRepository := repository.NewTokenRepository(model.DB)
	tokenService := service.NewTokenService(tokenRepository)
	service.SetTokenService(tokenService)
	service.InitReconciliationService(tokenRepository, repository.NewTokenReconciliationRepository(model.DB))
}
