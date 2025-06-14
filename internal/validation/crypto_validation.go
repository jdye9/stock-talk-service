package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock-talk-service/internal/config"
	"strings"
	"sync"
)

var (
	coinIDCache      = make(map[string]bool)
	vsCurrencyCache  = make(map[string]bool)
	coinIDOnce       sync.Once
	vsCurrencyOnce   sync.Once
	coinIDMutex      sync.Mutex
	vsCurrencyMutex  sync.Mutex
)

type ValidationResult struct {
	ValidCoinIDs        []string
	ValidVsCurrencies   []string
	InvalidCoinIDs      []string
	InvalidVsCurrencies []string
}

// Fetch valid coin IDs once and cache
func getValidCoinIDs(cfg *config.Config) (map[string]bool, error) {
	var err error
	coinIDOnce.Do(func() {
		resp, e := http.Get(cfg.CoingeckoBaseUrl + "/coins/list")
		if e != nil {
			err = e
			return
		}
		defer resp.Body.Close()

		var coins []struct {
			ID string `json:"id"`
		}
		if e := json.NewDecoder(resp.Body).Decode(&coins); e != nil {
			err = e
			return
		}

		coinIDMutex.Lock()
		defer coinIDMutex.Unlock()
		for _, coin := range coins {
			coinIDCache[coin.ID] = true
		}
	})
	return coinIDCache, err
}

// Fetch valid vs_currencies once and cache
func getValidVsCurrencies(cfg *config.Config) (map[string]bool, error) {
	var err error
	vsCurrencyOnce.Do(func() {
		resp, e := http.Get(cfg.CoingeckoBaseUrl + "/simple/supported_vs_currencies")
		if e != nil {
			err = e
			return
		}
		defer resp.Body.Close()

		var currencies []string
		if e := json.NewDecoder(resp.Body).Decode(&currencies); e != nil {
			err = e
			return
		}

		vsCurrencyMutex.Lock()
		defer vsCurrencyMutex.Unlock()
		for _, cur := range currencies {
			vsCurrencyCache[strings.ToLower(cur)] = true
		}
	})
	return vsCurrencyCache, err
}

// Validate crypto input values
func ValidateCryptoInputs(cfg *config.Config, coinIDs, vsCurrencies []string) (ValidationResult, error) {
	coinCache, err := getValidCoinIDs(cfg)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("error fetching coin IDs: %w", err)
	}
	vsCache, err := getValidVsCurrencies(cfg)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("error fetching vs currencies: %w", err)
	}

	var (
		validCoinIDs      []string
		invalidCoinIDs    []string
		validVsCurrencies []string
		invalidVsCurrencies []string
	)

	for _, cid := range coinIDs {
		if coinCache[cid] {
			validCoinIDs = append(validCoinIDs, cid)
		} else {
			invalidCoinIDs = append(invalidCoinIDs, cid)
		}
	}

	for _, cur := range vsCurrencies {
		lc := strings.ToLower(cur)
		if vsCache[lc] {
			validVsCurrencies = append(validVsCurrencies, lc)
		} else {
			invalidVsCurrencies = append(invalidVsCurrencies, cur)
		}
	}

	return ValidationResult{
		ValidCoinIDs:        validCoinIDs,
		ValidVsCurrencies:   validVsCurrencies,
		InvalidCoinIDs:      invalidCoinIDs,
		InvalidVsCurrencies: invalidVsCurrencies,
	}, nil
}

// Validate and raise errors like FastAPI's HTTPException
func ValidateAndRaise(cfg *config.Config, coinIDs, vsCurrencies []string) (ValidationResult, error) {
	result, err := ValidateCryptoInputs(cfg, coinIDs, vsCurrencies)
	if err != nil {
		return result, err
	}

	if len(result.ValidCoinIDs) == 0 && len(result.ValidVsCurrencies) == 0 {
		return result, fmt.Errorf("no valid coin_id(s) or vs_currency(ies) provided. Invalid coin_ids: %v, invalid vs_currencies: %v", result.InvalidCoinIDs, result.InvalidVsCurrencies)
	}
	if len(result.ValidCoinIDs) == 0 {
		return result, fmt.Errorf("no valid coin_id(s) provided. Invalid: %v", result.InvalidCoinIDs)
	}
	if len(result.ValidVsCurrencies) == 0 {
		return result, fmt.Errorf("no valid vs_currency(ies) provided. Invalid: %v", result.InvalidVsCurrencies)
	}

	return result, nil
}
