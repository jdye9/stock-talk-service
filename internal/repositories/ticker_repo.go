package repositories

import (
	"database/sql"
	"stock-talk-service/internal/models"
)

type TickerRepository struct {
	db *sql.DB
}

func NewTickerRepository(db *sql.DB) *TickerRepository {
	return &TickerRepository{db: db}
}

func (r *TickerRepository) SaveNasdaqTickers(tickers []models.Ticker) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM nasdaq_tickers")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO nasdaq_tickers (symbol, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tickers {
		if _, err := stmt.Exec(t.Symbol, t.Name); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TickerRepository) GetNasdaqTickers() ([]models.Ticker, error) {
	rows, err := r.db.Query("SELECT symbol, name FROM nasdaq_tickers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickers []models.Ticker
	for rows.Next() {
		var t models.Ticker
		if err := rows.Scan(&t.Symbol, &t.Name); err != nil {
			return nil, err
		}
		tickers = append(tickers, t)
	}
	return tickers, rows.Err()
}

func (r *TickerRepository) SaveOtherTickers(tickers []models.Ticker) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM other_tickers")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO other_tickers (symbol, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tickers {
		if _, err := stmt.Exec(t.Symbol, t.Name); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TickerRepository) GetOtherTickers() ([]models.Ticker, error) {
	rows, err := r.db.Query("SELECT symbol, name FROM other_tickers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickers []models.Ticker
	for rows.Next() {
		var t models.Ticker
		if err := rows.Scan(&t.Symbol, &t.Name); err != nil {
			return nil, err
		}
		tickers = append(tickers, t)
	}
	return tickers, rows.Err()
}
