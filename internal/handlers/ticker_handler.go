package handlers

import (
	"net/http"
	"stock-talk-service/internal/services"

	"github.com/gin-gonic/gin"
)

type TickerGinHandler struct {
	Service *services.TickerService
}

func NewTickerGinHandler(service *services.TickerService) *TickerGinHandler {
	return &TickerGinHandler{Service: service}
}

// GET /tickers
func (h *TickerGinHandler) GetAllTickers(ctx *gin.Context) {
	nasdaq := h.Service.GetNasdaqTickers()
	other := h.Service.GetOtherTickers()
	ctx.JSON(http.StatusOK, gin.H{
		"nasdaq": nasdaq,
		"other":  other,
	})
}

// GET /tickers/nasdaq/:symbol
func (h *TickerGinHandler) GetNasdaqTickerBySymbol(ctx *gin.Context) {
	symbol := ctx.Param("symbol")
	ticker, ok := h.Service.GetNasdaqTickerBySymbol(symbol)
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ticker not found"})
		return
	}
	ctx.JSON(http.StatusOK, ticker)
}

// GET /tickers/other/:symbol
func (h *TickerGinHandler) GetOtherTickerBySymbol(ctx *gin.Context) {
	symbol := ctx.Param("symbol")
	ticker, ok := h.Service.GetOtherTickerBySymbol(symbol)
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ticker not found"})
		return
	}
	ctx.JSON(http.StatusOK, ticker)
}
