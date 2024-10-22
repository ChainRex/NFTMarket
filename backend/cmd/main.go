package main

import (
	"backend/api/controller"
	"backend/api/route"
	"backend/repository"
	"backend/usecase"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 初始化数据库连接
	dsn := "root:123456@tcp(127.0.0.1:3306)/nftmarket?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	// 读取合约地址
	addressJSON, err := ioutil.ReadFile("contracts/NFTMarket-address.json")
	if err != nil {
		log.Fatalf("无法读取合约地址文件: %v", err)
	}
	var addressData struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(addressJSON, &addressData); err != nil {
		log.Fatalf("无法解析合约地址JSON: %v", err)
	}
	contractAddress := addressData.Address

	ethClientURL := "wss://polygon-amoy.g.alchemy.com/v2/oUhC0fClZFJKJ09zzWsqj65EFq3X01y0" // 替换为您的以太坊节点URL

	// 初始化仓储层
	nftRepo := repository.NewNFTRepository(db)
	marketRepo := repository.NewMarketRepository(db)

	// 初始化用例层
	nftUC := usecase.NewNFTUseCase(nftRepo, ethClientURL)
	defer nftUC.Close() // 确保在程序退出时关闭 NFTUseCase
	marketUC, err := usecase.NewMarketUseCase(marketRepo, nftRepo, nftUC, ethClientURL, contractAddress)
	if err != nil {
		log.Fatalf("初始化MarketUseCase失败: %v", err)
	}
	defer marketUC.Close() // 确保在程序退出时关闭 MarketUseCase

	// 初始化控制器
	nftController := controller.NewNFTController(nftUC)
	marketController := controller.NewMarketController(marketUC)

	// 初始化Gin路由
	r := gin.Default()

	// 设置路由
	route.SetupRoutes(r, nftController, marketController)

	// 启动服务器
	if err := r.Run("0.0.0.0:8081"); err != nil {
		log.Fatalf("无法启动服务器: %v", err)
	}
}
