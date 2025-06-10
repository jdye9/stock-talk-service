package utils

import (
	"bufio"
	"io"
	"stock-talk-service/internal/models"
	"strings"
)

func ParseNasdaqListed(r io.Reader) ([]models.Ticker, error) {
    scanner := bufio.NewScanner(r)
    var tickers []models.Ticker
    var lines []string

    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    for i := 0; i < len(lines)-1; i++ {
        line := lines[i]
        if strings.Contains(line, "|") && !strings.HasPrefix(line, "Symbol|") {
            parts := strings.Split(line, "|")
            if len(parts) >= 2 {
                tickers = append(tickers, models.Ticker{Symbol: parts[0], Name: parts[1]})
            }
        }
    }

    return tickers, nil
}


func ParseOtherListed(r io.Reader) ([]models.Ticker, error) {
    scanner := bufio.NewScanner(r)
    var tickers []models.Ticker
    var lines []string

    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    // Exclude the last line
    for i := 0; i < len(lines)-1; i++ {
        line := lines[i]
        if strings.Contains(line, "|") && !strings.HasPrefix(line, "ACT Symbol|") {
            parts := strings.Split(line, "|")
            if len(parts) >= 2 {
                tickers = append(tickers, models.Ticker{Symbol: parts[0], Name: parts[1]})
            }
        }
    }

    return tickers, nil
}
