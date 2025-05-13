package repository

import (
	"context"
	"fmt"
	"github.com/Caknoooo/go-gin-clean-starter/repository"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/dto"
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	repository repository.UserRepository
	ctx        context.Context
	users      []entity.User
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	// Set up PostgreSQL connection for testing
	// You may want to use environment variables or test configuration here
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASS", "postgres"),
		getEnv("DB_NAME", "go_gin_test"),
		getEnv("DB_PORT", "5432"),
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      false,
		},
	)

	db, err := gorm.Open(
		postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		},
	)
	if err != nil {
		suite.T().Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&entity.User{})
	if err != nil {
		suite.T().Fatalf("Failed to migrate schema: %v", err)
	}

	suite.db = db
	suite.repository = repository.NewUserRepository(db)
	suite.ctx = context.Background()
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Clear the database before each test
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	// Create some test users
	suite.users = []entity.User{
		{
			ID:         uuid.New(),
			Name:       "Test User 1",
			Email:      "test1@example.com",
			TelpNumber: "12345678901",
			Password:   "password123",
			Role:       "user",
			ImageUrl:   "https://example.com/image1.jpg",
			IsVerified: true,
			Timestamp: entity.Timestamp{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			ID:         uuid.New(),
			Name:       "Test User 2",
			Email:      "test2@example.com",
			TelpNumber: "12345678902",
			Password:   "password123",
			Role:       "admin",
			ImageUrl:   "https://example.com/image2.jpg",
			IsVerified: false,
			Timestamp: entity.Timestamp{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			ID:         uuid.New(),
			Name:       "Test User 3",
			Email:      "test3@example.com",
			TelpNumber: "12345678903",
			Password:   "password123",
			Role:       "user",
			ImageUrl:   "https://example.com/image3.jpg",
			IsVerified: true,
			Timestamp: entity.Timestamp{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Insert test users into the database
	for _, user := range suite.users {
		err := suite.db.Create(&user).Error
		if err != nil {
			suite.T().Fatalf("Failed to create test user: %v", err)
		}
	}
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	// Clean up after all tests
	sqlDb, err := suite.db.DB()
	if err == nil {
		err := sqlDb.Close()
		if err != nil {
			suite.T().Fatalf("Failed to close database connection: %v", err)
		}
	}
}

func (suite *UserRepositoryTestSuite) TestRegister() {
	newUser := entity.User{
		ID:         uuid.New(),
		Name:       "New User",
		Email:      "new@example.com",
		TelpNumber: "12345678904",
		Password:   "password123",
		Role:       "user",
		ImageUrl:   "https://example.com/new.jpg",
		IsVerified: false,
		Timestamp: entity.Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Test with nil transaction
	createdUser, err := suite.repository.Register(suite.ctx, nil, newUser)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newUser.ID, createdUser.ID)
	assert.Equal(suite.T(), newUser.Name, createdUser.Name)
	assert.Equal(suite.T(), newUser.Email, createdUser.Email)

	// Test with transaction
	newUser2 := entity.User{
		ID:         uuid.New(),
		Name:       "New User 2",
		Email:      "new2@example.com",
		TelpNumber: "12345678905",
		Password:   "password123",
		Role:       "user",
		ImageUrl:   "https://example.com/new2.jpg",
		IsVerified: false,
		Timestamp: entity.Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	tx := suite.db.Begin()
	createdUser2, err := suite.repository.Register(suite.ctx, tx, newUser2)
	assert.NoError(suite.T(), err)
	tx.Commit()

	assert.Equal(suite.T(), newUser2.ID, createdUser2.ID)
	assert.Equal(suite.T(), newUser2.Name, createdUser2.Name)
	assert.Equal(suite.T(), newUser2.Email, createdUser2.Email)

	// Verify both users were saved to database
	var count int64
	suite.db.Model(&entity.User{}).Count(&count)
	assert.Equal(suite.T(), int64(5), count) // 3 initial users + 2 new ones
}

func (suite *UserRepositoryTestSuite) TestGetAllUserWithPagination() {
	// Test default pagination (no search)
	req := dto.PaginationRequest{
		Page:    1,
		PerPage: 2,
		Search:  "",
	}

	result, err := suite.repository.GetAllUserWithPagination(suite.ctx, nil, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(result.Users))
	assert.Equal(suite.T(), int64(3), result.Count)
	assert.Equal(suite.T(), int64(2), result.MaxPage)

	// Test pagination with search
	reqWithSearch := dto.PaginationRequest{
		Page:    1,
		PerPage: 10,
		Search:  "Test User 1",
	}

	resultWithSearch, err := suite.repository.GetAllUserWithPagination(suite.ctx, nil, reqWithSearch)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(resultWithSearch.Users))
	assert.Equal(suite.T(), int64(1), resultWithSearch.Count)
	assert.Equal(suite.T(), "Test User 1", resultWithSearch.Users[0].Name)

	// Test second page
	reqPage2 := dto.PaginationRequest{
		Page:    2,
		PerPage: 2,
		Search:  "",
	}

	resultPage2, err := suite.repository.GetAllUserWithPagination(suite.ctx, nil, reqPage2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(resultPage2.Users))
	assert.Equal(suite.T(), int64(3), resultPage2.Count)
}

func (suite *UserRepositoryTestSuite) TestGetUserById() {
	// Test with valid ID
	user, err := suite.repository.GetUserById(suite.ctx, nil, suite.users[0].ID.String())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.users[0].ID, user.ID)
	assert.Equal(suite.T(), suite.users[0].Name, user.Name)
	assert.Equal(suite.T(), suite.users[0].Email, user.Email)

	// Test with invalid ID
	invalidID := uuid.New().String()
	_, err = suite.repository.GetUserById(suite.ctx, nil, invalidID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)

	// Test with transaction
	tx := suite.db.Begin()
	user, err = suite.repository.GetUserById(suite.ctx, tx, suite.users[1].ID.String())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.users[1].ID, user.ID)
	assert.Equal(suite.T(), suite.users[1].Name, user.Name)
	tx.Commit()
}

func (suite *UserRepositoryTestSuite) TestGetUserByEmail() {
	// Test with valid email
	user, err := suite.repository.GetUserByEmail(suite.ctx, nil, suite.users[0].Email)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.users[0].ID, user.ID)
	assert.Equal(suite.T(), suite.users[0].Name, user.Name)
	assert.Equal(suite.T(), suite.users[0].Email, user.Email)

	// Test with invalid email
	_, err = suite.repository.GetUserByEmail(suite.ctx, nil, "nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)

	// Test with transaction
	tx := suite.db.Begin()
	user, err = suite.repository.GetUserByEmail(suite.ctx, tx, suite.users[1].Email)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.users[1].ID, user.ID)
	assert.Equal(suite.T(), suite.users[1].Email, user.Email)
	tx.Commit()
}

func (suite *UserRepositoryTestSuite) TestCheckEmail() {
	// Test with existing email
	user, exists, err := suite.repository.CheckEmail(suite.ctx, nil, suite.users[0].Email)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), suite.users[0].ID, user.ID)
	assert.Equal(suite.T(), suite.users[0].Email, user.Email)

	// Test with non-existing email
	_, exists, err = suite.repository.CheckEmail(suite.ctx, nil, "nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.False(suite.T(), exists)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)

	// Test with transaction
	tx := suite.db.Begin()
	user, exists, err = suite.repository.CheckEmail(suite.ctx, tx, suite.users[1].Email)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), suite.users[1].ID, user.ID)
	tx.Commit()
}

