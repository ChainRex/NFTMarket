package domain

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
