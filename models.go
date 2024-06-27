package main

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"unique"`
	Password  string
	Name      string
	Gender    string
	DOB       time.Time
	Latitude  float64
	Longitude float64
}

type SwipePair struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint
	TargetUserID uint
	Preference   string
	Match        bool
}
