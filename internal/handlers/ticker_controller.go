package handlers

import (
	"fmt"
	"net/http"
	"stock-talk-service/internal/models"
	"stock-talk-service/internal/services"

	"github.com/gin-gonic/gin"
)

func GetTickersHandler(ctx *gin.Context, tickerService *services.TickerService) {
	nasdaqTickers, err := tickerService.TickerRepo.GetNasdaqTickers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve NASDAQ tickers",
		})
		return
	}

	otherTickers, err := tickerService.TickerRepo.GetOtherTickers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve other tickers",
		})
		return
	}

	fmt.Println(nasdaqTickers)

	response := struct {
		Nasdaq []models.Ticker `json:"nasdaq"`
		Other  []models.Ticker `json:"other"`
	}{
		Nasdaq: nasdaqTickers,
		Other:  otherTickers,
	}

	ctx.JSON(http.StatusOK, response)
}

