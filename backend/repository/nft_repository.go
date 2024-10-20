package repository

import (
	"backend/domain"

	"gorm.io/gorm"
)

type NFTRepository struct {
	db *gorm.DB
}

func NewNFTRepository(db *gorm.DB) *NFTRepository {
	return &NFTRepository{db: db}
}

func (r *NFTRepository) GetByTokenID(contractAddress string, tokenID uint) (*domain.NFT, error) {
	var nft domain.NFT
	err := r.db.Where("contract_address = ? AND token_id = ?", contractAddress, tokenID).First(&nft).Error
	return &nft, err
}

func (r *NFTRepository) GetAttributes(nftID uint) ([]domain.NFTAttribute, error) {
	var attributes []domain.NFTAttribute
	err := r.db.Where("nft_id = ?", nftID).Find(&attributes).Error
	return attributes, err
}

func (r *NFTRepository) GetCollectionByAddress(contractAddress string) (*domain.NFTCollection, error) {
	var collection domain.NFTCollection
	err := r.db.Where("contract_address = ?", contractAddress).First(&collection).Error
	return &collection, err
}

func (r *NFTRepository) SaveCollection(collection *domain.NFTCollection) error {
	return r.db.Create(collection).Error
}

func (r *NFTRepository) SaveNFT(nft *domain.NFT) error {
	return r.db.Create(nft).Error
}

func (r *NFTRepository) SaveNFTAttribute(attribute *domain.NFTAttribute) error {
	return r.db.Create(attribute).Error
}

// 新增方法
func (r *NFTRepository) ClearNFTCollections() error {
	return r.db.Exec("TRUNCATE TABLE nft_collections").Error
}

func (r *NFTRepository) ClearNFTs() error {
	return r.db.Exec("TRUNCATE TABLE nfts").Error
}

func (r *NFTRepository) ClearNFTAttributes() error {
	return r.db.Exec("TRUNCATE TABLE nft_attributes").Error
}

// 更新或插入NFT集合
func (r *NFTRepository) UpsertCollection(collection *domain.NFTCollection) error {
	return r.db.Where(domain.NFTCollection{ContractAddress: collection.ContractAddress}).
		Assign(*collection).
		FirstOrCreate(collection).Error
}

// 更新或插入NFT
func (r *NFTRepository) UpsertNFT(nft *domain.NFT) error {
	return r.db.Where(domain.NFT{ContractAddress: nft.ContractAddress, TokenID: nft.TokenID}).
		Assign(*nft).
		FirstOrCreate(nft).Error
}

// 更新NFT所有者
func (r *NFTRepository) UpdateNFTOwner(contractAddress string, tokenID uint, newOwner string) error {
	return r.db.Model(&domain.NFT{}).
		Where("contract_address = ? AND token_id = ?", contractAddress, tokenID).
		Update("owner", newOwner).Error
}

// 获取所有NFT
func (r *NFTRepository) GetAllNFTs(contractAddress string) ([]domain.NFT, error) {
	var nfts []domain.NFT
	err := r.db.Where("contract_address = ?", contractAddress).Find(&nfts).Error
	return nfts, err
}
