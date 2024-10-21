package usecase

import (
	"backend/contracts"
	"backend/domain"
	"backend/repository"
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type MarketUseCase struct {
	repo     *repository.MarketRepository
	nftRepo  *repository.NFTRepository
	contract *contracts.NFTMarketContract
	nftUC    *NFTUseCase
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewMarketUseCase(repo *repository.MarketRepository, nftRepo *repository.NFTRepository, nftUC *NFTUseCase, ethClientURL, contractAddress string) (*MarketUseCase, error) {
	contract, err := contracts.NewNFTMarketContract(ethClientURL, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("创建NFTMarketContract失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	uc := &MarketUseCase{
		repo:     repo,
		nftRepo:  nftRepo,
		contract: contract,
		nftUC:    nftUC,
		ctx:      ctx,
		cancel:   cancel,
	}

	// 初始化数据库
	if err := uc.InitializeOrders(); err != nil {
		cancel()
		return nil, fmt.Errorf("初始化订单数据失败: %w", err)
	}

	// 启动事件监听协程
	go uc.startEventListener()

	return uc, nil
}

func (uc *MarketUseCase) startEventListener() {
	eventChan := make(chan *types.Log)
	go func() {
		if err := uc.contract.WatchEvents(uc.ctx, eventChan); err != nil {
			log.Printf("监听事件失败: %v", err)
			return
		}
	}()

	for {
		select {
		case event := <-eventChan:
			if err := uc.HandleEvent(event); err != nil {
				log.Printf("处理事件失败: %v", err)
			}
		case <-uc.ctx.Done():
			return
		}
	}
}

func (uc *MarketUseCase) Close() {
	uc.cancel()
}

func (uc *MarketUseCase) GetOrderByID(id uint) (*domain.Order, error) {
	return uc.repo.GetOrderByID(id + 1)
}

func (uc *MarketUseCase) GetAllOrders() ([]domain.Order, error) {
	return uc.repo.GetAllOrders()
}

func (uc *MarketUseCase) InitializeOrders() error {
	// 清空 orders 表
	if err := uc.repo.ClearOrders(); err != nil {
		return fmt.Errorf("清空 orders 表失败: %w", err)
	}

	// 清空 NFTs 和 NFT 属性表，但保留 NFT 集合表
	if err := uc.nftRepo.ClearNFTs(); err != nil {
		return fmt.Errorf("清空 NFTs 表失败: %w", err)
	}
	if err := uc.nftRepo.ClearNFTAttributes(); err != nil {
		return fmt.Errorf("清空 NFT 属性表失败: %w", err)
	}

	// 清空事件转移表
	if err := uc.nftRepo.ClearNFTTransferEvents(); err != nil {
		return fmt.Errorf("清空 NFT 转移事件表失败: %w", err)
	}

	// 从合约获取订单
	orders, err := uc.contract.GetOrders()
	if err != nil {
		return fmt.Errorf("从合约获取订单失败: %w", err)
	}

	// 用于存储唯一的NFT合约地址
	nftContracts := make(map[string]bool)

	// 批量插入订单
	if err := uc.repo.BatchInsertOrders(orders); err != nil {
		return fmt.Errorf("批量插入订单失败: %w", err)
	}

	// 收集所有涉及到的NFT合约地址
	for _, order := range orders {
		nftContracts[order.NFTContractAddress] = true
	}

	// 获取现有的 NFT 集合
	existingCollections, err := uc.nftRepo.GetAllCollections()
	if err != nil {
		return fmt.Errorf("获取现有 NFT 集合失败: %w", err)
	}

	// 将现有集合添加到 nftContracts 中
	for _, collection := range existingCollections {
		nftContracts[collection.ContractAddress] = true
	}

	// 初始化每个NFT合约
	for contractAddress := range nftContracts {
		// 初始化NFT合约
		if err := uc.nftUC.InitializeNFTCollection(contractAddress); err != nil {
			log.Printf("初始化NFT合约失败 (地址: %s): %v", contractAddress, err)
			// 继续初始化其他合约，不中断整个过程
		}
	}

	return nil
}

func (uc *MarketUseCase) GetOrderByNFT(contractAddress string, tokenID uint) (*domain.Order, error) {
	return uc.repo.GetOrderByNFT(contractAddress, tokenID)
}

// 定义事件签名常量
const (
	OrderCreatedSignature        = "OrderCreated(uint256,address,uint256,address,uint256,address)"
	OrderCancelledSignature      = "OrderCancelled(uint256)"
	OrderFulfilledSignature      = "OrderFulfilled(uint256,address)"
	NFTContractDeployedSignature = "NFTContractDeployed(address,string,string)"
)

func (uc *MarketUseCase) HandleEvent(event *types.Log) error {
	switch event.Topics[0] {
	case crypto.Keccak256Hash([]byte(OrderCreatedSignature)):
		return uc.handleOrderCreated(event)
	case crypto.Keccak256Hash([]byte(OrderCancelledSignature)):
		return uc.handleOrderCancelled(event)
	case crypto.Keccak256Hash([]byte(OrderFulfilledSignature)):
		return uc.handleOrderFulfilled(event)
	case crypto.Keccak256Hash([]byte(NFTContractDeployedSignature)):
		return uc.handleNFTContractDeployed(event)
	default:
		return fmt.Errorf("未知的事件类型")
	}
}

func (uc *MarketUseCase) handleOrderCreated(event *types.Log) error {
	orderId := new(big.Int).SetBytes(event.Topics[1].Bytes()).Uint64()
	nftAddress := common.HexToAddress(event.Topics[2].Hex())
	tokenId := new(big.Int).SetBytes(event.Topics[3].Bytes()).Uint64()

	data := event.Data
	token := common.BytesToAddress(data[:32])
	price := new(big.Int).SetBytes(data[32:64])
	seller := common.BytesToAddress(data[64:])

	// 检查并初始化NFT合约
	if err := uc.nftUC.InitializeNFTCollection(nftAddress.Hex()); err != nil {
		log.Printf("检查并初始化NFT合约失败: %v", err)
		// 继续处理订单，不中断流程
	}

	order := domain.Order{
		ID:                 uint(orderId + 1),
		NFTContractAddress: nftAddress.Hex(),
		TokenID:            uint(tokenId),
		TokenAddress:       token.Hex(),
		Price:              price.String(),
		Seller:             seller.Hex(),
		Status:             0,
	}

	return uc.repo.BatchInsertOrders([]domain.Order{order})
}

func (uc *MarketUseCase) handleOrderCancelled(event *types.Log) error {
	orderId := new(big.Int).SetBytes(event.Topics[1].Bytes()).Uint64()
	return uc.repo.UpdateOrderStatus(uint(orderId+1), 2)
}

func (uc *MarketUseCase) handleOrderFulfilled(event *types.Log) error {
	orderId := new(big.Int).SetBytes(event.Topics[1].Bytes()).Uint64()
	return uc.repo.UpdateOrderStatus(uint(orderId+1), 1)
}

func (uc *MarketUseCase) handleNFTContractDeployed(event *types.Log) error {
	nftAddress := common.HexToAddress(event.Topics[1].Hex())

	// 初始化NFT合约
	if err := uc.nftUC.InitializeNFTCollection(nftAddress.Hex()); err != nil {
		return fmt.Errorf("初始化NFT合约失败: %w", err)
	}

	return nil
}
