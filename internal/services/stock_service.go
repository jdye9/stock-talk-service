package services

import (
	"io"
	"log"
	"stock-talk-service/internal/ftp_client"
	"stock-talk-service/internal/models"
	"stock-talk-service/internal/repositories"
	"stock-talk-service/internal/utils"
)

type StockService struct {
	ftpClient *ftp_client.FTPClient
	stockRepo *repositories.StockRepository
}

func NewStockService(ftpClient *ftp_client.FTPClient, stockRepo *repositories.StockRepository) *StockService {
	return &StockService{ftpClient: ftpClient, stockRepo: stockRepo}
}

func (s *StockService) GetAllStocks() []models.Stock {
	return s.stockRepo.GetAllStocks()
}

func (s *StockService) GetStockById(id string) (models.Stock, bool) {
	return s.stockRepo.GetStockById(id)
}

func (s *StockService) SaveStocksInitialLoad(stocks []models.Stock) error {
	return s.stockRepo.SaveStocksInitialLoad(stocks)
}

func (s *StockService) SaveStocksWithReview(stocks []models.Stock) error {
	return s.stockRepo.SaveStocksWithReview(stocks)
}

// Fetch from FTP and parse both NASDAQ and other listed, return combined slice
func (s *StockService) FetchAllStocks() ([]models.Stock, error) {
	type source struct {
		Path      string
		Exchange  string
		ParseFunc func(file io.ReadCloser) ([]models.Stock, error)
	}

	sources := []source{
		{
			Path:      "/SymbolDirectory/nasdaqlisted.txt",
			Exchange:  "NASDAQ",
			ParseFunc: func(file io.ReadCloser) ([]models.Stock, error) {
				return utils.ParseNasdaqListed(file)
			},
		},
		{
			Path:      "/SymbolDirectory/otherlisted.txt",
			Exchange:  "OTHER",
			ParseFunc: func(file io.ReadCloser) ([]models.Stock, error) {
				return utils.ParseOtherListed(file)
			},
		},
	}

	var allStocks []models.Stock

	for _, src := range sources {
		file, err := s.ftpClient.RetrieveFile(src.Path)
		if err != nil {
			log.Printf("Error fetching %s: %v", src.Path, err)
			continue
		}

		stocks, err := src.ParseFunc(file)
		file.Close()
		if err != nil {
			log.Printf("Error parsing %s: %v", src.Path, err)
			continue
		}

		allStocks = append(allStocks, stocks...)
		log.Printf("%s stocks fetched: %d", src.Exchange, len(stocks))
	}

	return allStocks, nil
}

// Wrapper to fetch and save initial load
func (s *StockService) InitializeStocks() error {
	stocks, err := s.FetchAllStocks()
	if err != nil {
		return err
	}
	return s.SaveStocksInitialLoad(stocks)
}

// Wrapper to fetch and save with review
func (s *StockService) FetchAndUpdateAllStocks() error {
	stocks, err := s.FetchAllStocks()
	if err != nil {
		return err
	}
	return s.SaveStocksWithReview(stocks)
}
