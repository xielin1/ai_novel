package router

import (
	"gin-template/controller"
	"gin-template/middleware"

	"github.com/gin-gonic/gin"
)

// APIControllers 包含所有需要注入的API控制器
type APIControllers struct {
	// 基础控制器

	ReferralController *controller.ReferralController

	// 项目与大纲控制器
	ProjectController *controller.ProjectController
	OutlineController *controller.OutlineController

	// 套餐控制器
	PackageController *controller.PackageController

	HealthController *controller.HealthController

	AgentController *controller.AgentController
}

// SetApiRouter 使用依赖注入的控制器设置API路由
func SetApiRouter(router *gin.Engine, controllers *APIControllers) {

	apiRouter := router.Group("/api")
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	{

		// 健康检查
		apiRouter.GET("/v1/health", controllers.HealthController.HealthCheck)

		// 基础API
		apiRouter.GET("/status", controller.GetStatus)
		apiRouter.GET("/notice", controller.GetNotice)
		apiRouter.GET("/about", controller.GetAbout)
		apiRouter.GET("/homepage", controller.GetHomePageContent)
		apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)
		apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), controller.GitHubOAuth)
		apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), controller.WeChatAuth)
		apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.WeChatBind)
		apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.EmailBind)

		// OpenAI API路由
		aiRoute := apiRouter.Group("/ai")
		aiRoute.Use(middleware.UserAuth()) // 需要登录才能使用
		{
			aiRoute.POST("/prompt", controller.AIPrompt)   // 提交提示并获取响应
			aiRoute.GET("/models", controller.GetAIModels) // 获取可用模型列表
			//aiRoute.POST("/generate/:id", controller.AIGenerate) // AI续写
		}

		// 智能体相关路由
		agentGroup := apiRouter.Group("/v1/agent")
		{
			agentGroup.POST("/chat", controllers.AgentController.Chat)
			//agentGroup.POST("/chat/stream", agentController.StreamChat)
		}

		// 项目管理API路由
		projectRoute := apiRouter.Group("/projects")
		projectRoute.Use(middleware.UserAuth()) // 需要登录才能使用
		{
			projectRoute.GET("", controllers.ProjectController.GetProjects)          // 获取项目列表
			projectRoute.POST("", controllers.ProjectController.CreateProject)       // 创建新项目
			projectRoute.GET("/:id", controllers.ProjectController.GetProject)       // 获取项目详情
			projectRoute.PUT("/:id", controllers.ProjectController.UpdateProject)    // 更新项目信息
			projectRoute.DELETE("/:id", controllers.ProjectController.DeleteProject) // 删除项目
		}

		// 大纲管理API路由
		outlineRoute := apiRouter.Group("/outlines")
		outlineRoute.Use(middleware.UserAuth()) // 需要登录才能使用
		{
			outlineRoute.GET("/:id", controllers.OutlineController.GetOutline)           // 获取大纲内容
			outlineRoute.POST("/:id", controllers.OutlineController.SaveOutline)         // 保存大纲内容
			outlineRoute.GET("/versions/:id", controllers.OutlineController.GetVersions) // 获取版本历史
			outlineRoute.POST("/parse/:id", controllers.OutlineController.ParseOutline)  // 解析大纲文件
			outlineRoute.POST("/upload/:id", controllers.OutlineController.ParseOutline) // 上传大纲文件（兼容旧接口）
			//outlineRoute.POST("/export/:id", controller.ExportOutline)  // 导出大纲
		}

		// 套餐管理API路由
		packageRoute := apiRouter.Group("/package")
		packageRoute.Use(middleware.UserAuth()) // 需要登录才能使用
		{
			packageRoute.GET("/all", controllers.PackageController.GetPackages)               // 获取套餐列表
			packageRoute.POST("/subscribe", controllers.PackageController.SubscribePackage)   // 购买/订阅套餐
			packageRoute.POST("/cancel-renewal", controllers.PackageController.CancelRenewal) // 取消自动续费
			packageRoute.GET("/:id", controllers.PackageController.GetPackageByID)            // 新增根据ID获取套餐接口
		}

		// 用户管理API路由
		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
			userRoute.POST("/login", middleware.CriticalRateLimit(), controller.Login)
			userRoute.GET("/logout", controller.Logout)
			userRoute.POST("/referral", middleware.UserAuth(), controllers.ReferralController.UseReferral) // 使用他人推荐码

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth(), middleware.NoTokenAuth())
			{
				selfRoute.GET("/self", controller.GetSelf)
				selfRoute.PUT("/self", controller.UpdateSelf)
				selfRoute.DELETE("/self", controller.DeleteSelf)
				selfRoute.GET("/token", controller.GenerateToken)
				selfRoute.GET("/package", controllers.PackageController.GetUserPackage)         // 获取当前用户的套餐信息
				selfRoute.GET("/referral-code", controllers.ReferralController.GetReferralCode) // 获取个人推荐码
				selfRoute.GET("/referrals", controllers.ReferralController.GetReferrals)        // 获取推荐记录
			}

			adminRoute := userRoute.Group("/")
			adminRoute.Use(middleware.AdminAuth(), middleware.NoTokenAuth())
			{
				adminRoute.GET("/", controller.GetAllUsers)
				adminRoute.GET("/search", controller.SearchUsers)
				adminRoute.GET("/:id", controller.GetUser)
				adminRoute.POST("/", controller.CreateUser)
				adminRoute.POST("/manage", controller.ManageUser)
				adminRoute.PUT("/", controller.UpdateUser)
				adminRoute.DELETE("/:id", controller.DeleteUser)
			}
		}

		// 系统配置API路由
		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth(), middleware.NoTokenAuth())
		{
			optionRoute.GET("/", controller.GetOptions)
			optionRoute.PUT("/", controller.UpdateOption)
		}

		// 文件管理API路由
		fileRoute := apiRouter.Group("/file")
		fileRoute.Use(middleware.AdminAuth())
		{
			fileRoute.GET("/", controller.GetAllFiles)
			fileRoute.GET("/search", controller.SearchFiles)
			fileRoute.POST("/", middleware.UploadRateLimit(), controller.UploadFile)
			fileRoute.DELETE("/:id", controller.DeleteFile)
		}

		//后续需要启用时再处理
		//api := router.Group("/api/reconciliation")
		//{
		//	api.POST("/start", controller.StartService)
		//	api.POST("/stop", controller.StopService)
		//	api.POST("/full", controller.FullReconciliation)
		//	api.POST("/user", controller.UserReconciliation) // 使用JSON传参方式（可选）
		//
		//}

	}
}
