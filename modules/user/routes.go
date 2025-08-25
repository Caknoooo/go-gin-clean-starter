package user

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	userController := do.MustInvoke[controller.UserController](injector)

	userRoutes := server.Group("/api/v1")
	{
		userRoutes.GET("/user", userController.GetAllUser)
		userRoutes.GET("/user/:id", userController.Me)
		userRoutes.POST("/user", userController.Register)
		userRoutes.PUT("/user/:id", userController.Update)
		userRoutes.DELETE("/user/:id", userController.Delete)
		userRoutes.POST("/user/send-verification-email", userController.SendVerificationEmail)
		userRoutes.POST("/user/verify-email", userController.VerifyEmail)
		userRoutes.POST("/user/login", userController.Login)
		userRoutes.POST("/user/refresh", userController.Refresh)
	}
}
