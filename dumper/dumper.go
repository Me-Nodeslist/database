package dumper

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	com "github.com/memoio/contractsv2/common"
	"github.com/Me-Nodeslist/database/database"
	"github.com/Me-Nodeslist/database/logs"
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
		// store:        store,
		eventNameMap: make(map[common.Hash]string),
		indexedMap:   make(map[common.Hash]abi.Arguments),
	}

	_, endpoint := com.GetInsEndPointByChain(chain)
	dumper.endpoint = endpoint

	dumper.contractAddress = []common.Address{addrs.LicenseNFT, addrs.DelMEMO, addrs.Settlement, addrs.Delegation}

	licenseContractABI, err := abi.JSON(strings.NewReader(""))
	if err != nil {
		return dumper, err
	}

	delMemoContractABI, err := abi.JSON(strings.NewReader(""))
	if err != nil {
		return dumper, err
	}

	settlementContractABI, err := abi.JSON(strings.NewReader(""))
	if err != nil {
		return dumper, err
	}

	delegationContractABI, err := abi.JSON(strings.NewReader(""))
	if err != nil {
		return dumper, err
	}

	dumper.contractABI = []abi.ABI{licenseContractABI, delMemoContractABI, settlementContractABI, delegationContractABI}

	for name, event := range dumper.contractABI[0].Events {
		dumper.eventNameMap[event.ID] = name

		var indexed abi.Arguments
		for _, arg := range dumper.contractABI[0].Events[name].Inputs {
			if arg.Indexed {
				indexed = append(indexed, arg)
			}
		}
		dumper.indexedMap[event.ID] = indexed
	}

	for name, event := range dumper.contractABI[1].Events {
		dumper.eventNameMap[event.ID] = name

		var indexed abi.Arguments
		for _, arg := range dumper.contractABI[1].Events[name].Inputs {
			if arg.Indexed {
				indexed = append(indexed, arg)
			}
		}
		dumper.indexedMap[event.ID] = indexed
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

	eventsLicenseNFT, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		Addresses: []common.Address{d.contractAddress[0]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	eventsDelMEMO, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		Addresses: []common.Address{d.contractAddress[1]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	eventsSettlement, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		Addresses: []common.Address{d.contractAddress[2]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	eventsDelegation, err := client.FilterLogs(context.TODO(), ethereum.FilterQuery{
		FromBlock: d.blockNumber,
		Addresses: []common.Address{d.contractAddress[3]},
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	lastBlockNumber := d.blockNumber

	for _, event := range eventsLicenseNFT {
		eventName, ok1 := d.eventNameMap[event.Topics[0]]
		if !ok1 {
			continue
		}
		switch eventName {
		default:
			continue
		}
		if err != nil {
			logger.Error(err.Error())
			break
		}

		d.blockNumber = big.NewInt(int64(event.BlockNumber) + 1)
	}

	if d.blockNumber.Cmp(lastBlockNumber) == 1 {
		database.SetBlockNumber(d.blockNumber.Int64())
	}

	return nil
}

func (d *Dumper) unpack(log types.Log, contractType uint8, out interface{}) error {
	eventName := d.eventNameMap[log.Topics[0]]
	indexed := d.indexedMap[log.Topics[0]]
	switch contractType {
	case 0:
		err := d.contractABI[0].UnpackIntoInterface(out, eventName, log.Data)
		if err != nil {
			return err
		}
	default:
		err := d.contractABI[1].UnpackIntoInterface(out, eventName, log.Data)
		if err != nil {
			return err
		}
	}

	return abi.ParseTopics(out, indexed, log.Topics[1:])
}
