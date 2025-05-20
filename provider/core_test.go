package provider

import (
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/config"
	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/Caknoooo/go-gin-clean-starter/service"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockConfig struct {
	mock.Mock
}

func (m *mockConfig) SetUpDatabaseConnection() *gorm.DB {
	args := m.Called()
	db := args.Get(0).(*gorm.DB)
	if err := args.Error(1); err != nil {
		// Mimic the original function's behavior by panicking on error
		panic(err)
	}
	return db
}

type mockUserProvider struct {
	mock.Mock
}

func (m *mockUserProvider) ProvideUserDependencies(injector *do.Injector) {
	m.Called(injector)
}

func TestInitDatabase(t *testing.T) {

	injector := do.New()

	mockCfg := &mockConfig{}
	mockDB := &gorm.DB{}
	mockCfg.On("SetUpDatabaseConnection").Return(mockDB, nil)
	originalSetUp := config.SetUpDatabaseConnection
	config.SetUpDatabaseConnection = mockCfg.SetUpDatabaseConnection
	defer func() { config.SetUpDatabaseConnection = originalSetUp }()

	InitDatabase(injector)

	db, err := do.InvokeNamed[*gorm.DB](injector, constants.DB)
	assert.NoError(t, err, "should provide DB without error")
	assert.Equal(t, mockDB, db, "should provide the mock DB")
	mockCfg.AssertExpectations(t)
}

func TestRegisterDependencies(t *testing.T) {
	injector := do.New()

	mockCfg := &mockConfig{}
	mockDB := &gorm.DB{}
	mockCfg.On("SetUpDatabaseConnection").Return(mockDB, nil)
	originalSetUp := config.SetUpDatabaseConnection
	config.SetUpDatabaseConnection = mockCfg.SetUpDatabaseConnection
	defer func() { config.SetUpDatabaseConnection = originalSetUp }()

	mockUserProv := &mockUserProvider{}
	mockUserProv.On("ProvideUserDependencies", injector).Return()
	originalProvide := ProvideUserDependencies
	ProvideUserDependencies = mockUserProv.ProvideUserDependencies
	defer func() { ProvideUserDependencies = originalProvide }()

	RegisterDependencies(injector)

	db, err := do.InvokeNamed[*gorm.DB](injector, constants.DB)
	assert.NoError(t, err, "should provide DB without error")
	assert.Equal(t, mockDB, db, "should provide the mock DB")

	jwtService, err := do.InvokeNamed[service.JWTService](injector, constants.JWTService)
	assert.NoError(t, err, "should provide JWTService without error")
	assert.NotNil(t, jwtService, "JWTService should not be nil")

	mockUserProv.AssertExpectations(t)
	mockCfg.AssertExpectations(t)
}
