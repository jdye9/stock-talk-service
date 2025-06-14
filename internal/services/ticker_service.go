package services

import (
	"log"
	"stock-talk-service/internal/ftp_client"
	"stock-talk-service/internal/models"
	"stock-talk-service/internal/repositories"
	"stock-talk-service/internal/utils"
)

type TickerService struct {
	ftpClient  *ftp_client.FTPClient
	TickerRepo *repositories.TickerRepository
}

func NewTickerService(ftpClient *ftp_client.FTPClient, tickerRepo *repositories.TickerRepository) *TickerService {
	return &TickerService{ftpClient: ftpClient, TickerRepo: tickerRepo}
}

// GetNasdaqTickers returns all NASDAQ tickers from the in-memory cache.
func (s *TickerService) GetNasdaqTickers() []models.Ticker {
	return s.TickerRepo.GetNasdaqTickers()
}

// GetOtherTickers returns all Other tickers from the in-memory cache.
func (s *TickerService) GetOtherTickers() []models.Ticker {
	return s.TickerRepo.GetOtherTickers()
}

// GetNasdaqTickerBySymbol returns a NASDAQ ticker by symbol from the in-memory cache.
func (s *TickerService) GetNasdaqTickerBySymbol(symbol string) (models.Ticker, bool) {
	return s.TickerRepo.GetNasdaqTickerBySymbol(symbol)
}

// GetOtherTickerBySymbol returns an Other ticker by symbol from the in-memory cache.
func (s *TickerService) GetOtherTickerBySymbol(symbol string) (models.Ticker, bool) {
	return s.TickerRepo.GetOtherTickerBySymbol(symbol)
}

// SaveNasdaqTickers replaces all NASDAQ tickers in DB and refreshes the cache.
func (s *TickerService) SaveNasdaqTickers(tickers []models.Ticker) error {
	return s.TickerRepo.SaveNasdaqTickers(tickers)
}

// SaveOtherTickers replaces all Other tickers in DB and refreshes the cache.
func (s *TickerService) SaveOtherTickers(tickers []models.Ticker) error {
	return s.TickerRepo.SaveOtherTickers(tickers)
}

// ReloadNasdaqCache reloads the NASDAQ cache from the DB.
func (s *TickerService) ReloadNasdaqCache() error {
	return s.TickerRepo.LoadNasdaqCache()
}

// ReloadOtherCache reloads the Other cache from the DB.
func (s *TickerService) ReloadOtherCache() error {
	return s.TickerRepo.LoadOtherCache()
}

func (s *TickerService) FetchAndUpdateNasdaqTickers() {
	file, err := s.ftpClient.RetrieveFile("/SymbolDirectory/nasdaqlisted.txt")
	if err != nil {
		log.Printf("Error fetching nasdaqlisted.txt: %v", err)
		return
	}
	defer file.Close()

	tickers, err := utils.ParseNasdaqListed(file)
	if err != nil {
		log.Printf("Error parsing nasdaqlisted.txt: %v", err)
		return
	}

	if err := s.TickerRepo.SaveNasdaqTickers(tickers); err != nil {
		log.Printf("Error saving NASDAQ tickers: %v", err)
		return
	}

	log.Printf("NASDAQ tickers updated: %d", len(tickers))
}

func (s *TickerService) FetchAndUpdateOtherTickers() {
	file, err := s.ftpClient.RetrieveFile("/SymbolDirectory/otherlisted.txt")
	if err != nil {
		log.Printf("Error fetching otherlisted.txt: %v", err)
		return
	}
	defer file.Close()

	tickers, err := utils.ParseOtherListed(file)
	if err != nil {
		log.Printf("Error parsing otherlisted.txt: %v", err)
		return
	}

	if err := s.TickerRepo.SaveOtherTickers(tickers); err != nil {
		log.Printf("Error saving OTHER tickers: %v", err)
		return
	}

	log.Printf("OTHER tickers updated: %d", len(tickers))
}
