package providers

import (
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/service"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func ProvideUserDependencies(injector *do.Injector, db *gorm.DB, jwtService authService.JWTService) {
	// Repository
	userRepository := repository.NewUserRepository(db)
	refreshTokenRepository := authRepo.NewRefreshTokenRepository(db)

	// Service
	userService := service.NewUserService(userRepository, refreshTokenRepository, jwtService, db)

	// Controller
	do.Provide(
		injector, func(i *do.Injector) (controller.UserController, error) {
			return controller.NewUserController(i, userService), nil
		},
	)
}
