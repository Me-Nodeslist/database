package dumper

import (
	"context"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/Me-Nodeslist/database/database"
	"github.com/Me-Nodeslist/database/logs"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	com "github.com/memoio/contractsv2/common"
)

type ContractAddress struct {
	LicenseNFT common.Address
	DelMEMO    common.Address
	Settlement common.Address
	Delegation common.Address
}

var (
	// blockNumber = big.NewInt(0)
	logger = logs.Logger("dumper")
)

type Dumper struct {
	endpoint        string
	contractABI     []abi.ABI
	contractAddress []common.Address

	blockNumber *big.Int

	eventNameMap map[common.Hash]string
	indexedMap   map[common.Hash]abi.Arguments
}

func NewDumper(chain string, addrs *ContractAddress) (dumper *Dumper, err error) {
	dumper = &Dumper{
		eventNameMap: make(map[common.Hash]string),
		indexedMap:   make(map[common.Hash]abi.Arguments),
	}

	_, endpoint := com.GetInsEndPointByChain(chain)
	dumper.endpoint = endpoint

	dumper.contractAddress = []common.Address{addrs.LicenseNFT, addrs.DelMEMO, addrs.Settlement, addrs.Delegation}

	licenseNFTABI, err := os.ReadFile("../abi/LicenseNFT.abi")
	if err != nil {
		logger.Error("Failed to read licenseNFT abi file, ", err)
		return dumper, err
	}
	licenseContractABI, err := abi.JSON(strings.NewReader(string(licenseNFTABI)))
	if err != nil {
		return dumper, err
	}

	delMEMOABI, err := os.ReadFile("../abi/DelMEMO.abi")
	if err != nil {
		logger.Error("Failed to read delMEMO abi file, ", err)
		return dumper, err
	}
	delMemoContractABI, err := abi.JSON(strings.NewReader(string(delMEMOABI)))
	if err != nil {
		return dumper, err
	}

	settlementABI, err := os.ReadFile("../abi/Settlement.abi")
	if err != nil {
		logger.Error("Failed to read settlement abi file, ", err)
		return dumper, err
	}
	settlementContractABI, err := abi.JSON(strings.NewReader(string(settlementABI)))
	if err != nil {
		return dumper, err
	}

	delegationABI, err := os.ReadFile("../abi/Delegation.abi")
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
	// var last *big.Int
	for {
		d.Dump()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(10 * time.Second):
		}
	}
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
			break
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
			err = d.HandleDelMemoRedeem(event)
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
			break
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
			break
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
			err = d.HandleDelegationModifyCommissionRate(event)
		case "NodeWithdraw":
			logger.Info("Handle Delegation NodeWithdraw event")
			err = d.HandleDelegationNodeWithdraw(event)
		case "ConfirmNodeReward":
			logger.Info("Handle Delegation ConfirmNodeReward event")
			err = d.HandleDelegationConfirmNodeReward(event)
		case "NodeDailyDelegations":
			logger.Info("Handle Delegation NodeDailyDelegations event")
			err = d.HandleDelegationNodeDailyDelegations(event)
		case "Delegate":
			logger.Info("Handle Delegation Delegate event")
			err = d.HandleDelegationDelegate(event)
		case "Undelegate":
			logger.Info("Handle Delegation Undelegate event")
			err = d.HandleDelegationUndelegate(event)
		case "Redelegate":
			logger.Info("Handle Delegation Redelegate event")
			err = d.HandleDelegationRedelegate(event)
		case "ClaimReward":
			logger.Info("Handle Delegation ClaimReward event")
			err = d.HandleDelegationClaimReward(event)
		case "NodeRegister":
			logger.Info("Handle Delegation NodeRegister event")
			err = d.HandleDelegationNodeRegister(event)
		default:
			continue
		}
		if err != nil {
			logger.Error(err.Error())
			break
		}
	}

	if toBlock.Cmp(d.blockNumber) == 1 {
		database.SetBlockNumber(toBlock.Int64())
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
