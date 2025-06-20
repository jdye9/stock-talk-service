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
	cache      map[string]models.Stock // id -> Stock
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

func (r *StockRepository) GetStockById(id string) (models.Stock, bool) {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	s, ok := r.cache[id]
	return s, ok
}

// Initial load: Delete all and reinsert (safe in dev or once)
func (r *StockRepository) SaveStocksInitialLoad(stocks []models.Stock) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
			args = append(args, s.Ticker, s.Name, true, now)
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

// SaveStocksWithReview applies manual review logic according to your spec
func (r *StockRepository) SaveStocksWithReview(latestStocks []models.Stock) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM pending_stock_review")
	if err != nil {
		return err
	}

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
	tickers := make([]interface{}, 0, len(latestStocks))
	for _, s := range latestStocks {
		latestMap[s.Ticker] = s
		tickers = append(tickers, s.Ticker)
	}

	now := time.Now()

	// STEP 1: Resolve reappeared tickers previously marked as missing
	if len(tickers) > 0 {
		placeholders := make([]string, len(tickers))
		for i := range tickers {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query := fmt.Sprintf(`
			UPDATE pending_stock_review
			SET resolved = TRUE, resolved_at = $%d
			WHERE ticker IN (%s) AND reason = 'ticker_missing' AND resolved = FALSE
		`, len(tickers)+1, strings.Join(placeholders, ","))
		args := append(tickers, now)
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

	// STEP 2: Resolve 'ticker_new' for tickers no longer in latest
	if len(tickers) > 0 {
		placeholders := make([]string, len(tickers))
		for i := range tickers {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query := fmt.Sprintf(`
			UPDATE pending_stock_review
			SET resolved = TRUE, resolved_at = $%d
			WHERE ticker NOT IN (%s) AND reason = 'ticker_new' AND resolved = FALSE
		`, len(tickers)+1, strings.Join(placeholders, ","))
		args := append(tickers, now)
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

	// STEP 3: Auto-resolve name_changed if name reverted
	rows, err = tx.Query(`
		SELECT id, ticker, name FROM pending_stock_review
		WHERE reason = 'name_changed' AND resolved = FALSE
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type pendingReview struct {
		ID     string
		Ticker string
		Name   string
	}
	var pending []pendingReview
	for rows.Next() {
		var r pendingReview
		if err := rows.Scan(&r.ID, &r.Ticker, &r.Name); err != nil {
			return err
		}
		pending = append(pending, r)
	}

	for _, r := range pending {
		original, ok := existing[r.Ticker]
		if !ok {
			continue
		}
		latest, ok := latestMap[r.Ticker]
		if !ok {
			continue
		}
		if latest.Name == original.Name {
			if _, err := tx.Exec(`
				UPDATE pending_stock_review
				SET resolved = TRUE, resolved_at = $1
				WHERE id = $2
			`, now, r.ID); err != nil {
				return err
			}
		}
	}

	// STEP 4: Insert new review items for discrepancies
	for _, latest := range latestStocks {
		existingStock, exists := existing[latest.Ticker]

		if exists {
			if latest.Name != existingStock.Name {
				var count int
				err := tx.QueryRow(`
					SELECT COUNT(*) FROM pending_stock_review
					WHERE ticker = $1 AND name = $2 AND reason = 'name_changed' AND resolved = FALSE
				`, latest.Ticker, latest.Name).Scan(&count)
				if err != nil {
					return err
				}
				if count == 0 {
					_, err := tx.Exec(`
						INSERT INTO pending_stock_review (ticker, name, reason, resolved, created_at)
						VALUES ($1, $2, 'name_changed', FALSE, $3)
					`, latest.Ticker, latest.Name, now)
					if err != nil {
						return err
					}
				}
			}
			continue
		}

		// ticker is new → 'ticker_new'
		var count int
		err := tx.QueryRow(`
			SELECT COUNT(*) FROM pending_stock_review
			WHERE ticker = $1 AND name = $2 AND reason = 'ticker_new' AND resolved = FALSE
		`, latest.Ticker, latest.Name).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := tx.Exec(`
				INSERT INTO pending_stock_review (ticker, name, reason, resolved, created_at)
				VALUES ($1, $2, 'ticker_new', FALSE, $3)
			`, latest.Ticker, latest.Name, now)
			if err != nil {
				return err
			}
		}
	}

	// STEP 5: Flag missing tickers
	for oldTicker, oldStock := range existing {
		if _, found := latestMap[oldTicker]; !found {
			var count int
			err := tx.QueryRow(`
				SELECT COUNT(*) FROM pending_stock_review
				WHERE ticker = $1 AND name = $2 AND reason = 'ticker_missing' AND resolved = FALSE
			`, oldTicker, oldStock.Name).Scan(&count)
			if err != nil {
				return err
			}
			if count == 0 {
				_, err := tx.Exec(`
					INSERT INTO pending_stock_review (ticker, name, reason, resolved, created_at)
					VALUES ($1, $2, 'ticker_missing', FALSE, $3)
				`, oldTicker, oldStock.Name, now)
				if err != nil {
					return err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return r.LoadStockCache()
}

