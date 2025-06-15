package repositories

import (
	"database/sql"
	"fmt"
	"stock-talk-service/internal/models"
	"strings"
	"sync"
)

type StockRepository struct {
    db *sql.DB
    cache      map[string]models.Stock // ticker -> Stock
    cacheMutex sync.RWMutex
}

func NewStockRepository(db *sql.DB) *StockRepository {
	return &StockRepository{
		db:    db,
		cache: make(map[string]models.Stock),
	}
}

func (r *StockRepository) LoadStockCache() error {
    rows, err := r.db.Query("SELECT id, ticker, name FROM stock")
    if err != nil {
        return err
    }
    defer rows.Close()

    cache := make(map[string]models.Stock)
    for rows.Next() {
        var s models.Stock
        if err := rows.Scan(&s.Id, &s.Ticker, &s.Name); err != nil {
            return err
        }
        cache[s.Id] = s
    }

    r.cacheMutex.Lock()
    r.cache = cache
    r.cacheMutex.Unlock()
    return rows.Err()
}

func (r *StockRepository) GetAllStocks() []models.Stock {
    r.cacheMutex.RLock()
    defer r.cacheMutex.RUnlock()
    stocks := make([]models.Stock, 0, len(r.cache))
    for _, s := range r.cache {
        stocks = append(stocks, s)
    }
    return stocks
}

func (r *StockRepository) GetStockByTicker(ticker string) (models.Stock, bool) {
    r.cacheMutex.RLock()
    defer r.cacheMutex.RUnlock()
    s, ok := r.cache[ticker]
    return s, ok
}

func (r *StockRepository) SaveStocks(stocks []models.Stock) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.Exec("DELETE FROM stock")
    if err != nil {
        return err
    }

    batchSize := 1000
	total := len(stocks)

	for start := 0; start < total; start += batchSize {
		end := start + batchSize
		if end > total {
			end = total
		}
		batch := stocks[start:end]

		var (
			args        []interface{}
			placeholders []string
		)
		for i, s := range batch {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			args = append(args, s.Ticker, s.Name)
		}

		query := "INSERT INTO stock (ticker, name) VALUES " + strings.Join(placeholders, ",")
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

    if err := tx.Commit(); err != nil {
        return err
    }
    return r.LoadStockCache()
}