package database

import (
	"fmt"
	"log"
	"os"

	"github.com/susapa/Inventory_API/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error

	// Connection parameters
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "inventory_db")
	port := getEnv("DB_PORT", "5432")

	// 1. Connect to default 'postgres' database to check/create the target database
	dsnMaster := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable TimeZone=Asia/Bangkok",
		host, user, password, port)

	masterDB, err := gorm.Open(postgres.Open(dsnMaster), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to master postgres database: %v", err)
	}

	// Check if database exists
	var exists bool
	masterDB.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", dbname).Scan(&exists)

	if !exists {
		log.Printf("Database %s does not exist, creating it...", dbname)
		// We use masterDB.Exec because CREATE DATABASE cannot run inside a transaction
		if err := masterDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname)).Error; err != nil {
			log.Fatalf("failed to create database: %v", err)
		}
		log.Printf("Database %s created successfully", dbname)
	}

	// Close master connection
	sqlDB, _ := masterDB.DB()
	sqlDB.Close()

	// 2. Now connect to the actual target database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		host, user, password, dbname, port)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	log.Printf("Successfully connected to database: %s", dbname)

	// Auto Migration
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
