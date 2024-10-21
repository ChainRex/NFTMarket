package domain

import "time"

// NFTCollection 表示NFT系列
type NFTCollection struct {
	ID              uint   `gorm:"primaryKey;autoIncrement"`
	ContractAddress string `gorm:"uniqueIndex"`
	Name            string
	Symbol          string
	TokenIconURI    string
}

// NFT 表示单个NFT
type NFT struct {
	ID              uint   `gorm:"primaryKey;autoIncrement"`
	CollectionID    uint   `gorm:"index"`
	TokenID         uint   `gorm:"uniqueIndex:idx_collection_token"`
	ContractAddress string `gorm:"uniqueIndex:idx_collection_token"`
	Owner           string `gorm:"index"`
	TokenURI        string
	Name            string
	Description     string
	Image           string
}

// NFTAttribute 表示NFT的属性
type NFTAttribute struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	NFTID     uint `gorm:"index"`
	TraitType string
	Value     string
}

// Order 表示订单
type Order struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement"`
	NFTContractAddress string `gorm:"index"`
	TokenID            uint   `gorm:"index"`
	TokenAddress       string
	Price              string
	Seller             string `gorm:"index"`
	Status             uint   // 0: 未售出, 1: 已售出, 2: 已取消
}

// NFTTransferEvent 表示NFT的转移事件(包括mint和transfer)
type NFTTransferEvent struct {
	ID              uint   `gorm:"primaryKey;autoIncrement"`
	ContractAddress string `gorm:"index:idx_contract_token,priority:1"`
	TokenID         uint   `gorm:"index:idx_contract_token,priority:2"`
	EventType       string `gorm:"type:enum('mint','transfer')"`
	FromAddress     string `gorm:"index"`
	ToAddress       string `gorm:"index"`
	TransactionHash string
	BlockNumber     uint `gorm:"index"`
	BlockTimestamp  time.Time
}
