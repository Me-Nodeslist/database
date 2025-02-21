package database

import (
	"gorm.io/gorm"
)

type FoundationWithdrawInfo struct {
	gorm.Model
	Foundation string `gorm:"uniqueIndex;column:foundation"`
	Amount         string
}
