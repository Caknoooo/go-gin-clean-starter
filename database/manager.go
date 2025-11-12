package database

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

type MigrationManager struct {
	db         *gorm.DB
	migrations []MigrationFile
}

type MigrationFile struct {
	Name     string
	Path     string
	UpFunc   func(*gorm.DB) error
	DownFunc func(*gorm.DB) error
}

type MigrationInterface interface {
	Up(*gorm.DB) error
	Down(*gorm.DB) error
}

var registeredMigrations []MigrationFile

func RegisterMigration(name string, upFunc func(*gorm.DB) error, downFunc func(*gorm.DB) error) {
	registeredMigrations = append(registeredMigrations, MigrationFile{
		Name:     name,
		UpFunc:   upFunc,
		DownFunc: downFunc,
	})
}

func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: registeredMigrations,
	}
}

func (mm *MigrationManager) ensureMigrationsTable() error {
	return mm.db.AutoMigrate(&entities.Migration{})
}

func (mm *MigrationManager) getLastBatch() (int, error) {
	var lastBatch int
	err := mm.db.Model(&entities.Migration{}).
		Select("COALESCE(MAX(batch), 0)").
		Scan(&lastBatch).Error
	return lastBatch, err
}

func (mm *MigrationManager) getMigrationsByBatch(batch int) ([]entities.Migration, error) {
	var migrations []entities.Migration
	err := mm.db.Where("batch = ?", batch).Order("id ASC").Find(&migrations).Error
	return migrations, err
}

func (mm *MigrationManager) getRanMigrations() ([]entities.Migration, error) {
	var migrations []entities.Migration
	err := mm.db.Order("batch ASC, id ASC").Find(&migrations).Error
	return migrations, err
}