func (suite *UserRepositoryTestSuite) TestUpdate() {
	// Get a user to update
	user, err := suite.repository.GetUserById(suite.ctx, nil, suite.users[0].ID.String())
	assert.NoError(suite.T(), err)

	// Update user fields
	user.Name = "Updated Name"
	user.TelpNumber = "98765432109"
	user.UpdatedAt = time.Now()

	// Test update
	updatedUser, err := suite.repository.Update(suite.ctx, nil, user)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", updatedUser.Name)
	assert.Equal(suite.T(), "98765432109", updatedUser.TelpNumber)

	// Verify update in database
	var dbUser entity.User
	err = suite.db.Where("id = ?", suite.users[0].ID).First(&dbUser).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", dbUser.Name)
	assert.Equal(suite.T(), "98765432109", dbUser.TelpNumber)

	// Test with transaction
	user2, err := suite.repository.GetUserById(suite.ctx, nil, suite.users[1].ID.String())
	assert.NoError(suite.T(), err)

	user2.Name = "Updated with Tx"
	tx := suite.db.Begin()
	updatedUser2, err := suite.repository.Update(suite.ctx, tx, user2)
	assert.NoError(suite.T(), err)
	tx.Commit()

	assert.Equal(suite.T(), "Updated with Tx", updatedUser2.Name)
}

func (suite *UserRepositoryTestSuite) TestDelete() {
	// Test delete
	err := suite.repository.Delete(suite.ctx, nil, suite.users[0].ID.String())
	assert.NoError(suite.T(), err)

	// Verify user was deleted
	var count int64
	suite.db.Model(&entity.User{}).Count(&count)
	assert.Equal(suite.T(), int64(2), count) // Original 3 - 1 deleted

	// Verify the correct user was deleted
	_, err = suite.repository.GetUserById(suite.ctx, nil, suite.users[0].ID.String())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)

	// Test with transaction
	tx := suite.db.Begin()
	err = suite.repository.Delete(suite.ctx, tx, suite.users[1].ID.String())
	assert.NoError(suite.T(), err)
	tx.Commit()

	// Verify second user was deleted
	_, err = suite.repository.GetUserById(suite.ctx, nil, suite.users[1].ID.String())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)

	suite.db.Model(&entity.User{}).Count(&count)
	assert.Equal(suite.T(), int64(1), count) // 2 left - 1 more deleted = 1
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
