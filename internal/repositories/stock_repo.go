package repositories

import (
	"database/sql"
	"fmt"
	"stock-talk-service/internal/models"
	"strings"
	"sync"
	"time"
)

type StockRepository struct {
	db         *sql.DB
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
		cache[s.Ticker] = s
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

// Initial load: Insert all stocks, mark active, no manual review
func (r *StockRepository) SaveStocksInitialLoad(stocks []models.Stock) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing data
	_, err = tx.Exec("DELETE FROM stock")
	if err != nil {
		return err
	}

	now := time.Now()
	batchSize := 1000
	total := len(stocks)

	for start := 0; start < total; start += batchSize {
		end := start + batchSize
		if end > total {
			end = total
		}
		batch := stocks[start:end]

		var (
			args         []interface{}
			placeholders []string
		)
		for i, s := range batch {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
			args = append(args, s.Ticker, s.Name, true, now) // active = true, updated_at
		}

		query := "INSERT INTO stock (ticker, name, active, updated_at) VALUES " + strings.Join(placeholders, ",")
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return r.LoadStockCache()
}

// Subsequent updates: review changes, auto-update same name different ticker, mark missing for manual review
func (r *StockRepository) SaveStocksWithReview(latestStocks []models.Stock) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	existing := make(map[string]models.Stock)
	rows, err := tx.Query("SELECT id, ticker, name FROM stock")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var s models.Stock
		if err := rows.Scan(&s.Id, &s.Ticker, &s.Name); err != nil {
			return err
		}
		existing[s.Ticker] = s
	}

	latestMap := make(map[string]models.Stock)
	for _, s := range latestStocks {
		latestMap[s.Ticker] = s
	}

	now := time.Now()

	// Process incoming stocks
	for _, latest := range latestStocks {
		existingStock, exists := existing[latest.Ticker]
		if exists {
			// Same ticker exists, update name if changed
			if existingStock.Name != latest.Name {
				_, err := tx.Exec("UPDATE stock SET name = $1, updated_at = $2 WHERE ticker = $3", latest.Name, now, latest.Ticker)
				if err != nil {
					return err
				}
			}
			continue
		}

		// If ticker is new, check if name matches existing (changed ticker scenario)
		autoMatched := false
		for _, oldStock := range existing {
			if oldStock.Name == latest.Name {
				// Update old ticker to new ticker
				_, err := tx.Exec("UPDATE stock SET ticker = $1, updated_at = $2 WHERE id = $3", latest.Ticker, now, oldStock.Id)
				if err != nil {
					return err
				}
				autoMatched = true
				break
			}
		}
		if autoMatched {
			continue
		}

		// New ticker + name not matching existing - add to manual review
		_, err := tx.Exec("INSERT INTO pending_stock_review (ticker, name, reason) VALUES ($1, $2, $3) ON CONFLICT (ticker, name, reason) DO NOTHING", latest.Ticker, latest.Name, "new_ticker_and_name")
		if err != nil {
			return err
		}
	}

	// Identify existing stocks missing from latest data - mark for review for deactivation/removal
	for oldTicker, oldStock := range existing {
		if _, found := latestMap[oldTicker]; !found {
			_, err := tx.Exec("INSERT INTO pending_stock_review (ticker, name, reason) VALUES ($1, $2, $3) ON CONFLICT (ticker, name, reason) DO NOTHING", oldTicker, oldStock.Name, "ticker_missing")
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return r.LoadStockCache()
}
