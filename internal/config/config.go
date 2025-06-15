package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    NasdaqFTPAddress string
    StockDB string
	CoingeckoBaseUrl string
	CryptoDB string
	SupabaseConnectionString string
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
	stockDB := os.Getenv("STOCK_DB")
	cryptoDB := os.Getenv("CRYPTO_DB")

	// Coingecko API
	coingeckoBaseUrl := os.Getenv("COINGECKO_BASE_URL")

	// Supabase connection string
	supabaseConnectionString := os.Getenv("SUPABASE_CONNECTION_STRING")

    return &Config{
        NasdaqFTPAddress: nasdaqFTPAddress,
		StockDB: stockDB,
		CoingeckoBaseUrl: coingeckoBaseUrl,
		CryptoDB: cryptoDB,
		SupabaseConnectionString: supabaseConnectionString,
    }, nil
}