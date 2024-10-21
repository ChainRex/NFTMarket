package repository

import (
	"backend/domain"

	"gorm.io/gorm"
)

type MarketRepository struct {
	db *gorm.DB
}

func NewMarketRepository(db *gorm.DB) *MarketRepository {
	return &MarketRepository{db: db}
}

func (r *MarketRepository) GetOrderByID(id uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.First(&order, id).Error
	return &order, err
}

func (r *MarketRepository) GetOrderByNFT(contractAddress string, tokenID uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.Where("nft_contract_address = ? AND token_id = ?", contractAddress, tokenID).Order("id DESC").First(&order).Error
	return &order, err
}

func (r *MarketRepository) GetAllOrders() ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Find(&orders).Error
	return orders, err
}

func (r *MarketRepository) ClearOrders() error {
	return r.db.Exec("TRUNCATE TABLE orders").Error
}

func (r *MarketRepository) BatchInsertOrders(orders []domain.Order) error {
	return r.db.CreateInBatches(orders, 100).Error
}
func (r *MarketRepository) UpdateOrderStatus(id uint, status uint) error {
	return r.db.Model(&domain.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *MarketRepository) CreateNFTCollection(collection domain.NFTCollection) error {
	return r.db.Create(&collection).Error
}
