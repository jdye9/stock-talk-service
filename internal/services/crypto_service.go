package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"stock-talk-service/internal/config"
	"stock-talk-service/internal/models"
	"stock-talk-service/internal/repositories"
	"stock-talk-service/internal/validation"
	"strings"
)

type CryptoService struct {
	cryptoRepo *repositories.CryptoRepository
	cfg        *config.Config
}

func NewCryptoService(cryptoRepo *repositories.CryptoRepository, cfg *config.Config) *CryptoService {
	return &CryptoService{
		cryptoRepo: cryptoRepo,
		cfg:        cfg,
	}
}

// GetAllCrypto returns all crypto from the cache
func (s *CryptoService) GetAllCrypto() []models.Crypto {
	return s.cryptoRepo.GetAllCrypto()
}

// GetCryptoByID returns a crypto by coingeckoId from cache
func (s *CryptoService) GetCryptoByID(id string) (models.Crypto, bool) {
	return s.cryptoRepo.GetCryptoByID(id)
}

// SaveCryptoInitialLoad saves all cryptos (initial load)
func (s *CryptoService) SaveCryptoInitialLoad(cryptos []models.Crypto) error {
	return s.cryptoRepo.SaveCryptoInitialLoad(cryptos)
}

// SaveCryptoWithReview saves cryptos with review logic for updates
func (s *CryptoService) SaveCryptoWithReview(cryptos []models.Crypto) error {
	return s.cryptoRepo.SaveCryptoWithReview(cryptos)
}

// ReloadCryptoCache reloads cache from DB
func (s *CryptoService) ReloadCryptoCache() error {
	return s.cryptoRepo.LoadCryptoCache()
}

func (s *CryptoService) GetCryptoPrice(coinIDs, vsCurrencies []string) (*models.CryptoPriceResponse, error) {
	result, err := validation.ValidateAndRaise(s.cfg, coinIDs, vsCurrencies)
	if err != nil {
		return &models.CryptoPriceResponse{
			Prices:              map[string]interface{}{},
			InvalidCoinIDs:      result.InvalidCoinIDs,
			InvalidVsCurrencies: result.InvalidVsCurrencies,
		}, nil
	}

	url := fmt.Sprintf("%s/simple/price", s.cfg.CoingeckoBaseUrl)
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("ids", strings.Join(result.ValidCoinIDs, ","))
	q.Add("vs_currencies", strings.Join(result.ValidVsCurrencies, ","))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	missingIDs := []string{}
	for _, cid := range result.ValidCoinIDs {
		if _, ok := data[cid]; !ok {
			missingIDs = append(missingIDs, cid)
		}
	}

	return &models.CryptoPriceResponse{
		Prices:              data,
		InvalidCoinIDs:      append(result.InvalidCoinIDs, missingIDs...),
		InvalidVsCurrencies: result.InvalidVsCurrencies,
	}, nil
}

func (s *CryptoService) GetCryptoHistory(coinIDs []string, vsCurrency, days, interval string) (*models.CryptoHistoryResponse, error) {
	result, err := validation.ValidateAndRaise(s.cfg, coinIDs, []string{vsCurrency})
	if err != nil {
		return &models.CryptoHistoryResponse{
			Data:                map[string]models.CryptoHistoryData{},
			InvalidCoinIDs:      result.InvalidCoinIDs,
			InvalidVsCurrencies: result.InvalidVsCurrencies,
		}, nil
	}

	historyData := make(map[string]models.CryptoHistoryData)
	for _, coinID := range result.ValidCoinIDs {
		url := fmt.Sprintf("%s/coins/%s/market_chart", s.cfg.CoingeckoBaseUrl, coinID)
		req, _ := http.NewRequest("GET", url, nil)
		q := req.URL.Query()
		q.Add("vs_currency", result.ValidVsCurrencies[0])
		q.Add("days", days)
		if interval != "" {
			q.Add("interval", interval)
		}
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			historyData[coinID] = models.CryptoHistoryData{
				CoinID: coinID, VsCurrency: vsCurrency, Days: days, Error: err.Error(),
			}
			continue
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			historyData[coinID] = models.CryptoHistoryData{
				CoinID: coinID, VsCurrency: vsCurrency, Days: days, Error: err.Error(),
			}
			continue
		}

		historyData[coinID] = models.CryptoHistoryData{
			CoinID:       coinID,
			VsCurrency:   vsCurrency,
			Days:         days,
			Prices:       data["prices"],
			MarketCaps:   data["market_caps"],
			TotalVolumes: data["total_volumes"],
		}
	}

	return &models.CryptoHistoryResponse{
		Data:                historyData,
		InvalidCoinIDs:      result.InvalidCoinIDs,
		InvalidVsCurrencies: result.InvalidVsCurrencies,
	}, nil
}

