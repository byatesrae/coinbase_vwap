package coinbase

import (
	"fmt"
	"strconv"
)

// ProductID is a Coinbase [Product ID].
//
// [Product ID]: https://docs.cloud.coinbase.com/exchange/docs/websocket-overview#specifying-product-ids
type ProductID string

const (
	ProductIDUnknown ProductID = ""
	ProductIDBtcUsd  ProductID = "BTC-USD"
	ProductIDEthUsd  ProductID = "ETH-USD"
	ProductIDEthBtc  ProductID = "ETH-BTC"
)

// MessageType is a Coinbase message type.
type MessageType string

const (
	MessageTypeUnknown       MessageType = ""
	MessageTypeError         MessageType = "error"
	MessageTypeLastMatch     MessageType = "last_match"
	MessageTypeMatch         MessageType = "match"
	MessageTypeSubscriptions MessageType = "subscriptions"
)

// channelName is a Coinbase [Channel name].
//
// [Channel name]: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels
type ChannelName string

const (
	ChannelNameMatches ChannelName = "matches"
)

// SubscribeRequest can be used to [Subscribe to Coinbase Channels].
//
// [Subscribe to Coinbase Channels]: https://docs.cloud.coinbase.com/exchange/docs/websocket-overview#subscribe
type SubscribeRequest struct {
	Type       string                    `json:"type"`
	ProductIDs []ProductID               `json:"product_ids"`
	Channels   []SubscribeChannelRequest `json:"channels"`
}

// SubscribeChannelRequest nests under SubscribeRequest.
type SubscribeChannelRequest struct {
	Name       ChannelName `json:"name"`
	ProductIDs []ProductID `json:"product_ids"`
}

// Match is a [Coinbase Match].
//
// [Coinbase Match]: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#match
type Match struct {
	Type      MessageType `json:"type"`
	ProductID ProductID   `json:"product_id"`
	Size      string      `json:"size"`
	Price     string      `json:"price"`
	Message   string      `json:"message"`
}

// MatchResponse represents a Coinbase [Match] Message returned over the websocket.
//
// [Match]: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#match
type MatchResponse struct {
	// Match is populated if a valid Match message is received.
	Match Match

	// Err is populated in the event Match is not.
	Err error
}

// ToUnitsAndUnitPrice parses units & price from the MatchResponse.
func (m *MatchResponse) ToUnitsAndUnitPrice() (units, unitPrice float64, err error) {
	if m.Err != nil {
		return 0, 0, fmt.Errorf("match response: %w", m.Err)
	}

	units, err = strconv.ParseFloat(m.Match.Size, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse units from size: %w", err)
	}

	unitPrice, err = strconv.ParseFloat(m.Match.Price, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse unitPrice from price: %w", err)
	}

	return
}
