package handlers

import (
	"net/http"
	"stock-talk-service/internal/models"
	"stock-talk-service/internal/services"

	"github.com/gin-gonic/gin"
)

type CryptoGinHandler struct {
	Service *services.CryptoService
}

func NewCryptoGinHandler(service *services.CryptoService) *CryptoGinHandler {
	return &CryptoGinHandler{Service: service}
}

// GET /crypto
func (h *CryptoGinHandler) GetAllCrypto(ctx *gin.Context) {
	crypto := h.Service.GetAllCrypto()
	ctx.JSON(http.StatusOK, crypto)
}

// GET /crypto/:id
func (h *CryptoGinHandler) GetCryptoByID(ctx *gin.Context) {
	id := ctx.Param("id")
	crypto, ok := h.Service.GetCryptoByID(id)
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "crypto not found"})
		return
	}
	ctx.JSON(http.StatusOK, crypto)
}

func (h *CryptoGinHandler) GetCryptoPrice(ctx *gin.Context) {
	var req models.CryptoPriceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Service.GetCryptoPrice(req.CoinIDs, req.VsCurrencies)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (h *CryptoGinHandler) GetCryptoHistory(ctx *gin.Context) {
	var req models.CryptoHistoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Service.GetCryptoHistory(req.CoinIDs, req.VsCurrency, req.Days, req.Interval)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (h *CryptoGinHandler) GetCryptoHistoryOHLC(ctx *gin.Context) {
	var req models.CryptoHistoryOHLCRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Service.GetCryptoHistoryOHLC(req.CoinIDs, req.VsCurrency, req.Days, req.Interval)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, result)
}