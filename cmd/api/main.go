package main

import (
	"fmt"
	"log"
	"stock-talk-service/internal/config"
	"stock-talk-service/internal/db"
	"stock-talk-service/internal/ftp_client"
	"stock-talk-service/internal/handlers"
	"stock-talk-service/internal/repositories"
	"stock-talk-service/internal/services"
	"stock-talk-service/internal/tasks"

	"github.com/gin-gonic/gin"
)

func main() {

	//load .env variables
	cfg, err := config.Load()
	if err != nil {
    	log.Fatal(err)
		return
	}

	// Init Ticker DB
	tickerDB, err := db.InitSQLite(cfg.TickerDB)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer tickerDB.Close()

	// Initialize FTP client
	ftpAddress := fmt.Sprint(cfg.NasdaqFTPAddress)
	ftpClient, err := ftp_client.NewFTPClient(ftpAddress)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ftpClient.Close()

	// Initialize Ticker Repository and Ticker Service
	tickerRepo := repositories.NewTickerRepository(tickerDB)
	tickerService := services.NewTickerService(ftpClient, tickerRepo)

	// On startup: fetch from FTP if DB is empty
	nasdaqTickers, err := tickerRepo.GetNasdaqTickers()
	if err != nil || len(nasdaqTickers) == 0 {
		log.Println("Fetching Nasdaq tickers at startup...")
		tickerService.FetchAndUpdateNasdaqTickers()
	}

	otherTickers, err := tickerRepo.GetOtherTickers()
	if err != nil || len(otherTickers) == 0 {
		log.Println("Fetching Other tickers at startup...")
		tickerService.FetchAndUpdateOtherTickers()
	}

	// schedule daily ticker updates in the background
	tasks.ScheduleDailyUpdates(tickerService)

	// Setup Gin server
	r := gin.Default()
	
	r.GET("/tickers", func(ctx *gin.Context) {
		handlers.GetTickersHandler(ctx, tickerService)
	})

	port := ":8080"
	log.Printf("Server running on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
