package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/migrations"
	"os"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/entity"
	"github.com/Caknoooo/go-gin-clean-starter/tests/integration/container"
)

func TestMigrate_Integration(t *testing.T) {
	// Start test container
	testContainer, err := container.StartTestContainer()
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}
	defer func() {
		if err := testContainer.Stop(); err != nil {
			t.Errorf("Failed to stop test container: %v", err)
		}
	}()

	// Set environment variables for database connection
	os.Setenv("DB_HOST", testContainer.Host)
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpassword")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_PORT", testContainer.Port)
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASS")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_PORT")
	}()

	// Set up database connection
	db := container.SetUpDatabaseConnection()
	defer func() {
		if err := container.CloseDatabaseConnection(db); err != nil {
			t.Errorf("Failed to close database connection: %v", err)
		}
	}()

	// Run the migration
	err = migrations.Migrate(db)
	if err != nil {
		t.Errorf("Migrate() returned an error: %v", err)
	}

	// Verify that the tables were created
	var tableCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'refresh_tokens')").Scan(&tableCount)
	if tableCount != 2 {
		t.Errorf("Expected 2 tables to be created, but found %d", tableCount)
	}

	// Verify that User and RefreshToken tables exist and can be queried
	var user entity.User
	var refreshToken entity.RefreshToken

	if err := db.Model(&user).Limit(1).Find(&user).Error; err != nil {
		t.Errorf("Failed to query User table: %v", err)
	}

	if err := db.Model(&refreshToken).Limit(1).Find(&refreshToken).Error; err != nil {
		t.Errorf("Failed to query RefreshToken table: %v", err)
	}
}
