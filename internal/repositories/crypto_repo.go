package repositories

import (
	"database/sql"
	"fmt"
	"stock-talk-service/internal/models"
	"strings"
	"sync"
	"time"
)

type CryptoRepository struct {
	db         *sql.DB
	cache      map[string]models.Crypto // coingecko_id -> Crypto
	cacheMutex sync.RWMutex
}

func NewCryptoRepository(db *sql.DB) *CryptoRepository {
	return &CryptoRepository{
		db:    db,
		cache: make(map[string]models.Crypto),
	}
}

func (r *CryptoRepository) LoadCryptoCache() error {
	rows, err := r.db.Query("SELECT id, uid, coingecko_id, ticker, name FROM crypto")
	if err != nil {
		return err
	}
	defer rows.Close()

	cache := make(map[string]models.Crypto)
	for rows.Next() {
		var c models.Crypto
		if err := rows.Scan(&c.Id, &c.Uid, &c.CoingeckoId, &c.Ticker, &c.Name); err != nil {
			return err
		}
		cache[c.CoingeckoId] = c
	}

	r.cacheMutex.Lock()
	r.cache = cache
	r.cacheMutex.Unlock()
	return rows.Err()
}

func (r *CryptoRepository) GetAllCrypto() []models.Crypto {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	cryptos := make([]models.Crypto, 0, len(r.cache))
	for _, c := range r.cache {
		cryptos = append(cryptos, c)
	}
	return cryptos
}

func (r *CryptoRepository) GetCryptoByID(coingeckoId string) (models.Crypto, bool) {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	c, ok := r.cache[coingeckoId]
	return c, ok
}

// Initial load: Insert all cryptos, mark active, no manual review
func (r *CryptoRepository) SaveCryptoInitialLoad(cryptos []models.Crypto) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing data
	_, err = tx.Exec("DELETE FROM crypto")
	if err != nil {
		return err
	}

	now := time.Now()
	batchSize := 1000
	total := len(cryptos)

	for start := 0; start < total; start += batchSize {
		end := start + batchSize
		if end > total {
			end = total
		}
		batch := cryptos[start:end]

		var (
			args         []interface{}
			placeholders []string
		)
		for i, c := range batch {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
			args = append(args, c.Uid, c.CoingeckoId, c.Ticker, c.Name, true, now) // active = true, updated_at
		}

		query := "INSERT INTO crypto (uid, coingecko_id, ticker, name, active, updated_at) VALUES " + strings.Join(placeholders, ",")
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return r.LoadCryptoCache()
}

// Subsequent updates: review changes, auto-update same name different ID, mark missing for manual review
func (r *CryptoRepository) SaveCryptoWithReview(latestCryptos []models.Crypto) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	existing := make(map[string]models.Crypto)
	rows, err := tx.Query("SELECT id, uid, coingecko_id, ticker, name FROM crypto")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Crypto
		if err := rows.Scan(&c.Id, &c.Uid, &c.CoingeckoId, &c.Ticker, &c.Name); err != nil {
			return err
		}
		existing[c.Uid] = c
	}

	latestMap := make(map[string]models.Crypto)
	for _, c := range latestCryptos {
		latestMap[c.Uid] = c
	}

	now := time.Now()
	// Process incoming cryptos
	for _, latest := range latestCryptos {
		existingCrypto, exists := existing[latest.Uid]
		if exists {
			// Same Uid exists, update fields if changed
			if existingCrypto.Name != latest.Name {
				_, err := tx.Exec("UPDATE crypto SET uid = $1, coingecko_id = $2, ticker = $3, name = $4, updated_at = $5 WHERE id = $6", latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name, now, existingCrypto.Id)
				if err != nil {
					return err
				}
			}
			continue
		}

		// If UID is new, check if Name or Ticker matches existing (changed UID scenario)
		autoMatched := false
		for _, oldCrypto := range existing {
			if oldCrypto.Name == latest.Name || oldCrypto.Ticker == latest.Ticker {
				_, err := tx.Exec("UPDATE crypto SET uid = $1, coingecko_id = $2, ticker = $3, name = $4, updated_at = $5 WHERE id = $6", latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name, now, oldCrypto.Id)
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

		// New UID + name and ticker not matching existing - add to manual review
		_, err := tx.Exec("INSERT INTO pending_crypto_review (uid, coingecko_id, ticker, name, reason) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (uid, coingecko_id, ticker, name, reason) DO NOTHING", latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name, "new_id_and_name")
		if err != nil {
			return err
		}
	}

	// Identify existing cryptos missing from latest data - mark for review for deactivation/removal
	for oldID, oldCrypto := range existing {
		if _, found := latestMap[oldID]; !found {
			_, err := tx.Exec("INSERT INTO pending_crypto_review (uid, coingecko_id, ticker, name, reason) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (uid, coingecko_id, ticker, name, reason) DO NOTHING", oldCrypto.Uid, oldCrypto.CoingeckoId, oldCrypto.Ticker, oldCrypto.Name, "id_missing")
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return r.LoadCryptoCache()
}
