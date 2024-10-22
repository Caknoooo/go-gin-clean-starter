package command

import (
	"log"
	"os"
	"strings"

	"github.com/Caknoooo/go-gin-clean-starter/migrations"
	"github.com/Caknoooo/go-gin-clean-starter/script"
	"gorm.io/gorm"
)

func Commands(db *gorm.DB) bool {
	var scriptName string

	migrate := false
	seed := false
	run := false
	scriptFlag := false

	for _, arg := range os.Args[1:] {
		if arg == "--migrate" {
			migrate = true
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

	if migrate {
		if err := migrations.Migrate(db); err != nil {
			log.Fatalf("error migration: %v", err)
		}
		log.Println("migration completed successfully")
	}

	if seed {
		if err := migrations.Seeder(db); err != nil {
			log.Fatalf("error migration seeder: %v", err)
		}
		log.Println("seeder completed successfully")
	}

	if scriptFlag {
		if err := script.Script(scriptName, db); err != nil {
			log.Fatalf("error script: %v", err)
		}
		log.Println("script run successfully")
	}

	if run {
		return true
	}

	return false
}
