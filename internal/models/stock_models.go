package models

type Stock struct {
	Id string `json:"id"`
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
}
