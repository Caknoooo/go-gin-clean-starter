package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestScript(t *testing.T) {
	tests := []struct {
		name       string
		scriptName string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "example script - success",
			scriptName: "example_script",
			wantErr:    false,
		},
		{
			name:       "unknown script - error",
			scriptName: "unknown_script",
			wantErr:    true,
			errMsg:     "script not found",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				mockDB := new(MockDB)
				mockGormDB := &gorm.DB{}
				mockDB.On("GormDB").Return(mockGormDB)

				err := Script(tt.scriptName, mockDB.GormDB())

				if tt.wantErr {
					assert.Error(t, err)
					assert.Equal(t, tt.errMsg, err.Error())
				} else {
					assert.NoError(t, err)
				}
			},
		)
	}
}
