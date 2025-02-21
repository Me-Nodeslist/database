package database

import (
	"gorm.io/gorm"
)

type LicenseInfo struct {
	gorm.Model
	TokenID       string `gorm:"uniqueIndex;column:tokenID"`
	Owner         int64
	Delegated     bool
	DelegatedNode string
}
