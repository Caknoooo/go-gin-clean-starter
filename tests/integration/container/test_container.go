package container

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestDatabaseContainer struct {
	testcontainers.Container
	Host string
	Port string
}

func StartTestContainer() (*TestDatabaseContainer, error) {
	ctx := context.Background()

	// Set up the container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	// Start the container
	container, err := testcontainers.GenericContainer(
		ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Get the container host
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	return &TestDatabaseContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Port(),
	}, nil
}

func (c *TestDatabaseContainer) Stop() error {
	ctx := context.Background()
	timeout := 10 * time.Second
	return c.Container.Stop(ctx, &timeout)
}

// CloseDatabaseConnection closes the GORM database connection.
func CloseDatabaseConnection(db *gorm.DB) error {
	dbSQL, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}
	return dbSQL.Close()
}

// SetUpDatabaseConnection establishes a GORM database connection to the test container.
func SetUpDatabaseConnection() *gorm.DB {
	// Construct the DSN using environment variables set by the test suite
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Open GORM connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	// Enable UUID extension
	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		panic(fmt.Errorf("failed to enable uuid-ossp extension: %w", err))
	}

	return db
}
