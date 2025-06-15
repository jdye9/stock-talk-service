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
	ftpClient  *ftp_client.FTPClient
	stockRepo *repositories.StockRepository
}

func NewStockService(ftpClient *ftp_client.FTPClient, stockRepo *repositories.StockRepository) *StockService {
	return &StockService{ftpClient: ftpClient, stockRepo: stockRepo}
}

func (s *StockService) GetAllStocks() []models.Stock {
    return s.stockRepo.GetAllStocks()
}

func (s *StockService) GetStockByTicker(ticker string) (models.Stock, bool) {
    return s.stockRepo.GetStockByTicker(ticker)
}

func (s *StockService) SaveStocks(stocks []models.Stock) error {
    return s.stockRepo.SaveStocks(stocks)
}

func (s *StockService) FetchAndUpdateAllStocks() {
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
        if err != nil {
            log.Printf("Error parsing %s: %v", src.Path, err)
            continue
        }


        allStocks = append(allStocks, stocks...)
        log.Printf("%s stocks updated: %d", src.Exchange, len(stocks))
		file.Close()
    }

    if err := s.stockRepo.SaveStocks(allStocks); err != nil {
        log.Printf("Error saving stocks: %v", err)
        return
    }

    log.Printf("All stocks updated: %d", len(allStocks))
}