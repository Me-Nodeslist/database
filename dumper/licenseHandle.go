package dumper

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Transfer struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
}

func (d *Dumper) HandleLicenseMint(log types.Log) error {
	var out Transfer
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	// store info to db
	return nil
}
