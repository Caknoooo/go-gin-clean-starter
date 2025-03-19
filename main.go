package main

import (
	"log"
	"os"

	"github.com/Caknoooo/go-gin-clean-starter/command"
	"github.com/Caknoooo/go-gin-clean-starter/middleware"
	"github.com/Caknoooo/go-gin-clean-starter/provider"
	"github.com/Caknoooo/go-gin-clean-starter/routes"
	"github.com/samber/do"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
)

func args(injector *do.Injector) bool {
	if len(os.Args) > 1 {
		flag := command.Commands(injector)
		if !flag {
			return false
		}
	}

	return true
}

func main() {
	var (
		injector = do.New()
	)

	if !args(injector) {
		return
	}

	provider.RegisterDependencies(injector)

	server := gin.Default()
	server.Use(middleware.CORSMiddleware())

	// routes
	routes.User(server, injector)

	server.Static("/assets", "./assets")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	var serve string
	if os.Getenv("APP_ENV") == "localhost" {
		serve = "127.0.0.1:" + port
	} else {
		serve = ":" + port
	}

	myFigure := figure.NewColorFigure("Caknoo", "", "green", true)
	myFigure.Print()

	if err := server.Run(serve); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}
