package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/repository"

	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRefreshTokenRepository(t *testing.T) {

	testContainer, err := container.StartTestContainer()
	if err != nil {
		t.Fatalf("failed to start test container: %v", err)
	}
	defer func(testContainer *container.TestDatabaseContainer) {
		err := testContainer.Stop()
		if err != nil {
			panic(err)
		}
	}(testContainer)

	err = os.Setenv("DB_HOST", testContainer.Host)
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PASS", "testpassword")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("DB_PORT", testContainer.Port)
	if err != nil {
		panic(err)
	}

	db := container.SetUpDatabaseConnection()
	defer func(db *gorm.DB) {
		err := container.CloseDatabaseConnection(db)
		if err != nil {
			panic(err)
		}
	}(db)

	err = db.AutoMigrate(&entity.User{}, &entity.RefreshToken{})
	if err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	repo := repository.NewRefreshTokenRepository(db)

	user := entity.User{
		ID:         uuid.New(),
		Name:       "Test User",
		Email:      "test@example.com",
		TelpNumber: "1234567890",
		Password:   "password123",
		Role:       "user",
		Timestamp: entity.Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	err = db.Create(&user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	ctx := context.Background()

	t.Run(
		"Create", func(t *testing.T) {
			token := entity.RefreshToken{
				ID:        uuid.New(),
				Token:     "test-token",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(time.Hour),
				Timestamp: entity.Timestamp{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			createdToken, err := repo.Create(ctx, nil, token)
			assert.NoError(t, err)
			assert.Equal(t, token.Token, createdToken.Token)
			assert.Equal(t, token.UserID, createdToken.UserID)
			assert.Equal(t, token.ID, createdToken.ID)
		},
	)

	t.Run(
		"FindByToken", func(t *testing.T) {
			token := entity.RefreshToken{
				ID:        uuid.New(),
				Token:     "find-token",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(time.Hour),
				Timestamp: entity.Timestamp{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}
			_, err := repo.Create(ctx, nil, token)
			assert.NoError(t, err)

			foundToken, err := repo.FindByToken(ctx, nil, "find-token")
			assert.NoError(t, err)
			assert.Equal(t, token.Token, foundToken.Token)
			assert.Equal(t, token.UserID, foundToken.UserID)
			assert.Equal(t, token.ID, foundToken.ID)
			assert.NotNil(t, foundToken.User)
			assert.Equal(t, user.ID, foundToken.User.ID)
			assert.Equal(t, user.Email, foundToken.User.Email)
		},
	)

	t.Run(
		"DeleteByUserID", func(t *testing.T) {
			token := entity.RefreshToken{
				ID:        uuid.New(),
				Token:     "delete-user-token",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(time.Hour),
				Timestamp: entity.Timestamp{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}
			_, err := repo.Create(ctx, nil, token)
			assert.NoError(t, err)

			err = repo.DeleteByUserID(ctx, nil, user.ID.String())
			assert.NoError(t, err)

			_, err = repo.FindByToken(ctx, nil, "delete-user-token")
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		},
	)

	t.Run(
		"DeleteByToken", func(t *testing.T) {
			token := entity.RefreshToken{
				ID:        uuid.New(),
				Token:     "delete-token",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(time.Hour),
				Timestamp: entity.Timestamp{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}
			_, err := repo.Create(ctx, nil, token)
			assert.NoError(t, err)

			err = repo.DeleteByToken(ctx, nil, "delete-token")
			assert.NoError(t, err)

			_, err = repo.FindByToken(ctx, nil, "delete-token")
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		},
	)

	t.Run(
		"DeleteExpired", func(t *testing.T) {
			expiredToken := entity.RefreshToken{
				ID:        uuid.New(),
				Token:     "expired-token",
				UserID:    user.ID,
				ExpiresAt: time.Now().Add(-time.Hour),
				Timestamp: entity.Timestamp{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}
			_, err := repo.Create(ctx, nil, expiredToken)
			assert.NoError(t, err)

			err = repo.DeleteExpired(ctx, nil)
			assert.NoError(t, err)

			_, err = repo.FindByToken(ctx, nil, "expired-token")
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		},
	)
}
