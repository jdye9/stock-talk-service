package repositories

import (
	"bufio"
	"log"
	"stock-talk-service/internal/models"
	"strings"

	"github.com/jlaffaye/ftp"
)

// FTPRepository handles operations related to FTP
type FTPRepository struct {
	ftpClient *ftp.ServerConn
	tickerStore *models.TickerData
}

// NewFTPRepository initializes a new FTPRepository with an existing FTP client
func NewFTPRepository(ftpClient *ftp.ServerConn, tickerStore *models.TickerData) *FTPRepository {
	return &FTPRepository{ftpClient: ftpClient, tickerStore: tickerStore }
}

// FetchFile retrieves and returns lines from the given FTP file
func (repo *FTPRepository) FetchFile(filename string) ([]string, error) {
	resp, err := repo.ftpClient.Retr(filename)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	var lines []string
	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ParseNasdaqListed parses NASDAQ listed tickers
func (repo *FTPRepository) ParseNasdaqListed(lines []string) []models.Ticker {
	var tickers []models.Ticker
	for i := 0; i < len(lines)-1; i++ {
		line := lines[i]
		if strings.Contains(line, "|") && !strings.HasPrefix(line, "Symbol|") {
			parts := strings.Split(line, "|")
			tickers = append(tickers, models.Ticker{Symbol: parts[0], Name: parts[1]})
		}
	}
	return tickers
}

// ParseOtherListed parses other listed tickers (NYSE, AMEX, etc.)
func (repo *FTPRepository) ParseOtherListed(lines []string) []models.Ticker {
	var tickers []models.Ticker
	for i := 0; i < len(lines)-1; i++ {
		line := lines[i]
		if strings.Contains(line, "|") && !strings.HasPrefix(line, "ACT Symbol|") {
			parts := strings.Split(line, "|")
			tickers = append(tickers, models.Ticker{Symbol: parts[0], Name: parts[1]})
		}
	}
	return tickers
}

func (repo *FTPRepository) FetchAndUpdateTickers() {
	nasdaqLines, err := repo.FetchFile("/SymbolDirectory/nasdaqlisted.txt")
	if err != nil {
		log.Printf("Error fetching nasdaqlisted.txt: %v", err)
		return
	}
	nasdaqTickers := repo.ParseNasdaqListed(nasdaqLines)

	otherLines, err := repo.FetchFile("/SymbolDirectory/otherlisted.txt")
	if err != nil {
		log.Printf("Error fetching otherlisted.txt: %v", err)
		return
	}
	otherTickers := repo.ParseOtherListed(otherLines)

	repo.tickerStore.Lock()
	repo.tickerStore.Nasdaq = nasdaqTickers
	repo.tickerStore.Other = otherTickers
	repo.tickerStore.Unlock()

	log.Printf("Tickers updated: NASDAQ(%d), OTHER(%d)", len(nasdaqTickers), len(otherTickers))
}
