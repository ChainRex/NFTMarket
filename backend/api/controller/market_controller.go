package controller

import (
	"backend/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MarketController struct {
	useCase *usecase.MarketUseCase
}

func NewMarketController(useCase *usecase.MarketUseCase) *MarketController {
	return &MarketController{useCase: useCase}
}

func (c *MarketController) GetOrders(ctx *gin.Context) {
	orders, err := c.useCase.GetAllOrders()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单失败"})
		return
	}
	ctx.JSON(http.StatusOK, orders)
}

func (c *MarketController) GetOrderByNFT(ctx *gin.Context) {
	contractAddress := ctx.Param("contractAddress")
	tokenID, err := strconv.ParseUint(ctx.Param("tokenID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的TokenID"})
		return
	}

	order, err := c.useCase.GetOrderByNFT(contractAddress, uint(tokenID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "订单未找到"})
		return
	}

	ctx.JSON(http.StatusOK, order)
}
