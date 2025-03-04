package dumper

import (
	"math/big"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type RewardWithdrawEvent struct {
	Receiver common.Address
	Amount   *big.Int
}

type FoundationWithdrawEvent struct {
	Foundation common.Address
	Amount   *big.Int
}

func (d *Dumper) HandleSettlementRewardWithdraw(log types.Log) error {
	var out RewardWithdrawEvent
	err := d.unpack(log, 2, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.RewardWithdrawInfo{
		Receiver:  out.Receiver.Hex(),
		Amount:    out.Amount.String(),
	}
	return info.CreateRewardWithdrawInfo()
}

func (d *Dumper) HandleSettlementFoundationWithdraw(log types.Log) error {
	var out FoundationWithdrawEvent
	err := d.unpack(log, 2, &out)
	if err != nil {
		return err
	}

	// store info to db
	info := database.RewardWithdrawInfo{
		Receiver:  out.Foundation.Hex(),
		Amount:    out.Amount.String(),
	}
	return info.CreateRewardWithdrawInfo()
}
