package dumper

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Me-Nodeslist/database/database"
	"github.com/Me-Nodeslist/database/logs"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ContractAddress struct {
	LicenseNFT common.Address
	DelMEMO    common.Address
	Settlement common.Address
	Delegation common.Address
}

type Dumper struct {
	endpoint        string
	contractABI     []abi.ABI
	contractAddress []common.Address

	blockNumber *big.Int

	eventNameMap map[common.Hash]string
	indexedMap   map[common.Hash]abi.Arguments
}

type EtherscanResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		EthBTC           string `json:"ethbtc"`
		EthBTC_Timestamp string `json:"ethbtc_timestamp"`
		EthUSD           string `json:"ethusd"`
		EthUSD_Timestamp string `json:"ethusd_timestamp"`
	} `json:"result"`
}

const LICENSE_PAYMENT_RECEIVER = "0x389824fc8755039F165738139b255Fad711e2bCb"
const LICENSE_PRICE_USDT = 500
const PAYMENT_DEVIATION = 0.01 // accept 1% error

var (
	// blockNumber = big.NewInt(0)
	logger = logs.Logger("dumper")
)
var URL string
var EthUSD float64
var EthUSD_Timestamp int

func NewDumper(ethrpc string, addrs *ContractAddress) (dumper *Dumper, err error) {
	dumper = &Dumper{
		eventNameMap: make(map[common.Hash]string),
		indexedMap:   make(map[common.Hash]abi.Arguments),
	}

	//_, endpoint := com.GetInsEndPointByChain(chain)
	dumper.endpoint = ethrpc

	dumper.contractAddress = []common.Address{addrs.LicenseNFT, addrs.DelMEMO, addrs.Settlement, addrs.Delegation}

	projectDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	filePath := filepath.Join(projectDir, "abi", "LicenseNFT.abi")
	licenseNFTABI, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read licenseNFT abi file, ", err)
		return dumper, err
	}
	licenseContractABI, err := abi.JSON(strings.NewReader(string(licenseNFTABI)))
	if err != nil {
		return dumper, err
	}

	filePath = filepath.Join(projectDir, "abi", "DelMEMO.abi")
	delMEMOABI, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read delMEMO abi file, ", err)
		return dumper, err
	}
	delMemoContractABI, err := abi.JSON(strings.NewReader(string(delMEMOABI)))
	if err != nil {
		return dumper, err
	}

	filePath = filepath.Join(projectDir, "abi", "Settlement.abi")
	settlementABI, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read settlement abi file, ", err)
		return dumper, err
	}
	settlementContractABI, err := abi.JSON(strings.NewReader(string(settlementABI)))
	if err != nil {
		return dumper, err
	}

	filePath = filepath.Join(projectDir, "abi", "Delegation.abi")
	delegationABI, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read delegation abi file, ", err)
		return dumper, err
	}
	delegationContractABI, err := abi.JSON(strings.NewReader(string(delegationABI)))
	if err != nil {
		return dumper, err
	}

	dumper.contractABI = []abi.ABI{licenseContractABI, delMemoContractABI, settlementContractABI, delegationContractABI}

	for i := 0; i < len(dumper.contractABI); i++ {
		for name, event := range dumper.contractABI[i].Events {
			dumper.eventNameMap[event.ID] = name

			var indexed abi.Arguments
			for _, arg := range dumper.contractABI[i].Events[name].Inputs {
				if arg.Indexed {
					indexed = append(indexed, arg)
				}
			}
			dumper.indexedMap[event.ID] = indexed
		}
	}

	blockNumber, err := database.GetBlockNumber()
	if err != nil {
		blockNumber = 0
	}
	dumper.blockNumber = big.NewInt(blockNumber)

	return dumper, nil
}

func (d *Dumper) SubscribeEvents(ctx context.Context) error {
	for {
		d.Dump()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(10 * time.Second):
		}
	}
}

