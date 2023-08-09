package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DBCon *gorm.DB
)

func InitDB() {
	var err error
	nerostrLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,
		},
	)

	db_backend := os.Getenv("DB_BACKEND")
	if db_backend == "sqlite" {
		DBCon, err = gorm.Open(sqlite.Open("/app/db/whisper.db"), &gorm.Config{
			Logger: nerostrLogger,
		})
		if err != nil {
			log.Printf("Error decoding body into struct %v", err)
			return
		}
		return
	} else {
		dsn := "webwhisper:webwhisper@tcp(database:3306)/webwhisper?charset=utf8mb4&parseTime=True&loc=Local"
		DBCon, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: nerostrLogger,
		})
	}

	if err != nil {
		log.Printf("Error decoding body into struct %v", err)
		return
	}
}

func Migrate() {
	DBCon.AutoMigrate(TranscriptionJob{})
	DBCon.AutoMigrate(Translation{})
}
