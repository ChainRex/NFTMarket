package contracts

import (
	"backend/domain"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NFTMarketContract struct {
	client  *ethclient.Client
	address common.Address
	abi     abi.ABI
}

func NewNFTMarketContract(ethClientURL, contractAddress string) (*NFTMarketContract, error) {
	client, err := ethclient.Dial(ethClientURL)
	if err != nil {
		return nil, fmt.Errorf("连接以太坊客户端失败: %w", err)
	}

	abiJSON, err := ioutil.ReadFile("contracts/NFTMarket-abi.json")
	if err != nil {
		return nil, fmt.Errorf("读取ABI文件失败: %w", err)
	}

	var abiData struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(abiJSON, &abiData); err != nil {
		return nil, fmt.Errorf("解析ABI JSON失败: %w", err)
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiData.ABI)))
	if err != nil {
		return nil, fmt.Errorf("解析ABI失败: %w", err)
	}

	return &NFTMarketContract{
		client:  client,
		address: common.HexToAddress(contractAddress),
		abi:     contractABI,
	}, nil
}

func (c *NFTMarketContract) GetOrders() ([]domain.Order, error) {
	data, err := c.abi.Pack("getOrders")
	if err != nil {
		return nil, fmt.Errorf("打包getOrders函数调用失败: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}

	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("调用getOrders函数失败: %w", err)
	}

	var orders []struct {
		NFT     common.Address
		TokenID *big.Int
		Token   common.Address
		Price   *big.Int
		Seller  common.Address
		Status  *big.Int
	}

	if err := c.abi.UnpackIntoInterface(&orders, "getOrders", result); err != nil {
		return nil, fmt.Errorf("解析getOrders返回结果失败: %w", err)
	}

	domainOrders := make([]domain.Order, len(orders))
	for i, order := range orders {
		domainOrders[i] = domain.Order{
			NFTContractAddress: order.NFT.Hex(),
			TokenID:            uint(order.TokenID.Uint64()),
			TokenAddress:       order.Token.Hex(),
			Price:              order.Price.String(),
			Seller:             order.Seller.Hex(),
			Status:             uint(order.Status.Uint64()),
		}
	}

	return domainOrders, nil
}

func (c *NFTMarketContract) WatchEvents(ctx context.Context, eventChan chan<- *types.Log) error {
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
