package usecase

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

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

func (uc *NFTUseCase) GetCollectionByAddress(contractAddress string) (*domain.NFTCollection, error) {
	collection, err := uc.nftRepo.GetCollectionByAddress(contractAddress)
	if err == nil {
		return collection, nil
	}

	// 如果数据库中没有找到，尝试初始化
	err = uc.InitializeNFTCollection(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("该地址不是有效的NFT合约: %w", err)
	}

	// 再次尝试从数据库获取
	return uc.nftRepo.GetCollectionByAddress(contractAddress)
}

func (uc *NFTUseCase) GetNFTByTokenID(contractAddress string, tokenID uint) (*domain.NFT, []domain.NFTAttribute, error) {
	nft, err := uc.nftRepo.GetByTokenID(contractAddress, tokenID)
	if err == nil {
		attributes, err := uc.nftRepo.GetAttributes(nft.ID)
		return nft, attributes, err
	}

	// 如果数据库中没有找到，尝试初始化整个集合
	err = uc.InitializeNFTCollection(contractAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("该地址不是有效的NFT合约或者TokenID不存在: %w", err)
	}

	// 再次尝试从数据库获取
	nft, err = uc.nftRepo.GetByTokenID(contractAddress, tokenID)
	if err != nil {
		return nil, nil, fmt.Errorf("TokenID不存在: %w", err)
	}

	attributes, err := uc.nftRepo.GetAttributes(nft.ID)
	return nft, attributes, err
}

func (uc *NFTUseCase) InitializeNFTCollection(contractAddress string) error {
	// ... 保留现有的初始化逻辑

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
	tokenID := new(big.Int).SetBytes(event.Topics[1].Bytes()).Uint64()

	nftContract, err := uc.getNFTContract(contractAddress)
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

	err = uc.nftRepo.UpsertNFT(nft)
	if err != nil {
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
	to := common.HexToAddress(event.Topics[2].Hex())
	tokenID := new(big.Int).SetBytes(event.Topics[3].Bytes()).Uint64()

	err := uc.nftRepo.UpdateNFTOwner(contractAddress, uint(tokenID), to.Hex())
	if err != nil {
		log.Printf("更新NFT所有者失败: %v", err)
	}
}

func (uc *NFTUseCase) Close() {
	uc.cancel()
}
