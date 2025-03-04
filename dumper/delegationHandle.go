package dumper

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ModifyCommissionRateEvent struct {
	Node           common.Address
	CommissionRate uint8
}

type NodeWithdrawEvent struct {
	Node   common.Address
	Reward *big.Int
}

type ConfirmNodeRewardEvent struct {
	Node              common.Address
	SelfTotalRewards  *big.Int
	DelegationRewards *big.Int
}

type NodeDailyDelegationsEvent struct {
	Node             common.Address
	Date             uint32
	DelegationAmount uint16
}

type DelegateEvent struct {
	TokenID *big.Int
	To    common.Address
}

type ClaimRewardEvent struct {
	Owner   common.Address
	TokenID *big.Int
	Amount  *big.Int
}

type NodeRegisterEvent struct {
	Node           common.Address
	Recipient      common.Address
	CommissionRate uint8
}

func (d *Dumper) HandleModifyCommissionRate(log types.Log, time uint64) error {
	var out ModifyCommissionRateEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.NodeInfo{
		NodeAddress:                out.Node.Hex(),
		CommissionRate:             out.CommissionRate,
		CommissionRateLastModifyAt: strconv.FormatUint(time, 10),
	}
	return info.UpdateNodeCommissionRate()
}

func (d *Dumper) HandleNodeWithdraw(log types.Log) error {
	var out NodeWithdrawEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	info, err := database.GetNodeInfoByNodeAddress(out.Node)
	if err != nil {
		return err
	}
	value, ok := new(big.Int).SetString(info.SelfWithdrawedReward, 10)
	if !ok {
		return errors.New("transfer string to big.Int error")
	}
	tr := info.SelfTotalReward
	dr := info.DelegationReward

	// store info to db
	info = database.NodeInfo{
		NodeAddress:          out.Node.Hex(),
		SelfTotalReward:      tr,
		SelfWithdrawedReward: out.Reward.Add(out.Reward, value).String(),
		DelegationReward:     dr,
	}
	return info.UpdateNodeRewardInfo()
}

func (d *Dumper) HandleConfirmNodeReward(log types.Log) error {
	var out ConfirmNodeRewardEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	info, err := database.GetNodeInfoByNodeAddress(out.Node)
	if err != nil {
		return err
	}

	// store info to db
	info.SelfTotalReward = out.SelfTotalRewards.String()
	info.DelegationReward = out.DelegationRewards.String()
	err = info.UpdateNodeRewardInfo()
	if err != nil {
		return err
	}

	// update license reward
	licenseInfos, err := database.GetLicenseInfosByNode(out.Node, int(0), int(info.DelegationAmount))
	if err != nil {
		return err
	}
	for _, licenseInfo := range licenseInfos {
		initialReward, ok := new(big.Int).SetString(licenseInfo.InitialReward, 10)
		if !ok {
			continue
		}
		totalReward, ok := new(big.Int).SetString(licenseInfo.TotalReward, 10)
		if !ok {
			continue
		}
		addReward := new(big.Int).Sub(out.DelegationRewards, initialReward)
		totalReward = totalReward.Add(totalReward, addReward)
		licenseInfo.TotalReward = totalReward.String()
		licenseInfo.InitialReward = info.DelegationReward
		licenseInfo.UpdateLicenseReward()
	}
	return nil
}

func (d *Dumper) HandleNodeDailyDelegations(log types.Log) error {
	var out NodeDailyDelegationsEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.NodeDailyDelegation{
		NodeAddress:      out.Node.Hex(),
		Date:             uint16(out.Date),
		DelegationAmount: out.DelegationAmount,
	}
	_, err = database.GetNodeDailyDelegation(out.Node, info.Date)
	if err == nil { // exist
		err = info.UpdateNodeDailyDelegation()
	} else {
		err = info.CreateNodeDailyDelegation()
	}
	if err != nil {
		return err
	}

	// online days
	nodeInfo, err := database.GetNodeInfoByNodeAddress(out.Node)
	if err != nil {
		return err
	}
	length_month, length_week, err := database.GetNodeRecentOnlineDays(out.Node, info.Date)
	if err != nil {
		return err
	}
	nodeInfo.OnlineDays++
	nodeInfo.OnlineDays_RecentMonth = length_month
	nodeInfo.OnlineDays_RecentWeek = length_week
	return nodeInfo.UpdateNodeOnlineDays()
}

