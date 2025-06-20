package handlers

import (
	"net/http"
	"stock-talk-service/internal/models"
	"stock-talk-service/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WatchlistHandler struct {
    watchlistService *services.WatchlistService
}

func NewWatchlistGinHandler(watchlistService *services.WatchlistService) *WatchlistHandler {
    return &WatchlistHandler{watchlistService: watchlistService}
}

// GET /watchlists
func (h *WatchlistHandler) GetAllWatchlists(ctx *gin.Context) {
    watchlists, err := h.watchlistService.GetAllWatchlists()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, watchlists)
}

// GET /watchlists/:id
func (h *WatchlistHandler) GetWatchlistByID(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    watchlist, err := h.watchlistService.GetWatchlistByID(id)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "watchlist not found"})
        return
    }
    ctx.JSON(http.StatusOK, watchlist)
}

// POST /watchlist
func (h *WatchlistHandler) CreateWatchlist(ctx *gin.Context) {
	var reqBody models.CreateWatchlistRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    watchlist := models.Watchlist{
		Name: reqBody.Name,
	}
	watchlistStocks := reqBody.Stocks
	watchlistCrypto := reqBody.Crypto

    if err := h.watchlistService.CreateWatchlist(&watchlist, &watchlistStocks, &watchlistCrypto); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusCreated, watchlist)
}

// PUT /watchlists/:id
func (h *WatchlistHandler) UpdateWatchlist(ctx *gin.Context) {
    id := ctx.Param("id")
    var w models.Watchlist
    if err := ctx.ShouldBindJSON(&w); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    w.Id = id
    if err := h.watchlistService.UpdateWatchlist(&w); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, w)
}

// DELETE /watchlists/:id
func (h *WatchlistHandler) DeleteWatchlist(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := h.watchlistService.DeleteWatchlist(id); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}