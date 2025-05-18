package provider

import (
	"github.com/Caknoooo/go-gin-clean-starter/config"
	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/Caknoooo/go-gin-clean-starter/service"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(
		injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
			return config.SetUpDatabaseConnection(), nil
		},
	)
}

var RegisterDependencies = func(injector *do.Injector) {
	InitDatabase(injector)

	do.ProvideNamed(
		injector, constants.JWTService, func(i *do.Injector) (service.JWTService, error) {
			return service.NewJWTService(), nil
		},
	)

	ProvideUserDependencies(injector)
}
