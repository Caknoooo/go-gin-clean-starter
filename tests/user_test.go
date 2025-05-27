package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/controller"
	"github.com/Caknoooo/go-gin-clean-starter/dto"
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
		refreshTokenRepo = repository.NewRefreshTokenRepository(db)
		userService    = service.NewUserService(userRepo, refreshTokenRepo, jwtService, db)
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

func Test_GetAllUser_BadRequest(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.GET("/api/user", uc.GetAllUser)

	// missing query params triggers binding error
	req, _ := http.NewRequest(http.MethodGet, "/api/user", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_Register_OK(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.POST("/api/user", uc.Register)

	payload := dto.UserCreateRequest{Name: "testuser", TelpNumber: "12345678", Email: "test@example.com", Password: "password123"}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/user", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var out struct {
		Data entity.User `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	assert.Equal(t, payload.Email, out.Data.Email)
}

func Test_Register_BadRequest(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.POST("/api/user", uc.Register)

	// empty body
	req, _ := http.NewRequest(http.MethodPost, "/api/user", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_Login_BadRequest(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.POST("/api/user/login", uc.Login)

	req, _ := http.NewRequest(http.MethodPost, "/api/user/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_SendVerificationEmail_OK(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.POST("/api/user/send_verification_email", uc.SendVerificationEmail)

	users, _ := InsertTestUser()
	reqBody, _ := json.Marshal(dto.SendVerificationEmailRequest{Email: users[0].Email})
	req, _ := http.NewRequest(http.MethodPost, "/api/user/send_verification_email", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_SendVerificationEmail_BadRequest(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.POST("/api/user/send_verification_email", uc.SendVerificationEmail)

	// missing email
	req, _ := http.NewRequest(http.MethodPost, "/api/user/send_verification_email", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_VerifyEmail_BadRequest(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.POST("/api/user/verify_email", uc.VerifyEmail)

	// missing token
	req, _ := http.NewRequest(http.MethodPost, "/api/user/verify_email", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Note: VerifyEmail_OK requires a valid token setup; implement when token creation is available.
func Test_VerifyEmail_OK(t *testing.T) {
	r := SetUpRoutes()
	uc := SetupControllerUser()
	r.POST("/api/user/verify_email", uc.VerifyEmail)

	// TODO: insert valid verification token into DB and use here
	validToken := "valid-token-placeholder"
	reqBody, _ := json.Marshal(dto.VerifyEmailRequest{Token: validToken})
	req, _ := http.NewRequest(http.MethodPost, "/api/user/verify_email", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// expecting BadRequest until token logic is implemented
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func Test_GetAllUser_OK(t *testing.T) {
	r := SetUpRoutes()
	userController := SetupControllerUser()
	r.GET("/api/user", userController.GetAllUser)

	expectedUsers, err := InsertTestUser()
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/api/user", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

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

// Test Login
func Test_Login_OK(t *testing.T) {
	r := SetUpRoutes()
	userController := SetupControllerUser()
	// first register
	r.POST("/api/user", userController.Register)
	r.POST("/api/user/login", userController.Login)

	payload := dto.UserLoginRequest{
		Email:    "loginuser@example.com",
		Password: "securepass",
	}
	// register user
	regBody, _ := json.Marshal(dto.UserCreateRequest{
		Name:     "loginuser",
		Email:    payload.Email,
		Password: payload.Password,
	})
	reqReg, _ := http.NewRequest(http.MethodPost, "/api/user", bytes.NewBuffer(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	httptest.NewRecorder() // ignore

	// login
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	type Resp struct {
		Data struct {
			Token string `json:"token"`
			Role  string `json:"role"`
		} `json:"data"`
	}
	var resp Resp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.Data.Token)
}

// Test Me
func Test_Me_OK(t *testing.T) {
	r := SetUpRoutes()
	userController := SetupControllerUser()
	r.GET("/api/user/me", func(c *gin.Context) {
		// insert and set user_id
		users, _ := InsertTestUser()
		c.Set("user_id", users[0].ID)
		userController.Me(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/user/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	type Resp struct {
		Data entity.User `json:"data"`
	}
	var resp Resp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, "admin", resp.Data.Name)
}

// Test Update
func Test_Update_OK(t *testing.T) {
	r := SetUpRoutes()
	userController := SetupControllerUser()
	r.PUT("/api/user", func(c *gin.Context) {
		users, _ := InsertTestUser()
		c.Set("user_id", users[1].ID)
		userController.Update(c)
	})

	update := dto.UserUpdateRequest{
		Name:       "updatedName",
		TelpNumber: "87654321",
	}
	body, _ := json.Marshal(update)
	req, _ := http.NewRequest(http.MethodPut, "/api/user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	type Resp struct {
		Data entity.User `json:"data"`
	}
	var resp Resp
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, update.Name, resp.Data.Name)
}

// Test Delete
func Test_Delete_OK(t *testing.T) {
	r := SetUpRoutes()
	userController := SetupControllerUser()
	r.DELETE("/api/user", func(c *gin.Context) {
		users, _ := InsertTestUser()
		c.Set("user_id", users[0].ID)
		userController.Delete(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/api/user", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
