package config

import (
	"fmt"
	"os"

	"github.com/Caknoooo/golang-clean_template/constants"
	"github.com/Caknoooo/golang-clean_template/entities"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetUpDatabaseConnection() *gorm.DB{
	if os.Getenv("APP_ENV") != constants.ENUM_RUN_PRODUCTION {
		err := godotenv.Load(".env")
		if err != nil{
			fmt.Println(err)
			panic(err)
		}
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v TimeZone=Asia/Jakarta", dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
		// Menambahkan opsi berikut akan memungkinkan driver database
		// untuk mendukung tipe data UUID secara bawaan.
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil{
		fmt.Println(err)
		panic(err)
	}

	if err := db.AutoMigrate(
		entities.User{},
	); err != nil{
		fmt.Println(err)
		panic(err)
	}
	
	return db
}

func ClosDatabaseConnection(db *gorm.DB){
	dbSQL, err := db.DB()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	dbSQL.Close()
}