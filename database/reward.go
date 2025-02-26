package database

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type DelMEMOMintInfo struct {
	gorm.Model
	Depositer string
	Receiver  string
	Amount    string
}

type RedeemInfo struct {
	gorm.Model
	RedeemID     string `gorm:"uniqueIndex;column:redeemid"`
	Initiator    string
	RedeemAmount string
	ClaimAmount  string
	LockDuration uint32
	UnlockDate   int64
	Canceled     bool
	Claimed      bool
}

type RewardWithdrawInfo struct {
	gorm.Model
	Receiver string
	Amount   string
}

func InitDelMEMOMintInfoTable() error {
	return GlobalDataBase.AutoMigrate(&DelMEMOMintInfo{})
}

func (dm *DelMEMOMintInfo) CreateDelMEMOMintInfo() error {
	return GlobalDataBase.Create(dm).Error
}

func GetAllMintAmount() (*big.Int, error) {
	var delMemoMintInfos []DelMEMOMintInfo
	res := big.NewInt(0)
	err := GlobalDataBase.Model(&DelMEMOMintInfo{}).Find(&delMemoMintInfos).Error
	if err != nil {
		return nil, err
	}
	for _, info := range delMemoMintInfos {
		amount, ok := new(big.Int).SetString(info.Amount, 10)
		if !ok {
			continue
		}
		res = res.Add(res, amount)
	}
	return res, err
}

// ------------------RedeemInfo--------------------
func InitRedeemInfo() error {
	return GlobalDataBase.AutoMigrate(&RedeemInfo{})
}

func (r *RedeemInfo) CreateRedeemInfo() error {
	return GlobalDataBase.Create(r).Error
}

func (r *RedeemInfo) UpdateRedeemInfo(redeemID string) error {
	return GlobalDataBase.Model(&RedeemInfo{}).Where("redeemid = ?", r.RedeemID).Updates(map[string]interface{}{"canceled": r.Canceled, "claimed": r.Claimed}).Error
}

func GetRedeemInfosByInitiator(initiatorAddr common.Address, offset int, limit int) ([]RedeemInfo, error) {
	var infos []RedeemInfo
	initiator := initiatorAddr.Hex()
	err := GlobalDataBase.Model(&RedeemInfo{}).Where("initiator = ?", initiator).Offset(offset).Limit(limit).Find(&infos).Error
	if err != nil {
		return infos, err
	}
	return infos, nil
}

func GetRedeemingAmountByInitiator(initiatorAddr common.Address) (*big.Int, error) {
	var infos []RedeemInfo
	res := big.NewInt(0)
	initiator := initiatorAddr.Hex()
	now := time.Now().Unix()
	err := GlobalDataBase.Model(&RedeemInfo{}).Where("initiator = ? AND canceled = ? AND unlock_date > ?", initiator, false, now).Find(&infos).Error
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		redeemAmount, ok := new(big.Int).SetString(info.RedeemAmount, 10)
		if !ok {
			continue
		}
		res = res.Add(res, redeemAmount)
	}
	return res, nil
}

func GetLockedAmountByInitiator(initiatorAddr common.Address) (*big.Int, error) {
	var infos []RedeemInfo
	res := big.NewInt(0)
	initiator := initiatorAddr.Hex()
	now := time.Now().Unix()
	err := GlobalDataBase.Model(&RedeemInfo{}).Where("initiator = ? AND canceled = ? AND unlock_date > ?", initiator, false, now).Find(&infos).Error
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		claimAmount, ok := new(big.Int).SetString(info.ClaimAmount, 10)
		if !ok {
			continue
		}
		res = res.Add(res, claimAmount)
	}
	return res, nil
}

func GetUnClaimedAmountByInitiator(initiatorAddr common.Address) (*big.Int, error) {
	var infos []RedeemInfo
	res := big.NewInt(0)
	initiator := initiatorAddr.Hex()
	now := time.Now().Unix()
	err := GlobalDataBase.Model(&RedeemInfo{}).Where("initiator = ? AND canceled = ? AND unlock_date <= ? AND claimed = ?", initiator, false, now, false).Find(&infos).Error
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		claimAmount, ok := new(big.Int).SetString(info.ClaimAmount, 10)
		if !ok {
			continue
		}
		res = res.Add(res, claimAmount)
	}
	return res, nil
}

func GetUnClaimedRedeemIDsByInitiator(initiatorAddr common.Address) ([]*big.Int, error) {
	var infos []RedeemInfo
	res := make([]*big.Int, 0)
	initiator := initiatorAddr.Hex()
	now := time.Now().Unix()
	err := GlobalDataBase.Model(&RedeemInfo{}).Where("initiator = ? AND canceled = ? AND unlock_date <= ? AND claimed = ?", initiator, false, now, false).Find(&infos).Error
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		redeemID, ok := new(big.Int).SetString(info.RedeemID, 10)
		if !ok {
			continue
		}
		res = append(res, redeemID)
	}
	return res, nil
}

func GetClaimedAmountByInitiator(initiatorAddr common.Address) (*big.Int, error) {
	var infos []RedeemInfo
	res := big.NewInt(0)
	initiator := initiatorAddr.Hex()
	err := GlobalDataBase.Model(&RedeemInfo{}).Where("initiator = ? AND canceled = ? AND claimed = ?", initiator, false, true).Find(&infos).Error
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		claimAmount, ok := new(big.Int).SetString(info.ClaimAmount, 10)
		if !ok {
			continue
		}
		res = res.Add(res, claimAmount)
	}
	return res, nil
}

// ------------------RedeemInfo--------------------
func InitRewardWithdrawInfoTable() error {
	return GlobalDataBase.AutoMigrate(&RewardWithdrawInfo{})
}

func (rw *RewardWithdrawInfo) CreateRewardWithdrawInfo() error {
	return GlobalDataBase.Create(rw).Error
}

func GetWithdrawInfosByReceiver(receiverAddr common.Address) ([]RewardWithdrawInfo, error) {
	var infos []RewardWithdrawInfo
	receiver := receiverAddr.Hex()
	err := GlobalDataBase.Model(&RewardWithdrawInfo{}).Where("receiver = ?", receiver).Find(&infos).Error
	if err != nil {
		return nil, err
	}
	return infos, err
}

func GetTotalWithdrawAmountByReceiver(receiverAddr common.Address) (*big.Int, error) {
	var infos []RewardWithdrawInfo
	res := big.NewInt(0)
	receiver := receiverAddr.Hex()
	err := GlobalDataBase.Model(&RewardWithdrawInfo{}).Where("receiver = ?", receiver).Find(&infos).Error
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		amount,ok := new(big.Int).SetString(info.Amount, 10)
		if !ok {
			continue
		}
		res = res.Add(res, amount)
	}
	return res, err
}
