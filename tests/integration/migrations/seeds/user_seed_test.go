package seeds

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/helpers"
	"github.com/Caknoooo/go-gin-clean-starter/migrations/seeds"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type SeedsTestSuite struct {
	suite.Suite
	db           *gorm.DB
	testData     []SeedUserRequest
	tempJSONPath string
	projectRoot  string
	usedTestJSON bool // Tracks if users_test.json was created
}

type SeedUserRequest struct {
	Name       string `json:"name"`
	TelpNumber string `json:"telp_number"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`
}

func (suite *SeedsTestSuite) SetupSuite() {
	// Setup test database
	testContainer, err := container.StartTestContainer()
	if err != nil {
		suite.T().Fatalf("Failed to start test container: %v", err)
	}

	// Set environment variables for database connection
	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_PORT", testContainer.Port)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")

	// Setup database connection
	db := container.SetUpDatabaseConnection()
	suite.db = db

	// Get project root
	projectRoot, err := helpers.GetProjectRoot()
	if err != nil {
		suite.T().Fatalf("Failed to get project root: %v", err)
	}
	suite.projectRoot = projectRoot

	// Create test JSON file in a temporary directory
	suite.tempJSONPath = filepath.Join(os.TempDir(), "test_users.json")
	suite.testData = []SeedUserRequest{
		{
			Name:       "Test User 1",
			TelpNumber: "08123456789",
			Email:      "test1@example.com",
			Password:   "password123",
			Role:       "user",
			IsVerified: true,
		},
		{
			Name:       "Test Admin",
			TelpNumber: "08123456788",
			Email:      "admin@example.com",
			Password:   "admin123",
			Role:       "admin",
			IsVerified: true,
		},
	}

	err = createTestJSONFile(suite.tempJSONPath, suite.testData)
	if err != nil {
		suite.T().Fatalf("Failed to create test JSON file: %v", err)
	}
}

func (suite *SeedsTestSuite) TearDownSuite() {
	// Clean up database connection
	if suite.db != nil {
		if err := container.CloseDatabaseConnection(suite.db); err != nil {
			suite.T().Logf("Failed to close database connection: %v", err)
		}
	}

	// Remove temporary test JSON file
	os.Remove(suite.tempJSONPath)
}

func (suite *SeedsTestSuite) BeforeTest(suiteName, testName string) {
	// Ensure clean state for each test
	suite.db.Migrator().DropTable(&entity.User{})
	suite.usedTestJSON = false // Reset for each test
}

func (suite *SeedsTestSuite) setupTestJSON() (string, error) {
	// Create migrations/json directory if it doesn't exist
	testSeedDir := filepath.Join(suite.projectRoot, "migrations", "json")
	err := os.MkdirAll(testSeedDir, 0755)
	if err != nil {
		return "", err
	}

	// Check if users.json exists
	usersJSONPath := filepath.Join(testSeedDir, "users.json")
	if _, err := os.Stat(usersJSONPath); err == nil {
		// users.json exists, use it directly
		return usersJSONPath, nil
	}

	// users.json doesn't exist, create users_test.json
	testJSONPath := filepath.Join(testSeedDir, "users_test.json")
	err = copyFile(suite.tempJSONPath, testJSONPath)
	if err != nil {
		return "", err
	}
	suite.usedTestJSON = true
	return testJSONPath, nil
}

func (suite *SeedsTestSuite) cleanupTestJSON(jsonPath string) {
	if suite.usedTestJSON {
		os.Remove(jsonPath) // Only remove users_test.json
	}
	// Never remove users.json or migrations/json directory
}

func (suite *SeedsTestSuite) TestListUserSeeder_Success() {
	// Setup JSON file (either users.json or users_test.json)
	jsonPath, err := suite.setupTestJSON()
	if err != nil {
		suite.T().Fatalf("Failed to setup test JSON: %v", err)
	}
	defer suite.cleanupTestJSON(jsonPath)

	// Temporarily override GetProjectRoot to point to the directory containing the JSON
	oldGetProjectRoot := helpers.GetProjectRoot
	helpers.GetProjectRoot = func() (string, error) {
		return suite.projectRoot, nil
	}
	defer func() { helpers.GetProjectRoot = oldGetProjectRoot }()

	// Execute the seeder
	err = seeds.ListUserSeeder(suite.db)
	assert.NoError(suite.T(), err, "Seeder should not return error")

	// Read the JSON file to verify expected data
	var seededData []SeedUserRequest
	data, err := os.ReadFile(jsonPath)
	assert.NoError(suite.T(), err, "Should read JSON file")
	err = json.Unmarshal(data, &seededData)
	assert.NoError(suite.T(), err, "Should parse JSON file")

	// Verify data was inserted
	var users []entity.User
	result := suite.db.Find(&users)
	assert.NoError(suite.T(), result.Error, "Should be able to query users")
	assert.Equal(suite.T(), len(seededData), int(result.RowsAffected), "Should insert all test users")

	for _, testUser := range seededData {
		var user entity.User
		err := suite.db.Where("email = ?", testUser.Email).First(&user).Error
		assert.NoError(suite.T(), err, "Should find seeded user")
		assert.Equal(suite.T(), testUser.Name, user.Name, "User name should match")
		assert.Equal(suite.T(), testUser.Role, user.Role, "User role should match")
		assert.Equal(suite.T(), testUser.IsVerified, user.IsVerified, "User verification status should match")
	}
}

