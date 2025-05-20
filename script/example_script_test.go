package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) GormDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func TestNewExampleScript(t *testing.T) {
	mockDB := new(MockDB)
	mockGormDB := &gorm.DB{}
	mockDB.On("GormDB").Return(mockGormDB)

	script := NewExampleScript(mockDB.GormDB())

	assert.NotNil(t, script)
	assert.Equal(t, mockGormDB, script.db)
	mockDB.AssertExpectations(t)
}

func TestExampleScript_Run(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful run",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				mockDB := new(MockDB)
				mockGormDB := &gorm.DB{}
				mockDB.On("GormDB").Return(mockGormDB)

				script := NewExampleScript(mockDB.GormDB())

				err := script.Run()

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			},
		)
	}
}
