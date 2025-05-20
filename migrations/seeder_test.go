package migrations

import (
	"errors"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/migrations/seeds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockSeeder is a mock for the seeder functions
type MockSeeder struct {
	mock.Mock
}

func (m *MockSeeder) ListUserSeeder(db *gorm.DB) error {
	args := m.Called(db)
	return args.Error(0)
}

func TestSeeder(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {

			mockSeeder := new(MockSeeder)

			mockSeeder.On("ListUserSeeder", mock.AnythingOfType("*gorm.DB")).Return(nil)

			seeds.ListUserSeeder = mockSeeder.ListUserSeeder

			err := Seeder(nil)

			assert.NoError(t, err)
			mockSeeder.AssertExpectations(t)
		},
	)

	t.Run(
		"Error", func(t *testing.T) {

			mockSeeder := new(MockSeeder)

			expectedErr := errors.New("seeder error")
			mockSeeder.On("ListUserSeeder", mock.AnythingOfType("*gorm.DB")).Return(expectedErr)

			seeds.ListUserSeeder = mockSeeder.ListUserSeeder

			err := Seeder(nil)

			assert.Error(t, err)
			assert.Equal(t, expectedErr, err)
			mockSeeder.AssertExpectations(t)
		},
	)
}
