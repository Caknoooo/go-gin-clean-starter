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
		routes.POST("", userController.RegisterUser)
		routes.GET("", userController.GetAllUser)
		routes.POST("/login", userController.LoginUser)
		routes.DELETE("/", middleware.Authenticate(jwtService), userController.DeleteUser)
		routes.PATCH("/", middleware.Authenticate(jwtService), userController.UpdateUser)
		routes.GET("/me", middleware.Authenticate(jwtService), userController.MeUser)

		// Admin
		routes.PATCH("/verify", middleware.Authenticate(jwtService), userController.UpdateStatusIsVerified)
	}
}