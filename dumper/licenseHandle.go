package dumper

import (
	"context"
	"errors"
	"math/big"
	"os"

	"github.com/Me-Nodeslist/database/database"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type LicenseTransfer struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
}

type MetaData struct {
	Code  string
	Price uint64
	Tier  uint8
}

func (d *Dumper) HandleLicenseMint(log types.Log) error {
	from, to, tokenID := d.unpackLicenseTransfer(log)

	if from != common.BigToAddress(big.NewInt(0)).Hex() {
		logger.Debug("License Transfer event: From is not address(0), From is ", from)
		return nil
	}

	// store info to db
	licenseInfo := database.LicenseInfo{
		TokenID: tokenID,
		Owner:   to,
	}
	return licenseInfo.CreateLicenseInfo()
}

func (d *Dumper) PurchaseTxValid(txHash string, receiver string, shouldValue float64, amount int64) (bool, error) {
	logger.Debug("txHash:", txHash, " receiver:", receiver)

	client, err := ethclient.DialContext(context.TODO(), d.endpoint)
	if err != nil {
		logger.Error(err.Error())
		return false, err
	}
	defer client.Close()

	tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		logger.Error(err)
		return false, err
	}
	logger.Debug("tx is pending:", isPending)
	logger.Debug("tx timestamp:", tx.Time().Format("2006-01-02 15:04:05"))

	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		logger.Error(err)
		return false, err
	}

	if receipt.Status == types.ReceiptStatusFailed {
		logger.Debug("tx receipt status is failed")
		return false, errors.New("tx receipt status is failed")
	}

	if tx.To().Hex() != LICENSE_PAYMENT_RECEIVER {
		logger.Debug("tx 'to' is not our receiver")
		return false, errors.New("tx 'to' is not our receiver")
	}

	shouldValue_Wei := int64(shouldValue * 1e18)
	shouldValue_Wei_bigInt := big.NewInt(shouldValue_Wei)
	if tx.Value().Cmp(shouldValue_Wei_bigInt) < 0 {
		if new(big.Int).Sub(shouldValue_Wei_bigInt, tx.Value()).Int64() > int64(float64(shouldValue_Wei) * PAYMENT_DEVIATION) {
			logger.Debugf("tx value %n, should pay %n, license amount %n", tx.Value(), shouldValue_Wei, amount)
			return false, errors.New("tx value " + tx.Value().String() + ", should pay " + shouldValue_Wei_bigInt.String() + ", difference is too large")
		}
	}

	signer := types.LatestSignerForChainID(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		logger.Debug("parse 'from' failed")
		return false, errors.New("parse 'from' failed")
	}
	if from.Hex() != receiver {
		logger.Debug("from is", from.Hex(), " but receiver is", receiver)
		return false, errors.New("tx sender is different from receiver")
	}

	history, err := database.GetPurchaseHistoryByTxHash(txHash)
	if err == nil && history.Done {
		logger.Debug("this purchase had been done")
		return false, errors.New("this purchase had been done")
	}

	return true, nil
}

func (d *Dumper) MintNFT(receiver string, amount int64, price uint64) (string, error) {
	client, err := ethclient.DialContext(context.TODO(), d.endpoint)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		logger.Error("Load privateKey failed")
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logger.Errorf("Get nonce of %s failed", fromAddress.Hex())
		return "", err
	}

	userAddr := common.HexToAddress(receiver)
	metaData := MetaData{
		Price: price,
	}
	data, err := d.contractABI[0].Pack("mint", userAddr, big.NewInt(amount), metaData)
	if err != nil {
		logger.Error("Pack mint tx input failed")
		return "", err
	}

	gasLimit := uint64(200000)
	gasPrice, _ := client.SuggestGasPrice(context.Background())

	tx := types.NewTransaction(nonce, d.contractAddress[0], nil, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logger.Error("Get chainID failed")
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		logger.Error("Sign tx failed")
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logger.Error("Send tx failed")
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}
