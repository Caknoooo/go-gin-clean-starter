package config

import (
	"fmt"
	"os"

	"github.com/Caknoooo/go-gin-clean-starter/constants"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunExtension(db *gorm.DB) {
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
}

func loadEnv() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = constants.ENUM_RUN_DEVELOPMENT // Assuming this constant exists
	}

	var envFile string
	switch appEnv {
	case constants.ENUM_RUN_TESTING:
		envFile = ".env.test"
	case constants.ENUM_RUN_PRODUCTION:
		envFile = ".env.prod"
	default:
		envFile = ".env"
	}

	if err := godotenv.Load(envFile); err != nil {
		// Don't panic if .env file is not found in production (use actual env vars)
		if !os.IsNotExist(err) || appEnv != constants.ENUM_RUN_PRODUCTION {
			panic(fmt.Errorf("failed to load %s file: %w", envFile, err))
		}
	}
}

func SetUpDatabaseConnection() *gorm.DB {
	if os.Getenv("APP_ENV") != constants.ENUM_RUN_PRODUCTION {
		loadEnv()
	}

	config := DatabaseConfig{
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASS", ""),
		Host:     getEnv("DB_HOST", "localhost"),
		Name:     getEnv("DB_NAME", ""),
		Port:     getEnv("DB_PORT", "5432"),
	}

	if config.User == "" || config.Password == "" || config.Name == "" {
		panic("database configuration is incomplete")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		config.Host, config.User, config.Password, config.Name, config.Port,
	)

	db, err := gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true,
			},
		), &gorm.Config{
			Logger: SetupLogger(),
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	RunExtension(db)

	return db
}

func CloseDatabaseConnection(db *gorm.DB) {
	dbSQL, err := db.DB()
	if err != nil {
		panic(fmt.Errorf("failed to get database instance: %w", err))
	}
	if err := dbSQL.Close(); err != nil {
		panic(fmt.Errorf("failed to close database connection: %w", err))
	}
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Name     string
	Port     string
}

// getEnv is a helper function to get environment variables with fallback
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
