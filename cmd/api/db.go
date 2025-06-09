package main

import (
	"fmt"
	"stock-talk-service/internal/models"
)

func InitTickerData() *models.TickerData {
	fmt.Println("HERE")
	tickerStore := &models.TickerData{}
	return tickerStore
}