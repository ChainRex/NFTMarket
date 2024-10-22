package usecase

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"backend/contracts"
	"backend/domain"
	"backend/repository"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type NFTUseCase struct {
	nftRepo       *repository.NFTRepository
	ethClientURL  string
	contractCache map[string]*contracts.NFTContract
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewNFTUseCase(nftRepo *repository.NFTRepository, ethClientURL string) *NFTUseCase {
	ctx, cancel := context.WithCancel(context.Background())
	return &NFTUseCase{
		nftRepo:       nftRepo,
		ethClientURL:  ethClientURL,
		contractCache: make(map[string]*contracts.NFTContract),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (uc *NFTUseCase) getNFTContract(contractAddress string) (*contracts.NFTContract, error) {
	uc.mutex.RLock()
	contract, exists := uc.contractCache[contractAddress]
	uc.mutex.RUnlock()

	if exists {
		return contract, nil
	}

	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	contract, err := contracts.NewNFTContract(uc.ethClientURL, contractAddress)
	if err != nil {
		return nil, err
	}

	uc.contractCache[contractAddress] = contract
	return contract, nil
}

func (uc *NFTUseCase) GetAllCollections() ([]domain.NFTCollection, error) {
	return uc.nftRepo.GetAllCollections()
}

func (uc *NFTUseCase) GetCollectionByAddress(contractAddress string) (*domain.NFTCollection, []domain.NFT, error) {
	collection, err := uc.nftRepo.GetCollectionByAddress(contractAddress)
	if err == nil {
		nfts, err := uc.nftRepo.GetNFTsByCollectionID(collection.ID)
		if err != nil {
			return nil, nil, err
		}
		return collection, nfts, nil
	}

	// 如果数据库中没有找到，尝试初始化
	err = uc.InitializeNFTCollection(contractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("该地址不是有效的NFT合约: %w", err)
	}

	// 再次尝试从数据库获取
	collection, err = uc.nftRepo.GetCollectionByAddress(contractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("该地址不是有效的NFT合约: %w", err)
	}
	nfts, err := uc.nftRepo.GetNFTsByCollectionID(collection.ID)
	if err != nil {
		return nil, nil, err
	}
	return collection, nfts, nil
}

func (uc *NFTUseCase) GetNFTByTokenID(contractAddress string, tokenID uint) (*domain.NFT, []domain.NFTAttribute, error) {
	nft, err := uc.nftRepo.GetByTokenID(contractAddress, tokenID)
	if err == nil {
		attributes, err := uc.nftRepo.GetAttributes(nft.ID)
		return nft, attributes, err
	}
	// 如果数据库中没有找到，先检查NFT系列是否存在
	_, err = uc.nftRepo.GetCollectionByAddress(contractAddress)
	if err != nil {
		// NFT系列不存在，初始化NFT系列
		err = uc.InitializeNFTCollection(contractAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化NFT系列失败: %w", err)
		}
	} else {
		// NFT系列存在，初始化单个NFT
		err = uc.InitializeNFT(contractAddress, tokenID)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化NFT失败: %w", err)
		}
	}

	// 再次尝试从数据库获取
	nft, err = uc.nftRepo.GetByTokenID(contractAddress, tokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("TokenID不存在: %w", err)
	}

	attributes, err := uc.nftRepo.GetAttributes(nft.ID)
	return nft, attributes, err
}

func (uc *NFTUseCase) InitializeNFT(contractAddress string, tokenID uint) error {
	nftContract, err := uc.getNFTContract(contractAddress)
	if err != nil {
		return fmt.Errorf("获取NFT合约实例失败: %w", err)
	}

	collection, err := uc.nftRepo.GetCollectionByAddress(contractAddress)
	if err != nil {
		return fmt.Errorf("获取NFT集合失败: %w", err)
	}

	tokenURI, err := nftContract.TokenURI(tokenID)
	if err != nil {
		return fmt.Errorf("获取TokenURI失败: %w", err)
	}

	owner, err := nftContract.OwnerOf(tokenID)
	if err != nil {
		return fmt.Errorf("获取NFT所有者失败: %w", err)
	}

	metadata, err := nftContract.GetNFTMetadata(tokenURI)
	if err != nil {
		return fmt.Errorf("获取NFT元数据失败: %w", err)
	}

	nft := &domain.NFT{
		CollectionID:    collection.ID,
		TokenID:         tokenID,
		Name:            metadata.Name,
		Description:     metadata.Description,
		Image:           metadata.Image,
		ContractAddress: contractAddress,
		Owner:           owner,
		TokenURI:        tokenURI,
	}

	if err := uc.nftRepo.UpsertNFT(nft); err != nil {
		return fmt.Errorf("创建NFT记录失败: %w", err)
	}

	for _, attr := range metadata.Attributes {
		attribute := &domain.NFTAttribute{
			NFTID:     nft.ID,
			TraitType: attr.TraitType,
			Value:     attr.Value,
		}
		if err := uc.nftRepo.SaveNFTAttribute(attribute); err != nil {
			return fmt.Errorf("创建NFT属性失败: %w", err)
		}
	}

	return nil
}

func (uc *NFTUseCase) InitializeNFTCollection(contractAddress string) error {
	// 获取 NFT 合约实例
	nftContract, err := uc.getNFTContract(contractAddress)
	if err != nil {
		return fmt.Errorf("获取 NFT 合约实例失败: %w", err)
	}

	// 获取或创建 NFT 集合
	collection, err := uc.nftRepo.GetCollectionByAddress(contractAddress)
	if err != nil {
		// 如果集合不存在，创建新的集合
		name, err := nftContract.Name()
		if err != nil {
			return fmt.Errorf("获取合约名称失败: %w", err)
		}
		symbol, err := nftContract.Symbol()
		if err != nil {
			return fmt.Errorf("获取合约符号失败: %w", err)
		}
		tokenIconURI, err := nftContract.TokenIconURI()
		if err != nil {
			log.Printf("获取 TokenIconURI 失败: %v", err)
			// 如果获取失败，设置为空字符串
			tokenIconURI = ""
		}
		collection = &domain.NFTCollection{
			ContractAddress: contractAddress,
			Name:            name,
			Symbol:          symbol,
			TokenIconURI:    tokenIconURI,
		}
		if err := uc.nftRepo.UpsertCollection(collection); err != nil {
			return fmt.Errorf("创建 NFT 集合失败: %w", err)
		}
	} else if collection.TokenIconURI == "" {
		// 如果集合存在但 TokenIconURI 为空，尝试从链上获取
		tokenIconURI, err := nftContract.TokenIconURI()
		if err != nil {
			log.Printf("获取 TokenIconURI 失败: %v", err)
		} else {
			collection.TokenIconURI = tokenIconURI
			if err := uc.nftRepo.UpsertCollection(collection); err != nil {
				log.Printf("更新 NFT 集合的 TokenIconURI 失败: %v", err)
			}
		}
	}

	// 获取总供应量
	totalSupply, err := nftContract.TotalSupply()
	if err != nil {
		return fmt.Errorf("获取总供应量失败: %w", err)
	}

	// 扫描历史事件并记录转移事件
	if err := uc.scanHistoricalEvents(contractAddress); err != nil {
		log.Printf("扫描历史事件失败: %v", err)
		// 继续执行,不中断整个初始化过程
	}

	// 初始化所有 NFT
	for i := uint(0); i < totalSupply; i++ {
		if err := uc.InitializeNFT(contractAddress, i); err != nil {
			log.Printf("初始化NFT失败 (TokenID: %d): %v", i, err)
		}
	}

	// 启动事件监听
	go uc.startEventListener(contractAddress)

	return nil
}

func (uc *NFTUseCase) startEventListener(contractAddress string) {
	nftContract, err := uc.getNFTContract(contractAddress)
	if err != nil {
		log.Printf("获取NFT合约实例失败: %v", err)
		return
	}

	eventChan := make(chan *types.Log)
	err = nftContract.WatchEvents(uc.ctx, eventChan)
	if err != nil {
		log.Printf("启动事件监听失败: %v", err)
		return
	}

	for {
		select {
		case event := <-eventChan:
			uc.handleNFTEvent(contractAddress, event)
		case <-uc.ctx.Done():
			return
		}
	}
}

func (uc *NFTUseCase) handleNFTEvent(contractAddress string, event *types.Log) {
	switch event.Topics[0].Hex() {
	case crypto.Keccak256Hash([]byte("MetadataUpdate(uint256)")).Hex():
		uc.handleMetadataUpdate(contractAddress, event)
	case crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")).Hex():
		uc.handleTransfer(contractAddress, event)
	}
}

func (uc *NFTUseCase) handleMetadataUpdate(contractAddress string, event *types.Log) {
	tokenID := new(big.Int).SetBytes(event.Data).Uint64()

	nftContract, err := uc.getNFTContract(contractAddress)
	if err != nil {
		log.Printf("获取NFT合约实例失败: %v", err)
		return
	}

	timestamp, err := nftContract.GetBlockTimestamp(event.BlockNumber)
	if err != nil {
		log.Printf("获取区块时间戳失败: %v", err)
		timestamp = 0 // 如果获取失败,使用0作为默认值
	}

	// 处理MetadataUpdate事件作为mint事件
	transferEvent := &domain.NFTTransferEvent{
		ContractAddress: contractAddress,
		TokenID:         uint(tokenID),
		EventType:       "mint",
		FromAddress:     common.HexToAddress("0x0000000000000000000000000000000000000000").Hex(),
		ToAddress:       event.Address.Hex(),
		TransactionHash: event.TxHash.Hex(),
		BlockNumber:     uint(event.BlockNumber),
		BlockTimestamp:  time.Unix(int64(timestamp), 0),
	}

	if err := uc.nftRepo.SaveNFTTransferEvent(transferEvent); err != nil {
		log.Printf("保存NFT mint事件失败: %v", err)
	}

	// 更新NFT元数据
	nftContract, err = uc.getNFTContract(contractAddress)
	if err != nil {
		log.Printf("获取NFT合约实例失败: %v", err)
		return
	}

	tokenURI, err := nftContract.TokenURI(uint(tokenID))
	if err != nil {
		log.Printf("获取TokenURI失败: %v", err)
		return
	}

	metadata, err := nftContract.GetNFTMetadata(tokenURI)
	if err != nil {
		log.Printf("获取NFT元数据失败: %v", err)
		return
	}

	owner, err := nftContract.OwnerOf(uint(tokenID))
	if err != nil {
		log.Printf("获取NFT所有者失败: %v", err)
		return
	}

	nft := &domain.NFT{
		ContractAddress: contractAddress,
		TokenID:         uint(tokenID),
		Owner:           owner,
		TokenURI:        tokenURI,
		Name:            metadata.Name,
		Description:     metadata.Description,
		Image:           metadata.Image,
	}

	if err := uc.nftRepo.UpsertNFT(nft); err != nil {
		log.Printf("更新或插入NFT失败: %v", err)
		return
	}

	for _, attr := range metadata.Attributes {
		attribute := &domain.NFTAttribute{
			NFTID:     nft.ID,
			TraitType: attr.TraitType,
			Value:     attr.Value,
		}
		err = uc.nftRepo.SaveNFTAttribute(attribute)
		if err != nil {
			log.Printf("保存NFT属性失败: %v", err)
		}
	}
}

func (uc *NFTUseCase) handleTransfer(contractAddress string, event *types.Log) {
	from := common.HexToAddress(event.Topics[1].Hex())
	to := common.HexToAddress(event.Topics[2].Hex())
	tokenID := new(big.Int).SetBytes(event.Topics[3].Bytes()).Uint64()

	nftContract, err := uc.getNFTContract(contractAddress)
	if err != nil {
		log.Printf("获取NFT合约实例失败: %v", err)
		return
	}

	timestamp, err := nftContract.GetBlockTimestamp(event.BlockNumber)
	if err != nil {
		log.Printf("获取区块时间戳失败: %v", err)
		timestamp = 0 // 如果获取失败,使用0作为默认值
	}

	transferEvent := &domain.NFTTransferEvent{
		ContractAddress: contractAddress,
		TokenID:         uint(tokenID),
		EventType:       "transfer",
		FromAddress:     from.Hex(),
		ToAddress:       to.Hex(),
		TransactionHash: event.TxHash.Hex(),
		BlockNumber:     uint(event.BlockNumber),
		BlockTimestamp:  time.Unix(int64(timestamp), 0),
	}

	if from == common.HexToAddress("0x0000000000000000000000000000000000000000") {
		transferEvent.EventType = "mint"
	}

	if err := uc.nftRepo.SaveNFTTransferEvent(transferEvent); err != nil {
		log.Printf("保存NFT transfer事件失败: %v", err)
	}

	// 更新NFT所有者
	err = uc.nftRepo.UpdateNFTOwner(contractAddress, uint(tokenID), to.Hex())
	if err != nil {
		log.Printf("更新NFT所有者失败: %v", err)
	}
}

// 获取NFT的转移历史
func (uc *NFTUseCase) GetNFTTransferHistory(contractAddress string, tokenID uint) ([]domain.NFTTransferEvent, error) {
	return uc.nftRepo.GetNFTTransferEvents(contractAddress, tokenID)
}

// 获取NFT的当前所有者
func (uc *NFTUseCase) GetNFTCurrentOwner(contractAddress string, tokenID uint) (string, error) {
	latestEvent, err := uc.nftRepo.GetLatestNFTTransferEvent(contractAddress, tokenID)
	if err != nil {
		return "", fmt.Errorf("获取最新转移事件失败: %w", err)
	}
	return latestEvent.ToAddress, nil
}

func (uc *NFTUseCase) Close() {
	uc.cancel()
}

func (uc *NFTUseCase) scanHistoricalEvents(contractAddress string) error {
	nftContract, err := uc.getNFTContract(contractAddress)
	if err != nil {
		return fmt.Errorf("获取NFT合约实例失败: %w", err)
	}

	creationBlock, err := nftContract.GetCreationBlockNumber()
	if err != nil {
		return fmt.Errorf("获取合约创建区块号失败: %w", err)
	}

	latestBlock, err := nftContract.GetLatestBlockNumber()
	if err != nil {
		return fmt.Errorf("获取最新区块号失败: %w", err)
	}

	transferFilter := [][]common.Hash{{nftContract.GetTransferEventID()}}
	logs, err := nftContract.FilterLogs(big.NewInt(int64(creationBlock)), big.NewInt(int64(latestBlock)), transferFilter)
	if err != nil {
		return fmt.Errorf("过滤Transfer事件失败: %w", err)
	}

	for _, log := range logs {
		uc.handleTransfer(contractAddress, &log)
	}

	return nil
}