func (d *Dumper) SubscribeEthPrice(ctx context.Context, apikey string) error {
	URL = "https://api.etherscan.io/api?module=stats&action=ethprice&apikey=" + apikey
	log.Println("Etherscan api url:", URL)
	for {
		GetEthPriceFromEtherscan()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(30 * time.Second):
		}
	}
}

func GetEthPriceFromEtherscan() error {
	resp, err := http.Get(URL)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	var data EtherscanResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	EthUSD, err = strconv.ParseFloat(data.Result.EthUSD, 64)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	EthUSD_Timestamp, err = strconv.Atoi(data.Result.EthUSD_Timestamp)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (d *Dumper) Dump() error {
	client, err := ethclient.DialContext(context.TODO(), d.endpoint)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	defer client.Close()

	currentBlockNumber, err := client.BlockNumber(context.TODO())
	if err != nil {
		logger.Error("BlockNumber err: ", err.Error())
		return err
	}
	toBlock := big.NewInt(int64(currentBlockNumber - 1))

	eventsLicenseNFT, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		ToBlock:   toBlock,
		Addresses: []common.Address{d.contractAddress[0]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	eventsDelMEMO, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		ToBlock:   toBlock,
		Addresses: []common.Address{d.contractAddress[1]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	eventsSettlement, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		ToBlock:   toBlock,
		Addresses: []common.Address{d.contractAddress[2]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	eventsDelegation, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		ToBlock:   toBlock,
		Addresses: []common.Address{d.contractAddress[3]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	for _, event := range eventsLicenseNFT {
		eventName, ok1 := d.eventNameMap[event.Topics[0]]
		if !ok1 {
			continue
		}
		switch eventName {
		case "Transfer":
			logger.Info("Handle LicenseNFT Mint event")
			err = d.HandleLicenseMint(event)
		default:
			continue
		}
		if err != nil {
			logger.Error(err.Error())
			continue
		}
	}

	for _, event := range eventsDelMEMO {
		eventName, ok1 := d.eventNameMap[event.Topics[0]]
		if !ok1 {
			continue
		}
		switch eventName {
		case "Transfer":
			logger.Info("Handle DelMEMO transfer event")
			err = d.HandleDelMemoTransfer(event)
		case "Mint":
			logger.Info("Handle DelMEMO Mint event")
			err = d.HandleDelMemoMint(event)
		case "Redeem":
			logger.Info("Handle DelMEMO Redeem event")
			block, _ := client.BlockByNumber(context.Background(), big.NewInt(int64(event.BlockNumber)))
			err = d.HandleDelMemoRedeem(event, block.Time())
		case "CancelRedeem":
			logger.Info("Handle DelMEMO CancelRedeem event")
			err = d.HandleDelMemoCancelRedeem(event)
		case "Claim":
			logger.Info("Handle DelMEMO Claim event")
			err = d.HandleDelMemoClaim(event)
		default:
			continue
		}
		if err != nil {
			logger.Error(err.Error())
			continue
		}
	}

	for _, event := range eventsSettlement {
		eventName, ok1 := d.eventNameMap[event.Topics[0]]
		if !ok1 {
			continue
		}
		switch eventName {
		case "RewardWithdraw":
			logger.Info("Handle Settlement RewardWithdraw event")
			err = d.HandleSettlementRewardWithdraw(event)
		case "FoundationWithdraw":
			logger.Info("Handle Settlement FoundationWithdraw event")
			err = d.HandleSettlementFoundationWithdraw(event)
		default:
			continue
		}
		if err != nil {
			logger.Error(err.Error())
			continue
		}
	}

	for _, event := range eventsDelegation {
		eventName, ok1 := d.eventNameMap[event.Topics[0]]
		if !ok1 {
			continue
		}
		switch eventName {
		case "NodeRegister":
			logger.Info("Handle Delegation NodeRegister event")
			block, _ := client.BlockByNumber(context.Background(), big.NewInt(int64(event.BlockNumber)))
			nodeInfo, err := d.getNodeInfo(client, event)
			if err != nil {
				logger.Error(err.Error())
				continue
			}
			err = d.HandleNodeRegister(event, block.Time(), &nodeInfo)
			if err != nil {
				logger.Error(err.Error())
				continue
			}
		default:
			continue
		}
	}

	for _, event := range eventsDelegation {
		eventName, ok1 := d.eventNameMap[event.Topics[0]]
		if !ok1 {
			continue
		}
		switch eventName {
		case "ModifyCommissionRate":
			logger.Info("Handle Delegation ModifyCommissionRate event")
			block, _ := client.BlockByNumber(context.Background(), big.NewInt(int64(event.BlockNumber)))
			err = d.HandleModifyCommissionRate(event, block.Time())
		case "NodeWithdraw":
			logger.Info("Handle Delegation NodeWithdraw event")
			err = d.HandleNodeWithdraw(event)
		case "ConfirmNodeReward":
			logger.Info("Handle Delegation ConfirmNodeReward event")
			err = d.HandleConfirmNodeReward(event)
		case "NodeDailyDelegations":
			logger.Info("Handle Delegation NodeDailyDelegations event")
			err = d.HandleNodeDailyDelegations(event)
		case "Delegate":
			logger.Info("Handle Delegation Delegate event")
			err = d.HandleDelegate(event)
		case "Undelegate":
			logger.Info("Handle Delegation Undelegate event")
			err = d.HandleUndelegate(event)
		case "Redelegate":
			logger.Info("Handle Delegation Redelegate event")
			err = d.HandleRedelegate(event)
		case "ClaimReward":
			logger.Info("Handle Delegation ClaimReward event")
			err = d.HandleClaimReward(event)
		default:
			continue
		}
		if err != nil {
			logger.Error(err.Error())
			continue
		}
	}

	if toBlock.Cmp(d.blockNumber) == 1 {
		newBlockNumber := new(big.Int).Add(toBlock, big.NewInt(1))
		d.blockNumber = newBlockNumber
		err = database.SetBlockNumber(newBlockNumber.Int64())
		if err != nil {
			logger.Error(err.Error())
		}
	}

	return nil
}

func (d *Dumper) unpack(log types.Log, contractIndex uint8, out interface{}) error {
	eventName := d.eventNameMap[log.Topics[0]]
	indexed := d.indexedMap[log.Topics[0]]

	err := d.contractABI[contractIndex].UnpackIntoInterface(out, eventName, log.Data)
	if err != nil {
		return err
	}

	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

func (d *Dumper) unpackLicenseTransfer(log types.Log) (string, string, string) {
	from := common.BytesToAddress(log.Topics[1].Bytes()).Hex()
	to := common.BytesToAddress(log.Topics[2].Bytes()).Hex()
	tokenID := new(big.Int).SetBytes(log.Topics[3].Bytes()).String()
	logger.Debug("unpack License-Transfer, from:", from, " to:", to, " tokenID:", tokenID)
	return from, to, tokenID
}

func (d *Dumper) getNodeInfo(client *ethclient.Client, log types.Log) (database.NodeInfoOnChain, error) {
	var nodeInfo database.NodeInfoOnChain
	node, err := d.GetNodeAddr(log)
	if err != nil {
		return nodeInfo, err
	}
	data, err := d.contractABI[3].Pack("getNodeInfo", node)
	if err != nil {
		return nodeInfo, err
	}
	callMsg := ethereum.CallMsg{
		To:   &(d.contractAddress[3]),
		Data: data,
	}
	res, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nodeInfo, err
	}

	temp, err := d.contractABI[3].Unpack("getNodeInfo", res)
	if err != nil {
		return nodeInfo, err
	}

	nodeInfo = *abi.ConvertType(temp[0], new(database.NodeInfoOnChain)).(*database.NodeInfoOnChain)
	logger.Info("node info:", nodeInfo)
	return nodeInfo, nil
}