func (d *Dumper) HandleDelegate(log types.Log) error {
	var out DelegateEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	licenseInfo := database.LicenseInfo{
		TokenID:       out.TokenID.String(),
		Delegated:     true,
		DelegatedNode: out.To.Hex(),
	}
	err = licenseInfo.UpdateLicenseDelegation()
	if err != nil {
		return err
	}

	info, err := database.GetNodeInfoByNodeAddress(out.To)
	if err != nil {
		return err
	}
	amount := info.DelegationAmount + 1

	// store info to db
	info = database.NodeInfo{
		NodeAddress:      out.To.Hex(),
		DelegationAmount: amount,
		Active:           true,
	}
	return info.UpdateNodeDelegationAmount()
}

func (d *Dumper) HandleUndelegate(log types.Log) error {
	var out DelegateEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	info, err := database.GetNodeInfoByNodeAddress(out.To)
	if err != nil {
		return err
	}
	amount := info.DelegationAmount - 1
	active := true
	if amount == 0 {
		active = false
	}

	// store info to db
	info = database.NodeInfo{
		NodeAddress:      out.To.Hex(),
		DelegationAmount: amount,
		Active:           active,
	}
	err = info.UpdateNodeDelegationAmount()
	if err != nil {
		return err
	}
	licenseInfo := database.LicenseInfo{
		TokenID:       out.TokenID.String(),
		Delegated:     false,
		DelegatedNode: common.BigToAddress(big.NewInt(0)).Hex(),
	}
	return licenseInfo.UpdateLicenseDelegation()
}

func (d *Dumper) HandleRedelegate(log types.Log) error {
	var out DelegateEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	info, err := database.GetNodeInfoByNodeAddress(out.To)
	if err != nil {
		return err
	}
	licenseInfo, err := database.GetLicenseInfoByTokenID(out.TokenID.String())
	if err != nil {
		return err
	}
	infoOld, err := database.GetNodeInfoByNodeAddress(common.HexToAddress(licenseInfo.DelegatedNode))
	if err != nil {
		return err
	}

	// old node
	amount := infoOld.DelegationAmount - 1
	active := true
	if amount == 0 {
		active = false
	}
	// store info to db
	nodeInfo := database.NodeInfo{
		NodeAddress:      infoOld.NodeAddress,
		DelegationAmount: amount,
		Active:           active,
	}
	err = nodeInfo.UpdateNodeDelegationAmount()
	if err != nil {
		return err
	}

	// new node
	amount = info.DelegationAmount + 1
	// store info to db
	nodeInfo = database.NodeInfo{
		NodeAddress:      info.NodeAddress,
		DelegationAmount: amount,
		Active:           true,
	}
	err = nodeInfo.UpdateNodeDelegationAmount()
	if err != nil {
		return err
	}

	// license
	licenseInfo = database.LicenseInfo{
		TokenID:       out.TokenID.String(),
		Delegated:     true,
		DelegatedNode: out.To.Hex(),
	}
	return licenseInfo.UpdateLicenseDelegation()
}

func (d *Dumper) HandleClaimReward(log types.Log) error {
	var out ClaimRewardEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	info, err := database.GetLicenseInfoByTokenID(out.TokenID.String())
	if err != nil {
		return err
	}
	amount, ok := new(big.Int).SetString(info.WithdrawedReward, 10)
	if !ok {
		return errors.New("transfer string to big.Int error")
	}
	amount = amount.Add(amount, out.Amount)

	// store info to db
	info.WithdrawedReward = amount.String()
	return info.UpdateLicenseReward()
}

func (d *Dumper) GetNodeAddr(log types.Log) (common.Address, error) {
	var out NodeRegisterEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return common.Address{}, err
	}
	return out.Node, nil
}

func (d *Dumper) HandleNodeRegister(log types.Log, time uint64, nodeInfo *database.NodeInfoOnChain) error {
	var out NodeRegisterEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.NodeInfo{
		NodeID:                     nodeInfo.Id,
		NodeAddress:                out.Node.Hex(),
		Recipient:                  out.Recipient.Hex(),
		Active:                     nodeInfo.Active,
		CommissionRate:             out.CommissionRate,
		CommissionRateLastModifyAt: nodeInfo.CommissionRateLastModifyAt.String(),
		RegisterDate:               strconv.FormatUint(time, 10),
	}
	return info.CreateNodeInfo()
}
