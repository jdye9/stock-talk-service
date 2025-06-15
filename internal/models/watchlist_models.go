package models

import "time"

type Watchlist struct {
	Id     string  `json:"id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateWatchlistRequest struct {
	Name string `json:"name"`
	Stocks []string `json:"stocks"`
	Crypto []string `json:"crypto"`
}