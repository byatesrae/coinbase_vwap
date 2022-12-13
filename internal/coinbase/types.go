package coinbase

import (
	"fmt"
	"strconv"
)

type ProductID string

const (
	ProductIDUnknown ProductID = ""
	ProductIDBtcUsd  ProductID = "BTC-USD"
	ProductIDEthUsd  ProductID = "ETH-USD"
	ProductIDEthBtc  ProductID = "ETH-BTC"
)

type channelName string

const (
	channelNameMatches channelName = "matches"
)

type subscribeRequest struct {
	Type       string                    `json:"type"`
	ProductIDs []ProductID               `json:"product_ids"`
	Channels   []subscribeChannelRequest `json:"channels"`
}

type subscribeChannelRequest struct {
	Name       channelName `json:"name"`
	ProductIDs []ProductID `json:"product_ids"`
}

type Match struct {
	Type      string    `json:"type"`
	ProductID ProductID `json:"product_id"`
	Size      string    `json:"size"`
	Price     string    `json:"price"`
	Message   string    `json:"message"`
}

type MatchResponse struct {
	Match Match
	Err   error
}

func (m *MatchResponse) ToUnitsAndUnitPrice() (float64, float64, error) {
	if m.Err != nil {
		return 0, 0, fmt.Errorf("match response: %w", m.Err)
	}

	units, err := strconv.ParseFloat(m.Match.Size, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse units from size: %w", err)
	}

	unitPrice, err := strconv.ParseFloat(m.Match.Price, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse unitPrice from price: %w", err)
	}

	return units, unitPrice, err
}
