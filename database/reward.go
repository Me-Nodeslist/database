package database

import (
	"gorm.io/gorm"
)

type DelMEMOMintInfo struct {
	gorm.Model
	Depositer       string `gorm:"uniqueIndex;column:depositer"`
	Receiver         string
	Amount     string
}

type RedeemInfo struct {
	gorm.Model
	Initiator string `gorm:"uniqueIndex;column:initiator"`
	RedeemAmount string
	ClaimAmount string
	LockDuration uint32
	UnlockDate string
	Canceled bool
	Claimed bool
}

type ClaimedInfo struct {
	gorm.Model
	Initiator string `gorm:"uniqueIndex;column:initiator"`
	RedeemingAmount string
	WaitClaimAmount string
	TotalClaimedAmount string
}

type RewardWithdrawInfo struct {
	gorm.Model
	Receiver string `gorm:"uniqueIndex;column:receiver"`
	Amount         string
}
