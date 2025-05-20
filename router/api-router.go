package router

import (
	"gin-template/controller"
	"gin-template/middleware"

	"github.com/gin-gonic/gin"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	{
		// 初始化控制器
		referralController := controller.NewReferralController()

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
			aiRoute.POST("/prompt", controller.AIPrompt)         // 提交提示并获取响应
			aiRoute.GET("/models", controller.GetAIModels)       // 获取可用模型列表
			aiRoute.POST("/generate/:id", controller.AIGenerate) // AI续写
		}

		// 项目管理API路由
		projectRoute := apiRouter.Group("/projects")
		projectRoute.Use(middleware.UserAuth()) // 需要登录才能使用
		{
			projectRoute.GET("", controller.GetProjects)          // 获取项目列表
			projectRoute.POST("", controller.CreateProject)       // 创建新项目
			projectRoute.GET("/:id", controller.GetProject)       // 获取项目详情
			projectRoute.PUT("/:id", controller.UpdateProject)    // 更新项目信息
			projectRoute.DELETE("/:id", controller.DeleteProject) // 删除项目
		}

		// 大纲管理API路由
		apiRouter.GET("/outlines/:id", middleware.UserAuth(), controller.GetOutline)           // 获取大纲内容
		apiRouter.POST("/outlines/:id", middleware.UserAuth(), controller.SaveOutline)         // 保存大纲内容
		apiRouter.GET("/versions/:id", middleware.UserAuth(), controller.GetVersions)          // 获取版本历史
		apiRouter.POST("/outline/parse/:id", middleware.UserAuth(), controller.ParseOutline)   // 解析大纲文件
		apiRouter.POST("/upload/outline/:id", middleware.UserAuth(), controller.UploadOutline) // 上传大纲文件（兼容旧接口）
		apiRouter.POST("/exports/:id", middleware.UserAuth(), controller.ExportOutline)        // 导出大纲

		// 套餐管理API路由
		packageRoute := apiRouter.Group("/packages")
		packageRoute.Use(middleware.UserAuth()) // 需要登录才能使用
		{
			//packageRoute.GET("", controller.GetPackages)                   // 获取套餐列表
			//packageRoute.POST("/subscribe", controller.SubscribePackage)   // 购买/订阅套餐
			//packageRoute.POST("/cancel-renewal", controller.CancelRenewal) // 取消自动续费
		}

		//user
		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
			userRoute.POST("/login", middleware.CriticalRateLimit(), controller.Login)
			userRoute.GET("/logout", controller.Logout)
			userRoute.POST("/referral", middleware.UserAuth(), referralController.UseReferral) // 使用他人推荐码

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth(), middleware.NoTokenAuth())
			{
				selfRoute.GET("/self", controller.GetSelf)
				selfRoute.PUT("/self", controller.UpdateSelf)
				selfRoute.DELETE("/self", controller.DeleteSelf)
				selfRoute.GET("/token", controller.GenerateToken)
				//selfRoute.GET("/package", controller.GetUserPackage)                               // 获取当前用户的套餐信息
				selfRoute.GET("/referral-code", referralController.GetReferralCode)                // 获取个人推荐码
				selfRoute.GET("/referrals", referralController.GetReferrals)                       // 获取推荐记录
				selfRoute.POST("/generate-referral-code", referralController.GenerateReferralCode) // 生成新的推荐码
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

		//option
		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth(), middleware.NoTokenAuth())
		{
			optionRoute.GET("/", controller.GetOptions)
			optionRoute.PUT("/", controller.UpdateOption)
		}

		//file
		fileRoute := apiRouter.Group("/file")
		fileRoute.Use(middleware.AdminAuth())
		{
			fileRoute.GET("/", controller.GetAllFiles)
			fileRoute.GET("/search", controller.SearchFiles)
			fileRoute.POST("/", middleware.UploadRateLimit(), controller.UploadFile)
			fileRoute.DELETE("/:id", controller.DeleteFile)
		}
	}
}
