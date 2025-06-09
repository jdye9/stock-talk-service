package services

import (
	"stock-talk-service/internal/repositories"
)

// FTPService handles operations related to FTP
type FTPService struct {
	ftpRepo *repositories.FTPRepository
}

// NewFTPService initializes a new FTPService with an existing FTP repository
func NewFTPService(ftpRepo *repositories.FTPRepository) *FTPService {
	return &FTPService{ftpRepo: ftpRepo }
}

func (ftpService *FTPService) ProcessUpdateTickers() {
	ftpService.ftpRepo.FetchAndUpdateTickers()
}