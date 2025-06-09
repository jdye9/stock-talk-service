package models

import "sync"

type Ticker struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type TickerData struct {
	Nasdaq []Ticker
	Other  []Ticker
	sync.RWMutex
}
