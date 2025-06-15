package services

import (
	"stock-talk-service/internal/models"
	"stock-talk-service/internal/repositories"
)

type WatchlistService struct {
    watchlistRepo *repositories.WatchlistRepository
}

func NewWatchlistService(watchlistRepo *repositories.WatchlistRepository) *WatchlistService {
    return &WatchlistService{watchlistRepo: watchlistRepo}
}

// GetWatchlistByID returns a watchlist by its ID.
func (s *WatchlistService) GetWatchlistByID(id int) (*models.Watchlist, error) {
    return s.watchlistRepo.GetWatchlistByID(id)
}

// GetAllWatchlists returns all watchlists.
func (s *WatchlistService) GetAllWatchlists() ([]models.Watchlist, error) {
    return s.watchlistRepo.GetAllWatchlists()
}

// CreateWatchlist creates a new watchlist.
func (s *WatchlistService) CreateWatchlist(w *models.Watchlist, swi *[]string, cwi *[]string) error {
    return s.watchlistRepo.CreateWatchlist(w, swi, cwi)
}

// UpdateWatchlist updates an existing watchlist.
func (s *WatchlistService) UpdateWatchlist(w *models.Watchlist) error {
    return s.watchlistRepo.UpdateWatchlist(w)
}

// DeleteWatchlist deletes a watchlist by ID.
func (s *WatchlistService) DeleteWatchlist(id int64) error {
    return s.watchlistRepo.DeleteWatchlist(id)
}