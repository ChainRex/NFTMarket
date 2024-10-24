package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"backend/contracts/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NFTContract struct {
	contract *bind.BoundContract
	client   *ethclient.Client
	address  common.Address
	abi      abi.ABI
}

type NFTMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Attributes  []struct {
		TraitType string `json:"trait_type"`
		Value     string `json:"value"`
	} `json:"attributes"`
}

func NewNFTContract(ethClientURL, contractAddress string) (*NFTContract, error) {
	client, err := ethclient.Dial(ethClientURL)
	if err != nil {
		return nil, fmt.Errorf("连接以太坊客户端失败: %w", err)
	}

	// 读取NFT.json文件
	abiFile, err := ioutil.ReadFile("contracts/NFT.json")
	if err != nil {
		return nil, fmt.Errorf("读取NFT ABI文件失败: %v", err)
	}

	var abiData struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(abiFile, &abiData); err != nil {
		return nil, fmt.Errorf("解析NFT ABI JSON失败: %v", err)
	}

	nftABI, err := abi.JSON(strings.NewReader(string(abiData.ABI)))
	if err != nil {
		return nil, fmt.Errorf("解析NFT ABI失败: %v", err)
	}

	address := common.HexToAddress(contractAddress)
	contract := bind.NewBoundContract(address, nftABI, client, client, client)

	return &NFTContract{
		contract: contract,
		client:   client,
		address:  address,
		abi:      nftABI,
	}, nil
}

func (c *NFTContract) callMethod(method string, args ...interface{}) ([]interface{}, error) {
	return utils.CallMethod(c.client, c.abi, c.address, method, args...)
}

func (c *NFTContract) Name() (string, error) {
	result, err := c.callMethod("name")
	if err != nil {
		return "", err
	}
	return result[0].(string), nil
}

func (c *NFTContract) Symbol() (string, error) {
	result, err := c.callMethod("symbol")
	if err != nil {
		return "", err
	}
	return result[0].(string), nil
}

func (c *NFTContract) TokenIconURI() (string, error) {
	result, err := c.callMethod("tokenIconURI")
	if err != nil {
		return "", err
	}
	return utils.ConvertIPFSToHTTP(result[0].(string)), nil
}

func (c *NFTContract) TotalSupply() (uint, error) {
	result, err := c.callMethod("totalSupply")
	if err != nil {
		return 0, err
	}
	return uint(result[0].(*big.Int).Uint64()), nil
}

func (c *NFTContract) TokenURI(tokenID uint) (string, error) {
	result, err := c.callMethod("tokenURI", big.NewInt(int64(tokenID)))
	if err != nil {
		return "", err
	}
	return utils.ConvertIPFSToHTTP(result[0].(string)), nil
}

func (c *NFTContract) OwnerOf(tokenID uint) (string, error) {
	result, err := c.callMethod("ownerOf", big.NewInt(int64(tokenID)))
	if err != nil {
		return "", fmt.Errorf("获取NFT所有者失败: %w", err)
	}
	return result[0].(common.Address).Hex(), nil
}

func (c *NFTContract) GetNFTMetadata(tokenURI string) (*NFTMetadata, error) {
	httpURI := utils.ConvertIPFSToHTTP(tokenURI)
	resp, err := http.Get(httpURI)
	if err != nil {
		return nil, fmt.Errorf("获取NFT元数据失败: %w", err)
	}
	defer resp.Body.Close()

	var metadata NFTMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("解析NFT元数据失败: %w", err)
	}

	// 转换 Image URL
	metadata.Image = utils.ConvertIPFSToHTTP(metadata.Image)

	return &metadata, nil
}

func (c *NFTContract) WatchEvents(ctx context.Context, eventChan chan<- *types.Log) error {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{c.address},
	}

	logs := make(chan types.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("订阅事件失败: %w", err)
	}

	go func() {
		for {
			select {
			case err := <-sub.Err():
				fmt.Printf("事件订阅错误: %v\n", err)
				return
			case vLog := <-logs:
				eventChan <- &vLog
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (c *NFTContract) GetTransferEvents(fromBlock, toBlock *big.Int) ([]*types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{c.address},
		Topics: [][]common.Hash{{
			c.abi.Events["Transfer"].ID,
		}},
	}

	logs, err := c.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("获取Transfer事件失败: %w", err)
	}

	return convertLogsToPointers(logs), nil
}

func convertLogsToPointers(logs []types.Log) []*types.Log {
	result := make([]*types.Log, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}
	return result
}

func (c *NFTContract) GetCreationBlockNumber() (uint64, error) {
	// 这里我们使用一个较低的区块号作为起始点
	// 实际应用中,你可能需要根据具体情况调整这个值
	startBlock := uint64(12489691)
	currentBlock, err := c.client.BlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("获取当前区块号失败: %w", err)
	}

	for startBlock < currentBlock {
		midBlock := (startBlock + currentBlock) / 2
		code, err := c.client.CodeAt(context.Background(), c.address, big.NewInt(int64(midBlock)))
		if err != nil {
			return 0, fmt.Errorf("获取合约代码失败: %w", err)
		}

		if len(code) > 0 {
			currentBlock = midBlock
		} else {
			startBlock = midBlock + 1
		}
	}

	return startBlock, nil
}

func (c *NFTContract) GetLatestBlockNumber() (uint64, error) {
	return c.client.BlockNumber(context.Background())
}

func (c *NFTContract) GetTransferEventID() common.Hash {
	return c.abi.Events["Transfer"].ID
}

func (c *NFTContract) FilterLogs(fromBlock, toBlock *big.Int, topics [][]common.Hash) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{c.address},
		Topics:    topics,
	}

	return c.client.FilterLogs(context.Background(), query)
}

func (c *NFTContract) GetBlockTimestamp(blockNumber uint64) (uint64, error) {
	block, err := c.client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		return 0, fmt.Errorf("获取区块信息失败: %w", err)
	}
	return block.Time(), nil
}
