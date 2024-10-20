package route

import (
	"backend/api/controller"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, nftController *controller.NFTController, marketController *controller.MarketController) {
	api := r.Group("/api")
	{
		// NFT routes
		api.GET("/nft/:contractAddress", nftController.GetCollection)
		api.GET("/nft/:contractAddress/:tokenID", nftController.GetNFT)

		// Market routes
		api.GET("/order/:orderID", marketController.GetOrder)
	}
}
