package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caknoooo/go-gin-clean-template/controller"
	"github.com/Caknoooo/go-gin-clean-template/entity"
	"github.com/Caknoooo/go-gin-clean-template/repository"
	"github.com/Caknoooo/go-gin-clean-template/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func SetUpRoutes() *gin.Engine {
	r := gin.Default()
	return r
}

func SetupControllerUser() controller.UserController {
	var (
		db             = SetUpDatabaseConnection()
		userRepo       = repository.NewUserRepository(db)
		userService    = service.NewUserService(userRepo)
		jwtService     = service.NewJWTService()
		userController = controller.NewUserController(userService, jwtService)
	)

	return userController
}

func Test_GetAllUser_OK(t *testing.T) {
	r := SetUpRoutes()
	userController := SetupControllerUser()
	r.GET("/api/user", userController.GetAllUser)

	req, _ := http.NewRequest(http.MethodGet, "/api/user", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	users := []entity.User{
		{
			Name:  "testing",
			Email: "testing1@gmail.com",
		},
		{
			Name:  "testing2",
			Email: "testing2@gmail.com",
		},
	}

	expectedUsers, err := InsertTestBook()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, users, expectedUsers, "Success Get All User")
}

func InsertTestBook() ([]entity.User, error) {
	user := []entity.User{
		{
			Name:  "testing",
			Email: "testing1@gmail.com",
		},
		{
			Name:  "testing2",
			Email: "testing2@gmail.com",
		},
	}

	return user, nil
}
