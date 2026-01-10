package database

import (
	"case-trading/app/models"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

func GetDB() *gorm.DB {
	return db
}

func InitDB() {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" || user == "" || password == "" || dbname == "" || port == "" {
		panic("environment variables are required")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbname)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Error),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
}

func SyncDB() {
	// Auto migrate
	err := db.AutoMigrate(
		&models.Users{},
		&models.VerifyToken{},
		&models.Wallets{},
		&models.Market{},
		&models.Order{},
		&models.Trade{},
	)
	if err != nil {
		panic(err)
	}
}
