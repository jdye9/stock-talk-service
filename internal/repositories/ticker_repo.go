package repositories

import (
	"database/sql"
	"stock-talk-service/internal/models"
	"sync"
)

type TickerRepository struct {
	db             *sql.DB
	nasdaqCache    map[string]models.Ticker // symbol -> Ticker
	otherCache     map[string]models.Ticker // symbol -> Ticker
	cacheMutex     sync.RWMutex
}

func NewTickerRepository(db *sql.DB) *TickerRepository {
	return &TickerRepository{
		db:          db,
		nasdaqCache: make(map[string]models.Ticker),
		otherCache:  make(map[string]models.Ticker),
	}
}

// LoadNasdaqCache loads all Nasdaq tickers from DB into the in-memory cache.
func (r *TickerRepository) LoadNasdaqCache() error {
	rows, err := r.db.Query("SELECT symbol, name FROM nasdaq")
	if err != nil {
		return err
	}
	defer rows.Close()

	cache := make(map[string]models.Ticker)
	for rows.Next() {
		var t models.Ticker
		if err := rows.Scan(&t.Symbol, &t.Name); err != nil {
			return err
		}
		cache[t.Symbol] = t
	}

	r.cacheMutex.Lock()
	r.nasdaqCache = cache
	r.cacheMutex.Unlock()
	return rows.Err()
}

// LoadOtherCache loads all Other tickers from DB into the in-memory cache.
func (r *TickerRepository) LoadOtherCache() error {
	rows, err := r.db.Query("SELECT symbol, name FROM other")
	if err != nil {
		return err
	}
	defer rows.Close()

	cache := make(map[string]models.Ticker)
	for rows.Next() {
		var t models.Ticker
		if err := rows.Scan(&t.Symbol, &t.Name); err != nil {
			return err
		}
		cache[t.Symbol] = t
	}

	r.cacheMutex.Lock()
	r.otherCache = cache
	r.cacheMutex.Unlock()
	return rows.Err()
}

// GetNasdaqTickerBySymbol returns a Nasdaq ticker from the in-memory cache.
func (r *TickerRepository) GetNasdaqTickerBySymbol(symbol string) (models.Ticker, bool) {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	t, ok := r.nasdaqCache[symbol]
	return t, ok
}

// GetOtherTickerBySymbol returns an Other ticker from the in-memory cache.
func (r *TickerRepository) GetOtherTickerBySymbol(symbol string) (models.Ticker, bool) {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	t, ok := r.otherCache[symbol]
	return t, ok
}

// GetNasdaqTickers returns all Nasdaq tickers from the in-memory cache.
func (r *TickerRepository) GetNasdaqTickers() []models.Ticker {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	tickers := make([]models.Ticker, 0, len(r.nasdaqCache))
	for _, t := range r.nasdaqCache {
		tickers = append(tickers, t)
	}
	return tickers
}

// GetOtherTickers returns all Other tickers from the in-memory cache.
func (r *TickerRepository) GetOtherTickers() []models.Ticker {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	tickers := make([]models.Ticker, 0, len(r.otherCache))
	for _, t := range r.otherCache {
		tickers = append(tickers, t)
	}
	return tickers
}

// SaveNasdaqTickers replaces all Nasdaq tickers in DB and refreshes the cache.
func (r *TickerRepository) SaveNasdaqTickers(tickers []models.Ticker) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM nasdaq")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO nasdaq (symbol, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tickers {
		if _, err := stmt.Exec(t.Symbol, t.Name); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return r.LoadNasdaqCache()
}

// SaveOtherTickers replaces all Other tickers in DB and refreshes the cache.
func (r *TickerRepository) SaveOtherTickers(tickers []models.Ticker) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM other")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO other (symbol, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tickers {
		if _, err := stmt.Exec(t.Symbol, t.Name); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return r.LoadOtherCache()
}
