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

func (c *MarketController) GetOrder(ctx *gin.Context) {
	orderID, err := strconv.ParseUint(ctx.Param("orderID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
		return
	}

	order, err := c.useCase.GetOrderByID(uint(orderID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "订单未找到"})
		return
	}

	ctx.JSON(http.StatusOK, order)
}
