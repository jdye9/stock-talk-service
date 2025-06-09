package main

import (
	"log"
	"stock-talk-service/internal/controllers"
	"stock-talk-service/internal/repositories"
	"stock-talk-service/internal/services"
	"stock-talk-service/internal/tasks"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize shared data
	tickerStore := InitTickerData()

	// Initialize FTP client
	ftpClient, err := FTPConnect("ftp.nasdaqtrader.com:21")
	if err != nil {
		log.Printf("Error connecting to FTP: %v", err)
		return
	}
	defer ftpClient.Quit()

	ftpRepo := repositories.NewFTPRepository(ftpClient, tickerStore)
	ftpService := services.NewFTPService(ftpRepo)

	// Start daily updates in the background
	tasks.ScheduleDailyUpdates(ftpService)

	// Setup Gin server
	r := gin.Default()
	
	r.GET("/tickers", func(ctx *gin.Context) {
		controllers.TickersHandler(ctx, tickerStore)
	})


	port := ":8080"
	log.Printf("Server running on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
