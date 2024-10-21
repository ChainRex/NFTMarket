package route

import (
	"backend/api/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, nftController *controller.NFTController, marketController *controller.MarketController) {
	// 设置 CORS
	r.Use(cors.Default())

	api := r.Group("/api")
	{
		// NFT routes
		api.GET("/nft", nftController.GetCollections)
		api.GET("/nft/:contractAddress", nftController.GetCollection)
		api.GET("/nft/:contractAddress/:tokenID", nftController.GetNFT)
		api.GET("/nft/:contractAddress/:tokenID/history", nftController.GetNFTTransferHistory)
		// Market routes
		api.GET("/orders", marketController.GetOrders)
		api.GET("/order/:contractAddress/:tokenID", marketController.GetOrderByNFT)
	}
}
