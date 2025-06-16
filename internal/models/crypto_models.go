package models

type Crypto struct {
	Id string `json:"id"`
	Uid string `json:"uid"`
	Ticker string `json:"ticker"`
	Name string `json:"name"`
	CoingeckoId string `json:"coingecko_id"`
}

type CryptoJSON struct {
	Uid	 string `json:"uid"`
    Ticker  string `json:"ticker"`
    Name    string `json:"name"`
    Aliases struct {
        CoinGeckoID string `json:"coingecko_id"`
    } `json:"aliases"`
}

// To convert to your Crypto struct:
func (cj CryptoJSON) ToCrypto() Crypto {
    return Crypto{
    	Ticker: cj.Ticker,
        Name:   cj.Name,
		CoingeckoId: cj.Aliases.CoinGeckoID,
		Uid: cj.Uid,

    }
}

type CryptoPriceRequest struct {
	CoinIDs      []string `json:"coin_ids" binding:"required"`
	VsCurrencies []string `json:"vs_currencies" binding:"required"`
}

type CryptoHistoryRequest struct {
	CoinIDs    []string `json:"coin_ids" binding:"required"`
	VsCurrency string   `json:"vs_currency" binding:"required"`
	Days       string   `json:"days" binding:"required"`
	Interval   string   `json:"interval"`
}

type CryptoHistoryOHLCRequest struct {
	CoinIDs    []string `json:"coin_ids"`
	VsCurrency string   `json:"vs_currency"`
	Days       string   `json:"days"`
	Interval   string   `json:"interval"`
}

type CryptoPriceResponse struct {
    Prices               map[string]interface{} `json:"prices"`
    InvalidCoinIDs       []string               `json:"invalid_coin_ids"`
    InvalidVsCurrencies  []string               `json:"invalid_vs_currencies"`
}

type CryptoHistoryData struct {
    CoinID       string        `json:"coin_id"`
    VsCurrency   string        `json:"vs_currency"`
    Days         string        `json:"days"`
    Prices       interface{}   `json:"prices"`
    MarketCaps   interface{}   `json:"market_caps"`
    TotalVolumes interface{}   `json:"total_volumes"`
    Error        string        `json:"error,omitempty"`
}

type CryptoHistoryResponse struct {
    Data                map[string]CryptoHistoryData `json:"data"`
    InvalidCoinIDs      []string               `json:"invalid_coin_ids"`
    InvalidVsCurrencies []string               `json:"invalid_vs_currencies"`
}

type CryptoHistoryOHLCData struct {
    CoinID     string      `json:"coin_id"`
    VsCurrency string      `json:"vs_currency"`
    Days       string      `json:"days"`
    OHLC       interface{} `json:"ohlc"`
    Error      string      `json:"error,omitempty"`
}

type CryptoHistoryOHLCResponse struct {
    Data                map[string]CryptoHistoryOHLCData `json:"data"`
    InvalidCoinIDs      []string            `json:"invalid_coin_ids"`
    InvalidVsCurrencies []string            `json:"invalid_vs_currencies"`
}