package controllers

import (
	"net/http"
	"stock-talk-service/internal/models"

	"github.com/gin-gonic/gin"
)

func TickersHandler(ctx *gin.Context, tickerStore *models.TickerData) {
	tickerStore.RLock()
	defer tickerStore.RUnlock()

	response := struct {
		Nasdaq []models.Ticker`json:"nasdaq"`
		Other  []models.Ticker `json:"other"`
	}{
		Nasdaq: tickerStore.Nasdaq,
		Other:  tickerStore.Other,
	}

	ctx.JSON(http.StatusOK, response)
}
