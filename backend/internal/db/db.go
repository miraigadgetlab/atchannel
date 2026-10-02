package db

import (
	"atchannel-backend/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "atchannel_user")
	password := getEnv("DB_PASSWORD", "atchannel_password")
	dbname := getEnv("DB_NAME", "atchannel_db")
	port := getEnv("DB_PORT", "5432")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}

	fmt.Println("Database connection successful")
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.Channeler{},
		&models.Channel{},
		&models.Post{},
		&models.Comment{},
	)

	if err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	fmt.Println("Database migration completed successfully.")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
