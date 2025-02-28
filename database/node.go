package database

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type NodeInfoOnChain struct {
	Id                         uint32
	Active                     bool
	LastConfirmDate            uint32
	CommissionRate             uint8
	Recipient                  common.Address
	SelfTotalRewards           *big.Int
	SelfClaimedRewards         *big.Int
	DelegationRewards          *big.Int
	CommissionRateLastModifyAt *big.Int
}
type NodeInfo struct {
	gorm.Model
	NodeID                     uint32 `gorm:"uniqueIndex;column:nodeid"`
	NodeAddress                string `gorm:"uniqueIndex"`
	Recipient                  string
	Active                     bool
	CommissionRate             uint8
	DelegationAmount           uint16
	SelfTotalReward            string
	SelfWithdrawedReward       string
	DelegationReward           string
	CommissionRateLastModifyAt string
	RegisterDate               string
	OnlineDays                 int64
	OnlineDays_RecentMonth     int64
	OnlineDays_RecentWeek      int64
}

type NodeDailyDelegation struct {
	gorm.Model
	NodeAddress      string
	Date             uint16
	DelegationAmount uint16
}

func InitNodeInfoTable() error {
	return GlobalDataBase.AutoMigrate(&NodeInfo{})
}

func (n *NodeInfo) CreateNodeInfo() error {
	return GlobalDataBase.Create(n).Error
}

func (n *NodeInfo) UpdateNodeCommissionRate() error {
	return GlobalDataBase.Model(&NodeInfo{}).Where("node_address = ?", n.NodeAddress).Updates(map[string]interface{}{"commission_rate": n.CommissionRate, "commission_rate_last_modify_at": n.CommissionRateLastModifyAt}).Error
}

func (n *NodeInfo) UpdateNodeDelegationAmount() error {
	return GlobalDataBase.Model(&LicenseInfo{}).Where("node_address = ?", n.NodeAddress).Updates(map[string]interface{}{"delegation_amount": n.DelegationAmount, "active": n.Active}).Error
}

func (n *NodeInfo) UpdateNodeRewardInfo() error {
	return GlobalDataBase.Model(&LicenseInfo{}).Where("node_address = ?", n.NodeAddress).Updates(map[string]interface{}{"self_total_reward": n.SelfTotalReward, "self_withdrawed_reward": n.SelfWithdrawedReward, "delegation_reward": n.DelegationReward}).Error
}

func (n *NodeInfo) UpdateNodeOnlineDays() error {
	return GlobalDataBase.Model(&LicenseInfo{}).Where("node_address = ?", n.NodeAddress).Updates(map[string]interface{}{"online_days": n.OnlineDays, "online_days_recent_month": n.OnlineDays_RecentMonth, "online_days_recent_week": n.OnlineDays_RecentWeek}).Error
}

func GetNodeAmount() (int64, error) {
	var length int64
	err := GlobalDataBase.Model(&NodeInfo{}).Count(&length).Error
	return length, err
}

func GetActiveNodeAmount() (int64, error) {
	var length int64
	err := GlobalDataBase.Model(&NodeInfo{}).Where("active = ?", true).Count(&length).Error
	return length, err
}

func GetNodeInfoByNodeAddress(nodeAddr common.Address) (NodeInfo, error) {
	var nodeInfo NodeInfo
	node := nodeAddr.Hex()
	err := GlobalDataBase.Model(&NodeInfo{}).Where("node_address = ?", node).First(&nodeInfo).Error
	if err != nil {
		return NodeInfo{}, err
	}
	return nodeInfo, nil
}

func GetNodeInfoByNodeID(nodeID uint32) (NodeInfo, error) {
	var nodeInfo NodeInfo
	err := GlobalDataBase.Model(&NodeInfo{}).Where("nodeid = ?", nodeID).First(&nodeInfo).Error
	if err != nil {
		return NodeInfo{}, err
	}
	return nodeInfo, nil
}

func GetNodeInfos(offset int, limit int) ([]NodeInfo, error) {
	var nodeInfos []NodeInfo
	err := GlobalDataBase.Model(&NodeInfo{}).Offset(offset).Limit(limit).Find(&nodeInfos).Error
	if err != nil {
		return nodeInfos, err
	}
	return nodeInfos, nil
}

func GetActiveNodeInfos(offset int, limit int) ([]NodeInfo, error) {
	var nodeInfos []NodeInfo
	err := GlobalDataBase.Model(&NodeInfo{}).Where("active = ?", true).Offset(offset).Limit(limit).Find(&nodeInfos).Error
	if err != nil {
		return nodeInfos, err
	}
	return nodeInfos, nil
}

// ------------------NodeDailyDelegation--------------------
func InitNodeDailyDelegation() error {
	return GlobalDataBase.AutoMigrate(&NodeDailyDelegation{})
}

func (n *NodeDailyDelegation) CreateNodeDailyDelegation() error {
	return GlobalDataBase.Create(n).Error
}

func (n *NodeDailyDelegation) UpdateNodeDailyDelegation() error {
	return GlobalDataBase.Model(&NodeDailyDelegation{}).Where("node_address = ? AND date = ?", n.NodeAddress, n.Date).Updates(map[string]interface{}{"delegation_amount": n.DelegationAmount}).Error
}

func GetNodeDailyDelegation(nodeAddr common.Address, date uint16) (NodeDailyDelegation, error) {
	var nodeDailyDelegation NodeDailyDelegation
	node := nodeAddr.Hex()
	err := GlobalDataBase.Model(&NodeDailyDelegation{}).Where("node_address = ? AND date = ?", node, date).First(&nodeDailyDelegation).Error
	if err != nil {
		return NodeDailyDelegation{}, err
	}
	return nodeDailyDelegation, nil
}

func GetNodeRecentOnlineDays(nodeAddr common.Address, date uint16) (int64, int64, error) {
	var length_month int64
	var length_week int64
	node := nodeAddr.Hex()
	recentMonth := uint16(0)
	recentWeek := uint16(0)
	if date > 7 {
		recentWeek = date - 7
	}
	if date > 30 {
		recentMonth = date - 30
	}
	err := GlobalDataBase.Model(&NodeDailyDelegation{}).Where("node_address = ? AND date >= ?", node, recentMonth).Count(&length_month).Error
	if err != nil {
		return length_month, length_week, err
	}
	err = GlobalDataBase.Model(&NodeDailyDelegation{}).Where("node_address = ? AND date >= ?", node, recentWeek).Count(&length_week).Error
	if err != nil {
		return length_month, length_week, err
	}
	return length_month, length_week, nil
}

func GetGlobalDailyDelegation(date uint16) (uint32, error) {
	var globalDailyDelegation uint32
	err := GlobalDataBase.Model(&NodeDailyDelegation{}).Where("date = ?", date).Select("sum(delegation_amount)").Scan(&globalDailyDelegation).Error
	if err != nil {
		return 0, err
	}
	return globalDailyDelegation, nil
}
