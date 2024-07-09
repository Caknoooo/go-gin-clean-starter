package main

import (
	"log"
	"os"
	"sync"

	"github.com/Caknoooo/go-gin-clean-starter/config"
	"github.com/Caknoooo/go-gin-clean-starter/controller"
	"github.com/Caknoooo/go-gin-clean-starter/middleware"
	"github.com/Caknoooo/go-gin-clean-starter/migrations"
	"github.com/Caknoooo/go-gin-clean-starter/repository"
	"github.com/Caknoooo/go-gin-clean-starter/routes"
	"github.com/Caknoooo/go-gin-clean-starter/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	var (
		db             *gorm.DB                  = config.SetUpDatabaseConnection()
		jwtService     service.JWTService        = service.NewJWTService()
		userRepository repository.UserRepository = repository.NewUserRepository(db)
		userService    service.UserService       = service.NewUserService(userRepository, jwtService)
		userController controller.UserController = controller.NewUserController(userService)
	)

	server := gin.Default()
	server.Use(middleware.CORSMiddleware())

	// routes
	routes.User(server, userController, jwtService)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := migrations.Migrate(db); err != nil {
			log.Fatalf("error migration: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := migrations.Seeder(db); err != nil {
			log.Fatalf("error migration seeder: %v", err)
		}
	}()

	wg.Wait()

	server.Static("/assets", "./assets")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	if err := server.Run("localhost:" + port); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}