func (s *CryptoService) GetCryptoHistoryOHLC(coinIDs []string, vsCurrency, days, interval string) (*models.CryptoHistoryOHLCResponse, error) {
	result, err := validation.ValidateAndRaise(s.cfg, coinIDs, []string{vsCurrency})
	if err != nil {
		return &models.CryptoHistoryOHLCResponse{
			Data:                map[string]models.CryptoHistoryOHLCData{},
			InvalidCoinIDs:      result.InvalidCoinIDs,
			InvalidVsCurrencies: result.InvalidVsCurrencies,
		}, nil
	}

	ohlcData := make(map[string]models.CryptoHistoryOHLCData)
	for _, coinID := range result.ValidCoinIDs {
		url := fmt.Sprintf("%s/coins/%s/ohlc", s.cfg.CoingeckoBaseUrl, coinID)
		req, _ := http.NewRequest("GET", url, nil)
		q := req.URL.Query()
		q.Add("vs_currency", result.ValidVsCurrencies[0])
		q.Add("days", days)
		if interval != "" {
			q.Add("interval", interval)
		}
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ohlcData[coinID] = models.CryptoHistoryOHLCData{
				CoinID: coinID, VsCurrency: vsCurrency, Days: days, Error: err.Error(),
			}
			continue
		}
		defer resp.Body.Close()

		var ohlc interface{}
		if err := json.NewDecoder(resp.Body).Decode(&ohlc); err != nil {
			ohlcData[coinID] = models.CryptoHistoryOHLCData{
				CoinID: coinID, VsCurrency: vsCurrency, Days: days, Error: err.Error(),
			}
			continue
		}

		ohlcData[coinID] = models.CryptoHistoryOHLCData{
			CoinID:     coinID,
			VsCurrency: vsCurrency,
			Days:       days,
			OHLC:       ohlc,
		}
	}

	return &models.CryptoHistoryOHLCResponse{
		Data:                ohlcData,
		InvalidCoinIDs:      result.InvalidCoinIDs,
		InvalidVsCurrencies: result.InvalidVsCurrencies,
	}, nil
}

// FetchAllCrypto fetches cryptos from JSON file.
func (s *CryptoService) FetchAllCrypto() ([]models.Crypto, error) {
	file, err := os.Open("../../data/coin_mapping.json") // Adjust path as needed
	if err != nil {
		return nil, fmt.Errorf("error opening coin_mapping.json: %w", err)
	}
	defer file.Close()

	var wrapper struct {
		Data []models.CryptoJSON `json:"data"`
	}
	if err := json.NewDecoder(file).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	var crypto []models.Crypto
	for _, cj := range wrapper.Data {
		crypto = append(crypto, cj.ToCrypto())
	}

	log.Printf("Crypto fetched: %d", len(crypto))
	return crypto, nil
}

// Wrapper to fetch and save initial load
func (s *CryptoService) InitializeCrypto() error {
	cryptos, err := s.FetchAllCrypto()
	if err != nil {
		return err
	}
	return s.SaveCryptoInitialLoad(cryptos)
}

// Wrapper to fetch and save with review
func (s *CryptoService) FetchAndUpdateAllCrypto() error {
	cryptos, err := s.FetchAllCrypto()
	if err != nil {
		return err
	}
	return s.SaveCryptoWithReview(cryptos)
}


