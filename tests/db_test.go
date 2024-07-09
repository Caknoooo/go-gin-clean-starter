package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetUpDatabaseConnection() *gorm.DB {
	var (
		dbUser, dbPass, dbHost, dbName, dbPort string
		getenv                                 = os.Getenv
		godotenv                               = godotenv.Load
	)

	if getenv("APP_ENV") != constants.ENUM_RUN_PRODUCTION {
		err := godotenv("../.env")
		if err != nil {
			panic("Error loading .env file: " + err.Error())
		}
	}

	dbUser = getenv("DB_USER")
	dbPass = getenv("DB_PASS")
	dbHost = getenv("DB_HOST")
	dbName = getenv("DB_NAME")
	dbPort = getenv("DB_PORT")

	if dbUser == "" || dbPass == "" || dbHost == "" || dbName == "" || dbPort == "" {
		panic("Missing required environment variables")
	}

	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v TimeZone=Asia/Jakarta", dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	return db
}

func Test_DBConnection(t *testing.T) {
	db := SetUpDatabaseConnection()
	assert.NoError(t, db.Error, "Expected no error during database connection")
	assert.NotNil(t, db, "Expected a non-nil database connection")
}