func (suite *SeedsTestSuite) TestListUserSeeder_TableCreation() {
	// Setup JSON file
	jsonPath, err := suite.setupTestJSON()
	if err != nil {
		suite.T().Fatalf("Failed to setup test JSON: %v", err)
	}
	defer suite.cleanupTestJSON(jsonPath)

	// Temporarily override GetProjectRoot
	oldGetProjectRoot := helpers.GetProjectRoot
	helpers.GetProjectRoot = func() (string, error) {
		return suite.projectRoot, nil
	}
	defer func() { helpers.GetProjectRoot = oldGetProjectRoot }()

	// Ensure table doesn't exist
	suite.db.Migrator().DropTable(&entity.User{})

	// Execute the seeder
	err = seeds.ListUserSeeder(suite.db)
	assert.NoError(suite.T(), err, "Seeder should not return error")

	// Verify table was created
	hasTable := suite.db.Migrator().HasTable(&entity.User{})
	assert.True(suite.T(), hasTable, "Seeder should create table if it doesn't exist")
}

func (suite *SeedsTestSuite) TestListUserSeeder_DuplicateUsers() {
	// Setup JSON file
	jsonPath, err := suite.setupTestJSON()
	if err != nil {
		suite.T().Fatalf("Failed to setup test JSON: %v", err)
	}
	defer suite.cleanupTestJSON(jsonPath)

	// Temporarily override GetProjectRoot
	oldGetProjectRoot := helpers.GetProjectRoot
	helpers.GetProjectRoot = func() (string, error) {
		return suite.projectRoot, nil
	}
	defer func() { helpers.GetProjectRoot = oldGetProjectRoot }()

	// First run - should insert users
	err = seeds.ListUserSeeder(suite.db)
	assert.NoError(suite.T(), err, "First seeder run should not return error")

	// Get initial count
	var initialCount int64
	suite.db.Model(&entity.User{}).Count(&initialCount)

	// Second run - should not insert duplicates
	err = seeds.ListUserSeeder(suite.db)
	assert.NoError(suite.T(), err, "Second seeder run should not return error")

	// Get new count
	var newCount int64
	suite.db.Model(&entity.User{}).Count(&newCount)

	assert.Equal(suite.T(), initialCount, newCount, "Should not insert duplicate users")
}

func (suite *SeedsTestSuite) TestListUserSeeder_InvalidJSONPath() {
	// Temporarily modify GetProjectRoot to return invalid path
	oldGetProjectRoot := helpers.GetProjectRoot
	defer func() { helpers.GetProjectRoot = oldGetProjectRoot }()

	helpers.GetProjectRoot = func() (string, error) {
		return filepath.Join(os.TempDir(), "nonexistent_project"), nil
	}

	err := seeds.ListUserSeeder(suite.db)
	assert.Error(suite.T(), err, "Should return error for invalid JSON path")
}

func (suite *SeedsTestSuite) TestListUserSeeder_InvalidJSONContent() {
	// Create a temporary directory to act as a fake project root
	tempDir := suite.T().TempDir()
	testSeedDir := filepath.Join(tempDir, "migrations", "json")
	err := os.MkdirAll(testSeedDir, 0755)
	if err != nil {
		suite.T().Fatalf("Failed to create test seed directory: %v", err)
	}

	// Create invalid JSON file as users.json in the temporary directory
	invalidJSONPath := filepath.Join(testSeedDir, "users.json")
	err = os.WriteFile(invalidJSONPath, []byte("invalid json content"), 0644)
	if err != nil {
		suite.T().Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Temporarily override GetProjectRoot to point to tempDir
	oldGetProjectRoot := helpers.GetProjectRoot
	helpers.GetProjectRoot = func() (string, error) {
		return tempDir, nil
	}
	defer func() { helpers.GetProjectRoot = oldGetProjectRoot }()

	// Run the seeder, which should read the invalid users.json
	err = seeds.ListUserSeeder(suite.db)
	assert.Error(suite.T(), err, "Should return error for invalid JSON content")
}

func TestSeedsTestSuite(t *testing.T) {
	suite.Run(t, new(SeedsTestSuite))
}

// Helper function to create test JSON file
func createTestJSONFile(path string, data []SeedUserRequest) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}
