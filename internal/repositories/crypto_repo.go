package repositories

import (
	"database/sql"
	"fmt"
	"stock-talk-service/internal/models"
	"strings"
	"sync"
)

type CryptoRepository struct {
    db         *sql.DB
    cache      map[string]models.Crypto // id -> Crypto
    cacheMutex sync.RWMutex
}

func NewCryptoRepository(db *sql.DB) *CryptoRepository {
    return &CryptoRepository{
        db:    db,
        cache: make(map[string]models.Crypto),
    }
}

// LoadCryptoCache loads all crypto from DB into the in-memory cache.
func (r *CryptoRepository) LoadCryptoCache() error {
    rows, err := r.db.Query("SELECT id, coingecko_id, ticker, name FROM crypto")
    if err != nil {
        return err
    }
    defer rows.Close()

    cache := make(map[string]models.Crypto)
    for rows.Next() {
        var c models.Crypto
        if err := rows.Scan(&c.Id, &c.CoingeckoId, &c.Ticker, &c.Name); err != nil {
            return err
        }
        cache[c.Id] = c
    }

    r.cacheMutex.Lock()
    r.cache = cache
    r.cacheMutex.Unlock()
    return rows.Err()
}

// GetCryptoByID returns a Crypto from the in-memory cache.
func (r *CryptoRepository) GetCryptoByID(id string) (models.Crypto, bool) {
    r.cacheMutex.RLock()
    defer r.cacheMutex.RUnlock()
    c, ok := r.cache[id]
    return c, ok
}

// GetAllCrypto returns all crypto from the in-memory cache.
func (r *CryptoRepository) GetAllCrypto() []models.Crypto {
    r.cacheMutex.RLock()
    defer r.cacheMutex.RUnlock()
    crypto := make([]models.Crypto, 0, len(r.cache))
    for _, c := range r.cache {
        crypto = append(crypto, c)
    }
    return crypto
}

// SaveCrypto replaces all crypto in DB and refreshes the cache.
func (r *CryptoRepository) SaveCrypto(crypto []models.Crypto) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.Exec("DELETE FROM crypto")
    if err != nil {
        return err
    }

    batchSize := 1000
	total := len(crypto)

	for start := 0; start < total; start += batchSize {
		end := start + batchSize
		if end > total {
			end = total
		}
		batch := crypto[start:end]

		var (
			args        []interface{}
			placeholders []string
		)
		for i, s := range batch {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
			args = append(args, s.Ticker, s.CoingeckoId, s.Name,)
		}

		query := "INSERT INTO crypto (ticker, coingecko_id, name) VALUES " + strings.Join(placeholders, ",")
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

    if err := tx.Commit(); err != nil {
        return err
    }

    // Refresh cache after DB update
    return r.LoadCryptoCache()
}