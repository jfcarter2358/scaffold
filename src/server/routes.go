// routes.go

package main

import (
	"net/http"
	"scaffold/server/api"
	"scaffold/server/auth"
	"scaffold/server/config"
	"scaffold/server/constants"
	"scaffold/server/docs"
	"scaffold/server/manager"
	"scaffold/server/middleware"
	"scaffold/server/page"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func initializeRoutes() {
	router.Static("/static/css", "./static/css")
	router.Static("/static/img", "./static/img")
	router.Static("/static/js", "./static/js")

	// Swagger docs
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.GET("/", page.RedirectIndexPage)

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{})
	})

	healthRoutes := router.Group("/health", middleware.CORSMiddleware())
	{
		healthRoutes.GET("/healthy", api.Healthy)
		healthRoutes.GET("/ready", api.Ready)
		if config.Config.Node.Type == constants.NODE_TYPE_WORKER {
			healthRoutes.GET("/available", api.Available)
		} else {
			healthRoutes.GET("/status", middleware.EnsureLoggedIn(), manager.GetStatus)
			healthRoutes.POST("/ping/:name", middleware.EnsureLoggedIn(), api.Ping)
		}
	}

	if config.Config.Node.Type == constants.NODE_TYPE_MANAGER {
		authRoutes := router.Group("/auth", middleware.CORSMiddleware())
		{
			authRoutes.POST("/login", middleware.EnsureNotLoggedIn(), auth.PerformLogin)
			authRoutes.GET("/logout", middleware.EnsureLoggedIn(), auth.PerformLogout)
			authRoutes.POST("/reset/request", middleware.EnsureNotLoggedIn(), auth.RequestPasswordReset)
			authRoutes.POST("/reset/do", middleware.EnsureNotLoggedIn(), auth.DoPasswordReset)
			authRoutes.POST("/join", auth.JoinNode)
			authRoutes.POST("/token/:username/:name", middleware.EnsureLoggedIn(), middleware.EnsureSelf(), api.GenerateAPIToken)
			authRoutes.DELETE("/token/:username/:name", middleware.EnsureLoggedIn(), middleware.EnsureSelf(), api.RevokeAPIToken)
		}

		apiRoutes := router.Group("/api", middleware.CORSMiddleware())
		{
			v1Routes := apiRoutes.Group("/v1")
			{
				cascadeRoutes := v1Routes.Group("/cascade")
				{
					cascadeRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllCascades)
					cascadeRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("name"), api.GetCascadeByName)
					cascadeRoutes.DELETE("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("name"), api.DeleteCascadeByName)
					cascadeRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateCascade)
					cascadeRoutes.PUT("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("name"), api.UpdateCascadeByName)
				}
				datastoreRoutes := v1Routes.Group("/datastore")
				{
					datastoreRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllDataStores)
					datastoreRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("name"), api.GetDataStoreByName)
					datastoreRoutes.DELETE("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("name"), api.DeleteDataStoreByCascade)
					datastoreRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateDataStore)
					datastoreRoutes.PUT("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("name"), api.UpdateDataStoreByCascade)
				}
				fileRoutes := v1Routes.Group("/file")
				{
					fileRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllFiles)
					fileRoutes.GET("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("name"), api.GetFilesByCascade)
					fileRoutes.GET("/:name/:file", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("name"), api.GetFileByNames)
					// fileRoutes.GET("/:name/:file/view", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("name"), api.ViewFile)
					fileRoutes.GET("/:name/:file/download", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("name"), api.DownloadFile)
					fileRoutes.POST("/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("name"), api.UploadFile)
				}
				stateRoutes := v1Routes.Group("/state")
				{
					stateRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllStates)
					stateRoutes.GET("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("cascade"), api.GetStatesByCascade)
					stateRoutes.GET("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("cascade"), api.GetStateByNames)
					stateRoutes.DELETE("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.DeleteStatesByCascade)
					stateRoutes.DELETE("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.DeleteStateByNames)
					stateRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateState)
					stateRoutes.PUT("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.UpdateStateByNames)
				}
				inputRoutes := v1Routes.Group("/input")
				{
					inputRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllInputs)
					inputRoutes.GET("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("cascade"), api.GetInputsByCascade)
					inputRoutes.GET("/:cascade/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("cascade"), api.GetInputByNames)
					inputRoutes.DELETE("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.DeleteInputsByCascade)
					inputRoutes.DELETE("/:cascade/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.DeleteInputByNames)
					inputRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateInput)
					inputRoutes.POST("/:cascade/update", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.UpdateInputDependenciesByName)
					inputRoutes.PUT("/:cascade/:name", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.UpdateInputByNames)
				}
				taskRoutes := v1Routes.Group("/task")
				{
					taskRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllTasks)
					taskRoutes.GET("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("cascade"), api.GetTasksByCascade)
					taskRoutes.GET("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), middleware.EnsureCascadeGroup("cascade"), api.GetTaskByNames)
					taskRoutes.DELETE("/:cascade", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.DeleteTasksByCascade)
					taskRoutes.DELETE("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.DeleteTaskByNames)
					taskRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.CreateTask)
					taskRoutes.PUT("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.UpdateTaskByNames)
					taskRoutes.PUT("/:cascade/:task/enabled", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.ToggleTaskEnabled)
				}
				userRoutes := v1Routes.Group("/user")
				{
					userRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetAllUsers)
					userRoutes.GET("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write", "read"}), api.GetUserByUsername)
					userRoutes.DELETE("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.DeleteUserByUsername)
					userRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin"}), api.CreateUser)
					userRoutes.PUT("/:username", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), api.UpdateUserByUsername)
				}
				runRoutes := v1Routes.Group("/run")
				{
					runRoutes.POST(":cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.CreateRun)
					runRoutes.GET("/containers", middleware.EnsureLoggedIn(), api.GetAllContainers)
					runRoutes.DELETE("/:cascade/:task", middleware.EnsureLoggedIn(), middleware.EnsureCascadeGroup("cascade"), api.ManagerKillRun)
				}
				webhookRoutes := v1Routes.Group("/webhook")
				{
					webhookRoutes.GET("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin"}), api.GetAllWebhooks)
					webhookRoutes.GET("/:id", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin"}), api.GetWebhookByID)
					webhookRoutes.DELETE("/:id", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin"}), api.DeleteWebhookByID)
					webhookRoutes.POST("", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin"}), api.CreateWebhook)
					webhookRoutes.PUT("/:id", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin"}), api.UpdateWebhooksByID)
					webhookRoutes.POST("/:cascade/:id", middleware.EnsureLoggedIn(), middleware.EnsureRolesAllowed([]string{"admin", "write"}), middleware.EnsureCascadeGroup("cascade"), api.TriggerWebhookByID)
				}
			}
		}

		uiRoutes := router.Group("/ui", middleware.CORSMiddleware())
		{
			uiRoutes.GET("/login", middleware.EnsureNotLoggedIn(), page.ShowLoginPage)
			uiRoutes.GET("/forgot_password", middleware.EnsureNotLoggedIn(), page.ShowForgotPasswordPage)
			uiRoutes.GET("/email_success", middleware.EnsureNotLoggedIn(), page.ShowEmailSuccessPage)
			uiRoutes.GET("/email_failure", middleware.EnsureNotLoggedIn(), page.ShowEmailFailurePage)
			uiRoutes.GET("/reset_password/:reset_password", middleware.EnsureNotLoggedIn(), page.ShowResetPasswordPage)

			uiRoutes.GET("/cascades", middleware.EnsureLoggedIn(), page.ShowCascadesPage)
			uiRoutes.GET("/cascades/:name", middleware.EnsureLoggedIn(), page.ShowCascadePage)

			uiRoutes.GET("/files", middleware.EnsureLoggedIn(), page.ShowFilesPage)

			uiRoutes.GET("/users", middleware.EnsureLoggedIn(), page.ShowUsersPage)
			uiRoutes.GET("/user/:username", middleware.EnsureLoggedIn(), page.ShowUserPage)

			uiRoutes.GET("/webhooks", middleware.EnsureLoggedIn(), page.ShowWebhooksPage)
		}
	}
	if config.Config.Node.Type == constants.NODE_TYPE_WORKER {
		apiRoutes := router.Group("/api", middleware.CORSMiddleware())
		{
			v1Routes := apiRoutes.Group("/v1")
			{
				taskRoutes := v1Routes.Group("/run")
				{
					taskRoutes.DELETE("/:cascade/:task", middleware.EnsureLoggedIn(), api.KillRun)
				}
			}
		}
	}
}
