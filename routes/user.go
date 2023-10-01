package routes

import (
	"github.com/Caknoooo/golang-clean_template/controller"
	"github.com/Caknoooo/golang-clean_template/middleware"
	"github.com/Caknoooo/golang-clean_template/services"
	"github.com/gin-gonic/gin"
)

func User(route *gin.Engine, userController controller.UserController, jwtService services.JWTService) {
	routes := route.Group("/api/user")
	{
		// User
		routes.POST("", userController.Register)
		routes.GET("", userController.GetAllUser)
		routes.POST("/login", userController.Login)
		routes.DELETE("/", middleware.Authenticate(jwtService), userController.Delete)
		routes.PATCH("/", middleware.Authenticate(jwtService), userController.Update)
		routes.GET("/me", middleware.Authenticate(jwtService), userController.Me)
		routes.POST("/verify-email", userController.VerifyEmail)
		routes.POST("/verification-email", userController.SendVerificationEmail)

		// Admin
		routes.PATCH("/admin/verify", middleware.Authenticate(jwtService), userController.UpdateStatusIsVerified)
	}
}