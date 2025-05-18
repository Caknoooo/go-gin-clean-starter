package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

var RegisterRoutes = func(server *gin.Engine, injector *do.Injector) {
	User(server, injector)
}
