package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDatabase() {
	var err error
	db, err = gorm.Open(sqlite.Open("app.db"))
	if err != nil {
		log.Fatal("Failed to initialise database:", err)
	}

	db.AutoMigrate(&User{}, &SwipePair{})
}
