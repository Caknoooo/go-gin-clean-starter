package script

import (
	_ "github.com/Caknoooo/go-gin-clean-starter/database/migrations"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Caknoooo/go-gin-clean-starter/database"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func Commands(injector *do.Injector) bool {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)

	var scriptName string
	var migrationName string
	var rollbackBatch int

	migrateRun := false
	migrateRollback := false
	migrateRollbackAll := false
	migrateStatus := false
	migrateCreate := false
	seed := false
	run := false
	scriptFlag := false

	for i, arg := range os.Args[1:] {
		if arg == "--migrate" || arg == "--migrate:run" {
			migrateRun = true
		}
		if arg == "--migrate:rollback" {
			migrateRollback = true
			if i+2 < len(os.Args) && !strings.HasPrefix(os.Args[i+2], "--") {
				batch, err := strconv.Atoi(os.Args[i+2])
				if err == nil {
					rollbackBatch = batch
				}
			}
		}
		if arg == "--migrate:rollback:all" {
			migrateRollbackAll = true
		}
		if arg == "--migrate:status" {
			migrateStatus = true
		}
		if strings.HasPrefix(arg, "--migrate:create:") {
			migrateCreate = true
			migrationName = strings.TrimPrefix(arg, "--migrate:create:")
		}
		if arg == "--seed" {
			seed = true
		}
		if arg == "--run" {
			run = true
		}
		if strings.HasPrefix(arg, "--script:") {
			scriptFlag = true
			scriptName = strings.TrimPrefix(arg, "--script:")
		}
	}

	if migrateRun {
		if err := database.Migrate(db); err != nil {
			log.Fatalf("error migration: %v", err)
		}
		log.Println("migration completed successfully")
	}

	if migrateRollback {
		manager := database.NewMigrationManager(db)
		if rollbackBatch > 0 {
			if err := manager.Rollback(rollbackBatch); err != nil {
				log.Fatalf("error rollback migration batch %d: %v", rollbackBatch, err)
			}
		} else {
			if err := manager.Rollback(0); err != nil {
				log.Fatalf("error rollback migration: %v", err)
			}
		}
		log.Println("rollback completed successfully")
	}

	if migrateRollbackAll {
		manager := database.NewMigrationManager(db)
		if err := manager.RollbackAll(); err != nil {
			log.Fatalf("error rollback all migrations: %v", err)
		}
		log.Println("rollback all completed successfully")
	}

	if migrateStatus {
		manager := database.NewMigrationManager(db)
		if err := manager.Status(); err != nil {
			log.Fatalf("error getting migration status: %v", err)
		}
	}

	if migrateCreate {
		if migrationName == "" {
			log.Fatalf("migration name is required")
		}
		manager := database.NewMigrationManager(db)
		if err := manager.Create(migrationName); err != nil {
			log.Fatalf("error creating migration: %v", err)
		}
		log.Println("migration file created successfully")
	}

	if seed {
		if err := database.Seeder(db); err != nil {
			log.Fatalf("error migration seeder: %v", err)
		}
		log.Println("seeder completed successfully")
	}

	if scriptFlag {
		if err := Script(scriptName, db); err != nil {
			log.Fatalf("error script: %v", err)
		}
		log.Println("script run successfully")
	}

	if run {
		return true
	}

	if migrateRun || migrateRollback || migrateRollbackAll || migrateStatus || migrateCreate {
		return false
	}

	return false
}
