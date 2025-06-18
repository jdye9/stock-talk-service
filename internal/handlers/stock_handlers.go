package handlers

import (
	"net/http"
	"stock-talk-service/internal/services"

	"github.com/gin-gonic/gin"
)

type StockGinHandler struct {
    Service *services.StockService
}

func NewStockGinHandler(service *services.StockService) *StockGinHandler {
    return &StockGinHandler{Service: service}
}

// GET /stocks
func (h *StockGinHandler) GetAllStocks(ctx *gin.Context) {
    stocks := h.Service.GetAllStocks()
    ctx.JSON(http.StatusOK, stocks)
}

// Optionally, you can keep these for single ticker lookups if needed:

// GET /stocks/:ticker
func (h *StockGinHandler) GetStockByTicker(ctx *gin.Context) {
    id := ctx.Param("id")
    stock, ok := h.Service.GetStockById(id)
    if !ok {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
        return
    }
    ctx.JSON(http.StatusOK, stock)
}