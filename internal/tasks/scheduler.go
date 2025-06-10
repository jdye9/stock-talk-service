package tasks

import (
	"log"
	"stock-talk-service/internal/services"

	"github.com/robfig/cron/v3"
)

// ScheduleDailyUpdates sets up a cron job to fetch tickers at midnight every day
func ScheduleDailyUpdates(tickerService *services.TickerService) {
	c := cron.New()

	_, err := c.AddFunc("0 0 * * *", func() {
		log.Println("Running scheduled ticker update...")
		tickerService.FetchAndUpdateNasdaqTickers()
		tickerService.FetchAndUpdateOtherTickers()
	})

	if err != nil {
		log.Fatalf("Failed to schedule daily ticker update: %v", err)
	}

	c.Start()
}
