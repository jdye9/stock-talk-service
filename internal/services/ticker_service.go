package services

import (
	"log"
	"stock-talk-service/internal/ftp_client"
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
