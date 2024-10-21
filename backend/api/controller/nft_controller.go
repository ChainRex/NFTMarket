package controller

import (
	"backend/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NFTController struct {
	useCase *usecase.NFTUseCase
}

func NewNFTController(useCase *usecase.NFTUseCase) *NFTController {
	return &NFTController{useCase: useCase}
}

func (c *NFTController) GetCollection(ctx *gin.Context) {
	contractAddress := ctx.Param("contractAddress")

	collection, nfts, err := c.useCase.GetCollectionByAddress(contractAddress)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "NFT系列未找到"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"collection": collection,
		"nfts":       nfts,
	})
}

func (c *NFTController) GetCollections(ctx *gin.Context) {
	collections, err := c.useCase.GetAllCollections()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取NFT系列失败"})
		return
	}
	ctx.JSON(http.StatusOK, collections)
}

func (c *NFTController) GetNFT(ctx *gin.Context) {
	contractAddress := ctx.Param("contractAddress")
	tokenID, err := strconv.ParseUint(ctx.Param("tokenID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的tokenID"})
		return
	}

	nft, attributes, err := c.useCase.GetNFTByTokenID(contractAddress, uint(tokenID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "NFT未找到"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"nft":        nft,
		"attributes": attributes,
	})
}

func (c *NFTController) GetNFTTransferHistory(ctx *gin.Context) {
	contractAddress := ctx.Param("contractAddress")
	tokenID, err := strconv.ParseUint(ctx.Param("tokenID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的tokenID"})
		return
	}

	history, err := c.useCase.GetNFTTransferHistory(contractAddress, uint(tokenID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取转移历史失败"})
		return
	}

	ctx.JSON(http.StatusOK, history)
}
