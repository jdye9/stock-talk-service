package tasks

import (
	"log"
	"stock-talk-service/internal/services"
	"time"
)

func ScheduleDailyUpdates(ftpService *services.FTPService) {
	go func() {
		for {
			log.Println("Fetching data in the background...")
			ftpService.ProcessUpdateTickers()
			time.Sleep(24 * time.Hour)
		}
	}()
}
