package database

import (
	"fmt"
	"log"
	"petplate/internals/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error

	// Connection string for PostgreSQL
	dsn := "host=localhost user=postgres password=1234 dbname=petplate port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	// Opening the PostgreSQL database connection
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	} else {
		fmt.Println("Connection to PostgreSQL database: OK")
	}
	AutoMigrate()
}

func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.VerificationTable{},
		&models.GoogleResponse{},
		&models.Admin{},
	)
	if err != nil {
		log.Fatal("Failed to automigrate models:", err)
	}
}
