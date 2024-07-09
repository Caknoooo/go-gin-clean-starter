package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/controller"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/repository"
	"github.com/Caknoooo/go-gin-clean-starter/service"
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
		jwtService     = service.NewJWTService()
		userService    = service.NewUserService(userRepo, jwtService)
		userController = controller.NewUserController(userService)
	)

	return userController
}

func InsertTestUser() ([]entity.User, error) {
	db := SetUpDatabaseConnection()
	users := []entity.User{
		{
			Name:  "admin",
			Email: "admin1234@gmail.com",
		},
		{
			Name:  "user",
			Email: "user1234@gmail.com",
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}
	}

	return users, nil
}

func Test_GetAllUser_OK(t *testing.T) {
	r := SetUpRoutes()
	userController := SetupControllerUser()
	r.GET("/api/user", userController.GetAllUser)

	expectedUsers, err := InsertTestUser()
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/user", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	type Response struct {
		Data []entity.User `json:"data"`
	}

	var response Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	actualUsers := response.Data

	for _, expectedUser := range expectedUsers {
		found := false
		for _, actualUser := range actualUsers {
			if expectedUser.Name == actualUser.Name && expectedUser.Email == actualUser.Email {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected user not found in actual users: %v", expectedUser)
	}
}
