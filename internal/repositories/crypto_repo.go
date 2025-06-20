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
	cache      map[string]models.Crypto // id -> Crypto
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
		cache[c.Id] = c
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

func (r *CryptoRepository) GetCryptoByID(id string) (models.Crypto, bool) {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	c, ok := r.cache[id]
	return c, ok
}

func (r *CryptoRepository) SaveCryptoInitialLoad(cryptos []models.Crypto) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
			args = append(args, c.Uid, c.CoingeckoId, c.Ticker, c.Name, true, now)
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

func (r *CryptoRepository) SaveCryptoWithReview(latestCryptos []models.Crypto) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM pending_crypto_review")
	if err != nil {
		return err
	}

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

	// Prepare UIDs list for queries
	uids := make([]interface{}, 0, len(latestCryptos))
	for _, c := range latestCryptos {
		uids = append(uids, c.Uid)
	}

	// STEP 1: Resolve previously missing UIDs that now reappear
	if len(uids) > 0 {
		placeholders := make([]string, len(uids))
		for i := range uids {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query := fmt.Sprintf(`
			UPDATE pending_crypto_review
			SET resolved = TRUE, resolved_at = $%d
			WHERE uid IN (%s) AND reason = 'uid_missing' AND resolved = FALSE
		`, len(uids)+1, strings.Join(placeholders, ","))
		args := append(uids, now)
		if _, err = tx.Exec(query, args...); err != nil {
			return err
		}
	}

	// STEP 2: Resolve 'uid_new' if UID no longer in latest list
	if len(uids) > 0 {
		placeholders := make([]string, len(uids))
		for i := range uids {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		query := fmt.Sprintf(`
			UPDATE pending_crypto_review
			SET resolved = TRUE, resolved_at = $%d
			WHERE uid NOT IN (%s) AND reason = 'uid_new' AND resolved = FALSE
		`, len(uids)+1, strings.Join(placeholders, ","))
		args := append(uids, now)
		if _, err = tx.Exec(query, args...); err != nil {
			return err
		}
	}

	// STEP 3: Auto-resolve reverted name/ticker discrepancies
	rows, err = tx.Query(`
		SELECT id, uid, reason FROM pending_crypto_review
		WHERE resolved = FALSE AND reason IN ('name_changed', 'ticker_changed', 'name_ticker_changed')
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type pendingReview struct {
		ID     string
		UID    string
		Reason string
	}
	var pending []pendingReview
	for rows.Next() {
		var r pendingReview
		if err := rows.Scan(&r.ID, &r.UID, &r.Reason); err != nil {
			return err
		}
		pending = append(pending, r)
	}

	for _, r := range pending {
		var originalName, originalTicker string
		err := tx.QueryRow(`SELECT name, ticker FROM crypto WHERE uid = $1`, r.UID).Scan(&originalName, &originalTicker)
		if err != nil {
			// Crypto no longer exists — skip
			continue
		}

		latest, ok := latestMap[r.UID]
		if !ok {
			continue
		}

		resolved := false
		switch r.Reason {
		case "name_changed":
			if latest.Name == originalName {
				resolved = true
			}
		case "ticker_changed":
			if latest.Ticker == originalTicker {
				resolved = true
			}
		case "name_ticker_changed":
			if latest.Name == originalName || latest.Ticker == originalTicker {
				resolved = true
			}
		}

		if resolved {
			if _, err := tx.Exec(`UPDATE pending_crypto_review SET resolved = TRUE, resolved_at = $1 WHERE id = $2`, now, r.ID); err != nil {
				return err
			}
		}
	}

	// STEP 4: Insert/update pending_crypto_review with new discrepancies
	for _, latest := range latestCryptos {
		existingCrypto, exists := existing[latest.Uid]
		if exists {
			nameChanged := existingCrypto.Name != latest.Name
			tickerChanged := existingCrypto.Ticker != latest.Ticker

			if nameChanged && tickerChanged {
				// Check for existing unresolved combined reason
				var count int
				err := tx.QueryRow(`
					SELECT COUNT(*) FROM pending_crypto_review
					WHERE uid = $1 AND reason = 'name_ticker_changed' AND resolved = FALSE
				`, latest.Uid).Scan(&count)
				if err != nil {
					return err
				}
				if count == 0 {
					_, err := tx.Exec(`
						INSERT INTO pending_crypto_review (uid, coingecko_id, ticker, name, reason, resolved, created_at)
						VALUES ($1, $2, $3, $4, 'name_ticker_changed', FALSE, $5)
					`, latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name, now)
					if err != nil {
						return err
					}
				}

				// Mark individual name_changed and ticker_changed as resolved to avoid duplicates
				_, err = tx.Exec(`
					UPDATE pending_crypto_review
					SET resolved = TRUE, resolved_at = $1
					WHERE uid = $2 AND reason IN ('name_changed', 'ticker_changed') AND resolved = FALSE
				`, now, latest.Uid)
				if err != nil {
					return err
				}

			} else if nameChanged {
				var count int
				err := tx.QueryRow(`
					SELECT COUNT(*) FROM pending_crypto_review
					WHERE uid = $1 AND name = $2 AND reason = 'name_changed' AND resolved = FALSE
				`, latest.Uid, latest.Name).Scan(&count)
				if err != nil {
					return err
				}
				if count == 0 {
					_, err := tx.Exec(`
						INSERT INTO pending_crypto_review (uid, coingecko_id, ticker, name, reason, resolved, created_at)
						VALUES ($1, $2, $3, $4, 'name_changed', FALSE, $5)
					`, latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name, now)
					if err != nil {
						return err
					}
				}

			} else if tickerChanged {
				var count int
				err := tx.QueryRow(`
					SELECT COUNT(*) FROM pending_crypto_review
					WHERE uid = $1 AND ticker = $2 AND reason = 'ticker_changed' AND resolved = FALSE
				`, latest.Uid, latest.Ticker).Scan(&count)
				if err != nil {
					return err
				}
				if count == 0 {
					_, err := tx.Exec(`
						INSERT INTO pending_crypto_review (uid, coingecko_id, ticker, name, reason, resolved, created_at)
						VALUES ($1, $2, $3, $4, 'ticker_changed', FALSE, $5)
					`, latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name, now)
					if err != nil {
						return err
					}
				}
			}
			continue
		}

		// UID not found → uid_new
		var count int
		err := tx.QueryRow(`
			SELECT COUNT(*) FROM pending_crypto_review
			WHERE uid = $1 AND coingecko_id = $2 AND ticker = $3 AND name = $4 AND reason = 'uid_new' AND resolved = FALSE
		`, latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := tx.Exec(`
				INSERT INTO pending_crypto_review (uid, coingecko_id, ticker, name, reason, resolved, created_at)
				VALUES ($1, $2, $3, $4, 'uid_new', FALSE, $5)
			`, latest.Uid, latest.CoingeckoId, latest.Ticker, latest.Name, now)
			if err != nil {
				return err
			}
		}
	}

	// STEP 5: Mark missing UIDs for review
	for oldUID, oldCrypto := range existing {
		if _, found := latestMap[oldUID]; !found {
			var count int
			err := tx.QueryRow(`
				SELECT COUNT(*) FROM pending_crypto_review
				WHERE uid = $1 AND coingecko_id = $2 AND ticker = $3 AND name = $4 AND reason = 'uid_missing' AND resolved = FALSE
			`, oldCrypto.Uid, oldCrypto.CoingeckoId, oldCrypto.Ticker, oldCrypto.Name).Scan(&count)
			if err != nil {
				return err
			}
			if count == 0 {
				_, err := tx.Exec(`
					INSERT INTO pending_crypto_review (uid, coingecko_id, ticker, name, reason, resolved, created_at)
					VALUES ($1, $2, $3, $4, 'uid_missing', FALSE, $5)
				`, oldCrypto.Uid, oldCrypto.CoingeckoId, oldCrypto.Ticker, oldCrypto.Name, now)
				if err != nil {
					return err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return r.LoadCryptoCache()
}

