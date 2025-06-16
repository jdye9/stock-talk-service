package tasks

import (
	"log"
	"stock-talk-service/internal/services"

	"github.com/robfig/cron/v3"
)

// ScheduleDailyUpdates sets up a cron job to fetch stocks at midnight every day
func ScheduleDailyUpdates(stockService *services.StockService, cryptoService *services.CryptoService) {
	c := cron.New()

	_, err := c.AddFunc("@every 1m", func() {
		log.Println("Running scheduled stock update...")
		errS := stockService.FetchAndUpdateAllStocks()
		if errS != nil {
			log.Printf("Error fetching stocks: %v", errS)
		}
		log.Println("Running scheduled crypto update...")
		errC := cryptoService.FetchAndUpdateAllCrypto()
		if errC != nil {
			log.Printf("Error fetching crypto: %v", errC)
		}
	})

	if err != nil {
		log.Fatalf("Failed to schedule daily update: %v", err)
	}

	c.Start()
}
