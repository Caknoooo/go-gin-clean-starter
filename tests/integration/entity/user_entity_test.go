package entity_test

import (
	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"os"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	// Start the test container
	testContainer, err := container.StartTestContainer()
	if err != nil {
		panic(err)
	}

	// Set environment variables for database connection
	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_PORT", testContainer.Port)

	// Run tests
	code := m.Run()

	// Cleanup
	if err := testContainer.Stop(); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func setupTestDB(t *testing.T) *gorm.DB {
	db := container.SetUpDatabaseConnection()

	// Migrate the User schema
	err := db.AutoMigrate(&entity.User{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	// Drop the User table
	err := db.Migrator().DropTable(&entity.User{})
	if err != nil {
		t.Fatalf("Failed to drop table: %v", err)
	}

	// Close the database connection
	if err := container.CloseDatabaseConnection(db); err != nil {
		t.Fatalf("Failed to close database connection: %v", err)
	}
}

func TestUser_Integration_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	tests := []struct {
		name        string
		user        *entity.User
		expectError bool
		validate    func(t *testing.T, user *entity.User, db *gorm.DB)
	}{
		{
			name: "Valid user creation",
			user: &entity.User{
				Name:       "John Doe",
				Email:      "john-doe@example.com",
				TelpNumber: "1234567890",
				Password:   "password123",
				Role:       "user",
				ImageUrl:   "https://example.com/image.jpg",
			},
			expectError: false,
			validate: func(t *testing.T, user *entity.User, db *gorm.DB) {
				// Verify the user was saved
				var savedUser entity.User
				err := db.Where("email = ?", user.Email).First(&savedUser).Error
				assert.NoError(t, err, "User should exist in the database")
				assert.NotEqual(t, uuid.Nil, savedUser.ID, "ID should be generated")
				assert.Equal(t, user.Name, savedUser.Name, "Name should match")
				assert.Equal(t, user.Email, savedUser.Email, "Email should match")
				assert.Equal(t, user.TelpNumber, savedUser.TelpNumber, "TelpNumber should match")
				assert.NotEqual(t, "password123", savedUser.Password, "Password should be hashed")
				assert.Equal(t, "user", savedUser.Role, "Role should be user")
				assert.False(t, savedUser.IsVerified, "IsVerified should be false")
			},
		},
		{
			name: "Duplicate email",
			user: &entity.User{
				Name:       "Jane Doe",
				Email:      "john@example.com",
				TelpNumber: "0987654321",
				Password:   "password123",
				Role:       "user",
			},
			expectError: true,
			validate: func(t *testing.T, user *entity.User, db *gorm.DB) {
				var count int64
				db.Model(&entity.User{}).Where("email = ?", user.Email).Count(&count)
				assert.Equal(t, int64(1), count, "Only one user with this email should exist")
			},
		},
		{
			name: "Invalid role",
			user: &entity.User{
				Name:     "Invalid User",
				Email:    "invalid@example.com",
				Password: "password123",
				Role:     "invalid_role",
			},
			expectError: true,
		},
	}

	// Create a user with the same email for the duplicate email test
	db.Create(
		&entity.User{
			Name:       "Existing User",
			Email:      "john@example.com",
			TelpNumber: "1234567890",
			Password:   "password123",
			Role:       "user",
		},
	)

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := db.Create(tt.user).Error
				if tt.expectError {
					assert.Error(t, err, "Expected an error")
				} else {
					assert.NoError(t, err, "Expected no error")
					tt.validate(t, tt.user, db)
				}
			},
		)
	}
}

func TestUser_Integration_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create a test user
	user := &entity.User{
		Name:       "John Doe",
		Email:      "john@example.com",
		TelpNumber: "1234567890",
		Password:   "password123",
		Role:       "user",
	}
	err := db.Create(user).Error
	assert.NoError(t, err, "Failed to create test user")

	tests := []struct {
		name        string
		update      func(user *entity.User)
		expectError bool
		validate    func(t *testing.T, user *entity.User, db *gorm.DB)
	}{
		{
			name: "Update password and name",
			update: func(user *entity.User) {
				user.Name = "John Updated"
				user.Password = "newpassword123"
			},
			expectError: false,
			validate: func(t *testing.T, user *entity.User, db *gorm.DB) {
				var updatedUser entity.User
				err := db.Where("email = ?", user.Email).First(&updatedUser).Error
				assert.NoError(t, err, "User should exist in the database")
				assert.Equal(t, "John Updated", updatedUser.Name, "Name should be updated")
				assert.NotEqual(t, "newpassword123", updatedUser.Password, "Password should be hashed")
			},
		},
		{
			name: "Update without password change",
			update: func(user *entity.User) {
				user.TelpNumber = "0987654321"
				user.Role = "admin"
			},
			expectError: false,
			validate: func(t *testing.T, user *entity.User, db *gorm.DB) {
				var updatedUser entity.User
				err := db.Where("email = ?", user.Email).First(&updatedUser).Error
				assert.NoError(t, err, "User should exist in the database")
				assert.Equal(t, "0987654321", updatedUser.TelpNumber, "TelpNumber should be updated")
				assert.Equal(t, "admin", updatedUser.Role, "Role should be updated")
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.update(user)
				err := db.Save(user).Error
				if tt.expectError {
					assert.Error(t, err, "Expected an error")
				} else {
					assert.NoError(t, err, "Expected no error")
					tt.validate(t, user, db)
				}
			},
		)
	}
}
