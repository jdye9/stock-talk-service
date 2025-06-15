package repositories

import (
	"database/sql"
	"stock-talk-service/internal/models"
)

type WatchlistRepository struct {
	db *sql.DB
}

func NewWatchlistRepository(db *sql.DB) *WatchlistRepository {
	return &WatchlistRepository{db: db}
}

// Get a watchlist by ID
func (r *WatchlistRepository) GetWatchlistByID(id int) (*models.Watchlist, error) {
	var w models.Watchlist
	err := r.db.QueryRow("SELECT id, name, created_at FROM watchlist WHERE id = $1", id).Scan(&w.Id, &w.Name, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// Get all watchlists
func (r *WatchlistRepository) GetAllWatchlists() ([]models.Watchlist, error) {
	rows, err := r.db.Query("SELECT id, name, created_at FROM watchlist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchlists []models.Watchlist
	for rows.Next() {
		var w models.Watchlist
		if err := rows.Scan(&w.Id, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		watchlists = append(watchlists, w)
	}
	return watchlists, nil
}

// Create a new watchlist (transactional)
func (r *WatchlistRepository) CreateWatchlist(w *models.Watchlist, swi *[]string, cwi *[]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		"INSERT INTO watchlist (name) VALUES ($1) RETURNING id, name, created_at",
		w.Name,
	).Scan(&w.Id, &w.Name, &w.CreatedAt)
	if err != nil {
		return err
	}

	// Insert each watchlist item (stock), linking to the new watchlist ID
	for _, stockId := range *swi {
		if err := r.AddStockToWatchlist(tx, stockId, w.Id); err != nil {
			return err
		}
	}

	// Insert each watchlist item (crypto), linking to the new watchlist ID
	for _, cryptoId := range *cwi {
		if err := r.AddCryptoToWatchlist(tx, cryptoId, w.Id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddStockToWatchlist inserts a watchlist item using the provided transaction
func (r *WatchlistRepository) AddStockToWatchlist(tx *sql.Tx, stockId string, watchlistId string) error {
    _, err := tx.Exec(
        "INSERT INTO watchlist_stock (watchlist_id, stock_id) VALUES ($1, $2)",
        watchlistId, stockId,
    )
    return err
}

// AddStockToWatchlist inserts a watchlist item using the provided transaction
func (r *WatchlistRepository) AddCryptoToWatchlist(tx *sql.Tx, cryptoId string, watchlistId string) error {
    _, err := tx.Exec(
        "INSERT INTO watchlist_crypto (watchlist_id, crypto_id) VALUES ($1, $2)",
        watchlistId, cryptoId,
    )
    return err
}

// Update a watchlist (transactional)
func (r *WatchlistRepository) UpdateWatchlist(w *models.Watchlist) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE watchlist SET name = $1 WHERE id = $2",
		w.Name, w.Id,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete a watchlist (transactional)
func (r *WatchlistRepository) DeleteWatchlist(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM watchlist WHERE id = $1", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

