package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// Auth routes akan ditambahkan nanti ketika auth controller sudah dibuat
	authRoutes := server.Group("/api/v1/auth")
	{
		// authRoutes.POST("/refresh-token", authController.RefreshToken)
		_ = authRoutes // untuk menghindari unused variable
	}
}
