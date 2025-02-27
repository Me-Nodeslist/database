package dumper

import (
	"math/big"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type DelMEMOTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

type RedeemEvent struct {
	RedeemID    *big.Int
	Initiator   common.Address
	Amount      *big.Int
	ClaimAmount *big.Int
	Duration    uint32
}

type CancelRedeemEvent struct {
	RedeemID *big.Int
}

type ClaimEvent struct {
	RedeemID *big.Int
	Amount   *big.Int
}

func (d *Dumper) HandleDelMemoMint(log types.Log) error {
	var out DelMEMOTransfer
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.DelMEMOMintInfo{
		Depositer: out.From.Hex(),
		Receiver:  out.To.Hex(),
		Amount:    out.Value.String(),
	}
	return info.CreateDelMEMOMintInfo()
}

func (d *Dumper) HandleDelMemoTransfer(log types.Log) error {
	var out DelMEMOTransfer
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	if out.From.Hex() == common.BigToAddress(big.NewInt(0)).Hex() {
		logger.Debug("DelMEMO Transfer event: From is 0")
		return nil
	}

	// store info to db
	info := database.DelMEMOTransferInfo{
		From:   out.From.Hex(),
		To:     out.To.Hex(),
		Amount: out.Value.String(),
	}
	return info.CreateDelMEMOTransferInfo()
}

func (d *Dumper) HandleDelMemoRedeem(log types.Log, time uint64) error {
	var out RedeemEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.RedeemInfo{
		RedeemID:     out.RedeemID.String(),
		Initiator:    out.Initiator.Hex(),
		RedeemAmount: out.Amount.String(),
		ClaimAmount:  out.ClaimAmount.String(),
		LockDuration: out.Duration,
		UnlockDate:   int64(time) + int64(out.Duration),
	}
	return info.CreateRedeemInfo()
}

func (d *Dumper) HandleDelMemoCancelRedeem(log types.Log) error {
	var out CancelRedeemEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.RedeemInfo{
		RedeemID: out.RedeemID.String(),
		Canceled: true,
	}
	return info.UpdateRedeemInfo()
}

func (d *Dumper) HandleDelMemoClaim(log types.Log) error {
	var out ClaimEvent
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.RedeemInfo{
		RedeemID: out.RedeemID.String(),
		Claimed: true,
	}
	return info.UpdateRedeemInfo()
}
