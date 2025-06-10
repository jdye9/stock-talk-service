package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    NasdaqFTPAddress string
    TickerDB string
}

func Load() (*Config, error) {
    // Load .env file if present
    err := godotenv.Load("../../.env")
    if err != nil {
        // .env file not found or can't be read
        fmt.Println("Warning: .env file not loaded:", err)
        // Not necessarily fatal; env vars could come from elsewhere
    }

	// NASDAQ FTP
    nasdaqFTPAddress := os.Getenv("NASDAQ_FTP_ADDRESS")

	// SQLITE DB
	tickerDB := os.Getenv("TICKER_DB")

    return &Config{
        NasdaqFTPAddress: nasdaqFTPAddress,
		TickerDB: tickerDB,
    }, nil
}