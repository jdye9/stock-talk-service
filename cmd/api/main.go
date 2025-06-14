package main

import (
	"log"
	"stock-talk-service/internal/config"
	"stock-talk-service/internal/db"
	"stock-talk-service/internal/ftp_client"
	"stock-talk-service/internal/handlers"
	"stock-talk-service/internal/repositories"
	"stock-talk-service/internal/services"
	"stock-talk-service/internal/tasks"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Init stocks DB
	stocksDB, err := db.InitSQLite("../../data/stocks.db", db.StockSchemas)
	if err != nil {
		log.Fatalf("Failed to initialize stocks DB: %v", err)
	}
	defer stocksDB.Close()

	// Init crypto DB
	cryptoDB, err := db.InitSQLite("../../data/crypto.db", db.CryptoSchemas)
	if err != nil {
		log.Fatalf("Failed to initialize crypto DB: %v", err)
	}
	defer cryptoDB.Close()

	// FTP + service initialization
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ftpClient, err := ftp_client.NewFTPClient(cfg.NasdaqFTPAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer ftpClient.Close()

	tickerRepo := repositories.NewTickerRepository(stocksDB)
	tickerService := services.NewTickerService(ftpClient, tickerRepo)

	cryptoRepo := repositories.NewCryptoRepository(cryptoDB)
	cryptoService := services.NewCryptoService(cryptoRepo, cfg)

	// Initial data fetch if DBs are empty
	if nasdaq := tickerRepo.GetNasdaqTickers(); len(nasdaq) == 0 {
		log.Println("Fetching Nasdaq tickers at startup...")
		tickerService.FetchAndUpdateNasdaqTickers()
	}
	if other := tickerRepo.GetOtherTickers(); len(other) == 0 {
		log.Println("Fetching Other tickers at startup...")
		tickerService.FetchAndUpdateOtherTickers()
	}
	if crypto := cryptoRepo.GetAllCrypto(); len(crypto) == 0 {
		log.Println("Fetching Crypto at startup...")
		cryptoService.FetchAndUpdateCrypto()
	}

	// Daily update scheduler
	tasks.ScheduleDailyUpdates(tickerService, cryptoService)

	// Gin HTTP server setup
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	cryptoHandler := handlers.NewCryptoGinHandler(cryptoService)
	
	r.GET("/crypto", cryptoHandler.GetAllCrypto)
	r.GET("/crypto/:id", cryptoHandler.GetCryptoByID)
	r.POST("/crypto/price", cryptoHandler.GetCryptoPrice)
	r.POST("/crypto/history", cryptoHandler.GetCryptoHistory)
	r.POST("/crypto/history-ohlc", cryptoHandler.GetCryptoHistoryOHLC)

	tickerHandler := handlers.NewTickerGinHandler(tickerService)
	r.GET("/tickers", tickerHandler.GetAllTickers)
	r.GET("/tickers/nasdaq/:symbol", tickerHandler.GetNasdaqTickerBySymbol)
	r.GET("/tickers/other/:symbol", tickerHandler.GetOtherTickerBySymbol)

	port := ":8080"
	log.Printf("Server running on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
