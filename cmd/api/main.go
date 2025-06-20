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
	// Load Configurations
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	supabaseDB, err := db.InitSupabase(cfg.SupabaseConnectionString)
	if err != nil {
		log.Fatalf("Failed to initialize Supabase DB: %v", err)
	}
	defer supabaseDB.Close()

	// Initialize FTP client for stocks
	ftpClient, err := ftp_client.NewFTPClient(cfg.NasdaqFTPAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer ftpClient.Close()

	// Set up repositories and services
	stockRepo := repositories.NewStockRepository(supabaseDB)
	stockService := services.NewStockService(ftpClient, stockRepo)

	cryptoRepo := repositories.NewCryptoRepository(supabaseDB)
	cryptoService := services.NewCryptoService(cryptoRepo, cfg)

	watchlistRepo := repositories.NewWatchlistRepository(supabaseDB)
	watchlistService := services.NewWatchlistService(watchlistRepo)

	// Initial data fetch if DBs are empty
	if err := stockRepo.LoadStockCache(); err != nil {
		log.Fatalf("Failed to load stock cache: %v", err)
	}
	if stocks := stockRepo.GetAllStocks(); len(stocks) == 0 {
		log.Println("Fetching stocks at startup...")
		if err := stockService.InitializeStocks(); err != nil {
			log.Fatalf("Failed to initialize stock data: %v", err)
		}
	}
	if err := cryptoRepo.LoadCryptoCache(); err != nil {
		log.Fatalf("Failed to load crypto cache: %v", err)
	}
	if crypto := cryptoRepo.GetAllCrypto(); len(crypto) == 0 {
		log.Println("Fetching Crypto at startup...")
		if err := cryptoService.InitializeCrypto(); err != nil {
			log.Fatalf("Failed to initialize crypto data: %v", err)
		}
	}

	// Daily update scheduler
	tasks.ScheduleDailyUpdates(stockService, cryptoService)

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

	// Handlers
	cryptoHandler := handlers.NewCryptoGinHandler(cryptoService)
	r.GET("/crypto", cryptoHandler.GetAllCrypto)
	r.GET("/crypto/:id", cryptoHandler.GetCryptoByID)
	r.POST("/crypto/price", cryptoHandler.GetCryptoPrice)
	r.POST("/crypto/history", cryptoHandler.GetCryptoHistory)
	r.POST("/crypto/history-ohlc", cryptoHandler.GetCryptoHistoryOHLC)

	stockHandler := handlers.NewStockGinHandler(stockService)
	r.GET("/stocks", stockHandler.GetAllStocks)
	r.GET("/stocks/:ticker", stockHandler.GetStockByTicker)

	watchlistHandler := handlers.NewWatchlistGinHandler(watchlistService)
	r.GET("/watchlists", watchlistHandler.GetAllWatchlists)
	r.GET("/watchlists/:id", watchlistHandler.GetWatchlistByID)
	r.POST("/create-watchlist", watchlistHandler.CreateWatchlist)
	r.PUT("/watchlists/:id", watchlistHandler.UpdateWatchlist)
	r.DELETE("/watchlists/:id", watchlistHandler.DeleteWatchlist)

	port := ":8080"
	log.Printf("Server running on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
