package tasks

import (
	"log"
	"stock-talk-service/internal/services"

	"github.com/robfig/cron/v3"
)

// ScheduleDailyUpdates sets up a cron job to fetch stocks at midnight every day
func ScheduleDailyUpdates(stockService *services.StockService, cryptoService *services.CryptoService) {
	c := cron.New()

	_, err := c.AddFunc("0 0 * * *", func() {
		log.Println("Running scheduled stock update...")
		stockService.FetchAndUpdateAllStocks()
		log.Println("Running scheduled crypto update...")
		cryptoService.FetchAndUpdateCrypto()
	})

	if err != nil {
		log.Fatalf("Failed to schedule daily stock update: %v", err)
	}

	c.Start()
}
