package database

import (
	"gorm.io/gorm"
)

type NodeInfo struct {
	gorm.Model
	NodeID       uint32 `gorm:"uniqueIndex;column:nodeID"`
	NodeAddress         string `gorm:"uniqueIndex;column:nodeAddress"`
	Recipient     string
	Active bool
	CommissionRate uint8
	DelegationAmount uint16
	SelfTotalReward string
	SelfWithdrawedReward string
	DelegationReward string
	CommissionRateLastModifyAt string
	RegisterDate string
	OnlineDays string
	OnlineDays_RecentMonth uint8
	OnlineDays_recentWeek uint8
}

type DelegationInfo struct {
	gorm.Model
	LicenseOwner string `gorm:"uniqueIndex;column:nodeAddress"`
	TokenID string `gorm:"uniqueIndex;column:tokenID"`
	NodeAddress string `gorm:"uniqueIndex;column:nodeAddress"`
	TotalReward string
	WithdrawedReward string
}

type NodeDailyDelegation struct {
	gorm.Model
	NodeAddress string `gorm:"uniqueIndex;column:nodeAddress"`
	Date string `gorm:"uniqueIndex;column:date"`
	DelegationAmount uint16
}