func (mm *MigrationManager) isMigrationRan(name string) (bool, error) {
	var count int64
	err := mm.db.Model(&entities.Migration{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (mm *MigrationManager) recordMigration(name string, batch int) error {
	migration := entities.Migration{
		Name:  name,
		Batch: batch,
	}
	return mm.db.Create(&migration).Error
}

func (mm *MigrationManager) deleteMigration(name string) error {
	return mm.db.Where("name = ?", name).Delete(&entities.Migration{}).Error
}

func (mm *MigrationManager) Run() error {
	if err := mm.ensureMigrationsTable(); err != nil {
		return err
	}

	lastBatch, err := mm.getLastBatch()
	if err != nil {
		return err
	}

	newBatch := lastBatch + 1
	ranCount := 0

	for _, migration := range mm.migrations {
		ran, err := mm.isMigrationRan(migration.Name)
		if err != nil {
			return err
		}

		if !ran {
			if err := migration.UpFunc(mm.db); err != nil {
				return fmt.Errorf("error running migration %s: %v", migration.Name, err)
			}

			if err := mm.recordMigration(migration.Name, newBatch); err != nil {
				return fmt.Errorf("error recording migration %s: %v", migration.Name, err)
			}

			ranCount++
			fmt.Printf("Migration %s executed successfully\n", migration.Name)
		}
	}

	if ranCount == 0 {
		fmt.Println("No new migrations to run")
	} else {
		fmt.Printf("Ran %d migration(s)\n", ranCount)
	}

	return nil
}

func (mm *MigrationManager) Rollback(batch int) error {
	if err := mm.ensureMigrationsTable(); err != nil {
		return err
	}

	var migrationsToRollback []entities.Migration
	var err error

	if batch > 0 {
		migrationsToRollback, err = mm.getMigrationsByBatch(batch)
		if err != nil {
			return err
		}
		if len(migrationsToRollback) == 0 {
			return fmt.Errorf("no migrations found for batch %d", batch)
		}
	} else {
		lastBatch, err := mm.getLastBatch()
		if err != nil {
			return err
		}
		if lastBatch == 0 {
			return fmt.Errorf("no migrations to rollback")
		}
		migrationsToRollback, err = mm.getMigrationsByBatch(lastBatch)
		if err != nil {
			return err
		}
	}

	sort.Slice(migrationsToRollback, func(i, j int) bool {
		return migrationsToRollback[i].ID > migrationsToRollback[j].ID
	})

	rolledBackCount := 0
	for _, migrationRecord := range migrationsToRollback {
		var migrationFile *MigrationFile
		for _, m := range mm.migrations {
			if m.Name == migrationRecord.Name {
				migrationFile = &m
				break
			}
		}

		if migrationFile == nil {
			fmt.Printf("Warning: Migration file for %s not found, skipping\n", migrationRecord.Name)
			mm.deleteMigration(migrationRecord.Name)
			continue
		}

		if err := migrationFile.DownFunc(mm.db); err != nil {
			return fmt.Errorf("error rolling back migration %s: %v", migrationRecord.Name, err)
		}

		if err := mm.deleteMigration(migrationRecord.Name); err != nil {
			return fmt.Errorf("error deleting migration record %s: %v", migrationRecord.Name, err)
		}

		rolledBackCount++
		fmt.Printf("Migration %s rolled back successfully\n", migrationRecord.Name)
	}

	fmt.Printf("Rolled back %d migration(s)\n", rolledBackCount)
	return nil
}

func (mm *MigrationManager) RollbackAll() error {
	if err := mm.ensureMigrationsTable(); err != nil {
		return err
	}

	ranMigrations, err := mm.getRanMigrations()
	if err != nil {
		return err
	}

	if len(ranMigrations) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	sort.Slice(ranMigrations, func(i, j int) bool {
		if ranMigrations[i].Batch != ranMigrations[j].Batch {
			return ranMigrations[i].Batch > ranMigrations[j].Batch
		}
		return ranMigrations[i].ID > ranMigrations[j].ID
	})

	rolledBackCount := 0
	for _, migrationRecord := range ranMigrations {
		var migrationFile *MigrationFile
		for _, m := range mm.migrations {
			if m.Name == migrationRecord.Name {
				migrationFile = &m
				break
			}
		}

		if migrationFile == nil {
			fmt.Printf("Warning: Migration file for %s not found, skipping\n", migrationRecord.Name)
			mm.deleteMigration(migrationRecord.Name)
			continue
		}

		if err := migrationFile.DownFunc(mm.db); err != nil {
			return fmt.Errorf("error rolling back migration %s: %v", migrationRecord.Name, err)
		}

		if err := mm.deleteMigration(migrationRecord.Name); err != nil {
			return fmt.Errorf("error deleting migration record %s: %v", migrationRecord.Name, err)
		}

		rolledBackCount++
		fmt.Printf("Migration %s rolled back successfully\n", migrationRecord.Name)
	}

	fmt.Printf("Rolled back %d migration(s)\n", rolledBackCount)
	return nil
}

func (mm *MigrationManager) Status() error {
	if err := mm.ensureMigrationsTable(); err != nil {
		return err
	}

	ranMigrations, err := mm.getRanMigrations()
	if err != nil {
		return err
	}

	fmt.Println("\nMigration Status:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-50s %-10s %-20s\n", "Migration", "Batch", "Ran At")
	fmt.Println(strings.Repeat("-", 80))

	if len(ranMigrations) == 0 {
		fmt.Println("No migrations have been run")
	} else {
		for _, m := range ranMigrations {
			fmt.Printf("%-50s %-10d %-20s\n", m.Name, m.Batch, m.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d migration(s)\n", len(ranMigrations))

	pendingCount := len(mm.migrations) - len(ranMigrations)
	if pendingCount > 0 {
		fmt.Printf("Pending: %d migration(s)\n", pendingCount)
	}

	return nil
}

func (mm *MigrationManager) Create(name string) error {
	timestamp := time.Now().Format("20060102150405")
	normalizedName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	fileName := fmt.Sprintf("%s_%s.go", timestamp, normalizedName)

	migrationsDir := "database/migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("error creating migrations directory: %v", err)
	}

	filePath := filepath.Join(migrationsDir, fileName)

	funcName := strings.ReplaceAll(strings.Title(strings.ReplaceAll(name, "_", " ")), " ", "")
	migrationName := fmt.Sprintf("%s_%s", timestamp, normalizedName)

	var entityName string
	var entityFileName string
	var entityCreated bool
	var migrationTemplate string

	if strings.HasPrefix(normalizedName, "create_") && strings.HasSuffix(normalizedName, "_table") {
		tableName := strings.TrimPrefix(normalizedName, "create_")
		tableName = strings.TrimSuffix(tableName, "_table")

		entityName = strings.Title(strings.ReplaceAll(tableName, "_", " "))
		entityName = strings.ReplaceAll(entityName, " ", "")

		entityFileName = fmt.Sprintf("%s_entity.go", strings.ToLower(tableName))
		entityPath := filepath.Join("database/entities", entityFileName)

		if _, err := os.Stat(entityPath); os.IsNotExist(err) {
			receiverName := strings.ToLower(string(entityName[0]))
			entityTemplate := fmt.Sprintf(`package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type %s struct {
	ID uuid.UUID `+"`gorm:\"type:uuid;primary_key;default:uuid_generate_v4()\" json:\"id\"`"+`

	Timestamp
}

func (%s *%s) BeforeCreate(tx *gorm.DB) (err error) {
	if %s.ID == uuid.Nil {
		%s.ID = uuid.New()
	}
	return nil
}
`, entityName, receiverName, entityName, receiverName, receiverName)

			if err := ioutil.WriteFile(entityPath, []byte(entityTemplate), 0644); err != nil {
				return fmt.Errorf("error creating entity file: %v", err)
			}

			entityCreated = true
			fmt.Printf("Entity file created: %s\n", entityPath)

			if err := mm.addEntityToMigrationFile(entityName); err != nil {
				fmt.Printf("Warning: Failed to add entity to migration.go: %v\n", err)
			}
		} else {
			fmt.Printf("Entity file already exists: %s\n", entityPath)
		}

		migrationTemplate = fmt.Sprintf(`package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("%s", Up%s, Down%s)
}

func Up%s(db *gorm.DB) error {
	return db.AutoMigrate(&entities.%s{})
}

func Down%s(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.%s{})
}
`, migrationName, funcName, funcName, funcName, entityName, funcName, entityName)
	} else {
		migrationTemplate = fmt.Sprintf(`package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("%s", Up%s, Down%s)
}

func Up%s(db *gorm.DB) error {
	return nil
}

func Down%s(db *gorm.DB) error {
	return nil
}
`, migrationName, funcName, funcName, funcName, funcName)
	}

	if err := ioutil.WriteFile(filePath, []byte(migrationTemplate), 0644); err != nil {
		return fmt.Errorf("error creating migration file: %v", err)
	}

	fmt.Printf("Migration file created: %s\n", filePath)
	if entityCreated {
		fmt.Printf("Entity %s has been created and added to migration\n", entityName)
	}
	return nil
}

func (mm *MigrationManager) addEntityToMigrationFile(entityName string) error {
	migrationFilePath := "database/migration.go"

	content, err := ioutil.ReadFile(migrationFilePath)
	if err != nil {
		return fmt.Errorf("error reading migration file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inAutoMigrate := false
	entityAdded := false

	for i, line := range lines {
		if strings.Contains(line, "db.AutoMigrate(") {
			inAutoMigrate = true
			newLines = append(newLines, line)
			continue
		}

		if inAutoMigrate {
			if strings.Contains(line, "&entities."+entityName) {
				entityAdded = true
				newLines = append(newLines, line)
				continue
			}

			if strings.Contains(line, ");") {
				if !entityAdded {
					lastEntityIdx := -1
					for j := i - 1; j >= 0; j-- {
						if strings.Contains(lines[j], "&entities.") {
							lastEntityIdx = j
							break
						}
					}

					if lastEntityIdx >= 0 {
						indent := ""
						for _, char := range lines[lastEntityIdx] {
							if char == '\t' {
								indent += "\t"
							} else {
								break
							}
						}
						entityLine := fmt.Sprintf("%s\t&entities.%s{},", indent, entityName)
						newLines = append(newLines, entityLine)
					} else {
						indent := "\t\t"
						entityLine := fmt.Sprintf("%s&entities.%s{},", indent, entityName)
						newLines = append(newLines, entityLine)
					}
					entityAdded = true
				}
				inAutoMigrate = false
				newLines = append(newLines, line)
				continue
			}
		}

		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	if err := ioutil.WriteFile(migrationFilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("error writing migration file: %v", err)
	}

	return nil
}
