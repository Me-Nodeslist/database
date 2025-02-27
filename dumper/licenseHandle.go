package dumper

import (
	"math/big"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type LicenseTransfer struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
}

func (d *Dumper) HandleLicenseMint(log types.Log) error {
	var out LicenseTransfer
	err := d.unpack(log, 0, &out)
	if err != nil {
		return err
	}

	if out.From.Hex() != common.BigToAddress(big.NewInt(0)).Hex() {
		logger.Info("License Transfer event: From is not address(0), From is ", out.From.Hex())
		return nil
	}

	// store info to db
	licenseInfo := database.LicenseInfo{
		TokenID: out.TokenID.String(),
		Owner: out.To.Hex(),
	}
	return licenseInfo.CreateLicenseInfo()
}
