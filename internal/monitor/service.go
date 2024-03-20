package monitor

import (
	"context"
	"math/big"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Monitor(client *ethclient.Client, req *protos.CreateMonitorRequest) (<-chan types.Log, func(), <-chan error) {
	contract := common.HexToAddress(req.Contract)
	topics := make([][]common.Hash, 1)
	topic := make([]common.Hash, len(req.Topics))
	for i, val := range req.Topics {
		topic[i] = common.HexToHash(val)
	}
	topics[0] = topic
	fromBlock := big.NewInt(0)
	fromBlock = fromBlock.SetUint64(req.FromBlock)
	tx := common.HexToHash(req.TxHash)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contract},
		Topics:    topics,
		FromBlock: fromBlock,
	}
	logs := make(chan types.Log)
	errChan := make(chan error)
	data := make(chan types.Log)
	var stop chan struct{}
	go func() {
		sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
		if err != nil {
			errChan <- err
			return
		}

		for {
			var t <-chan time.Time
			if stop == nil {
				t = time.After(time.Second * 2)
			}
			select {
			case <-stop:
				sub.Unsubscribe()
				close(errChan)
				close(stop)
				return
			case err := <-sub.Err():
				errChan <- err
				sub.Unsubscribe()
				return
			case vLog := <-logs:
				if vLog.TxHash == tx {
					data <- vLog
					return
				}
			case <-t:
				continue
			}
		}
	}()
	return data, func() {
		stop = make(chan struct{})
		stop <- struct{}{}
	}, errChan
}
