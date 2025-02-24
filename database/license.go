package database

import (
	"gorm.io/gorm"
	"github.com/ethereum/go-ethereum/common"
)

type LicenseInfo struct {
	gorm.Model
	TokenID       string `gorm:"uniqueIndex;column:tokenid"`
	Owner         string
	Delegated     bool
	DelegatedNode string
	TotalReward string
	WithdrawedReward string
}

func InitLicenseInfoTable() error {
	return GlobalDataBase.AutoMigrate(&LicenseInfo{})
}

func (l *LicenseInfo) CreateLicenseInfo() error {
	return GlobalDataBase.Create(l).Error
}

func (l *LicenseInfo) UpdateLicenseOwner() error {
	return GlobalDataBase.Model(&LicenseInfo{}).Where("tokenid = ?", l.TokenID).Updates(map[string]interface{}{"owner": l.Owner}).Error
}

func (l *LicenseInfo) UpdateLicenseDelegation() error {
	return GlobalDataBase.Model(&LicenseInfo{}).Where("tokenid = ?", l.TokenID).Updates(map[string]interface{}{"delegated": l.Delegated, "delegated_node": l.DelegatedNode}).Error
}

func (l *LicenseInfo) UpdateLicenseReward() error {
	return GlobalDataBase.Model(&LicenseInfo{}).Where("tokenid = ?", l.TokenID).Updates(map[string]interface{}{"total_reward": l.TotalReward, "withdrawed_reward": l.WithdrawedReward}).Error
}

func GetLicenseAmount() (int64, error) {
	var length int64
	err := GlobalDataBase.Model(&LicenseInfo{}).Count(&length).Error
	return length, err
}

func GetLicenseAmountByOwner(ownerAddr common.Address) (int64, error) {
	var length int64
	owner := ownerAddr.Hex()
	err := GlobalDataBase.Model(&LicenseInfo{}).Where("owner = ?", owner).Count(&length).Error
	return length, err
}

func GetLicenseAmountByNode(delegatedNodeAddr common.Address) (int64, error) {
	var length int64
	delegatedNode := delegatedNodeAddr.Hex()
	err := GlobalDataBase.Model(&LicenseInfo{}).Where("delegated_node", delegatedNode).Count(&length).Error
	return length, err
}

func GetLicenseInfoByTokenID(tokenID string) (LicenseInfo, error) {
	var licenseInfo LicenseInfo
	err := GlobalDataBase.Model(&LicenseInfo{}).Where("tokenid = ?", tokenID).First(&licenseInfo).Error
	if err != nil {
		return LicenseInfo{}, err
	}
	return licenseInfo, nil
}
