package coinbase

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

// Conn should represent a Websocket connection.
type Conn interface {
	// ReadJSON should read the next JSON-encoded message from the connection and
	// store it in the value pointed to by v. If this method returns an error, it
	// will not be called again.
	ReadJSON(v interface{}) error

	// WriteJSON should write the JSON encoding of v as a message.
	WriteJSON(v interface{}) error

	// Close should close the underlying network connection without sending or waiting
	// for a close message.
	Close() error
}

// A Dialer should contain options for connecting to WebSocket server.
type Dialer interface {
	// DialContext should create a new client connection.
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (Conn, *http.Response, error)
}

type gorillaWebsocketDialler struct {
	d *websocket.Dialer
}

var _ Dialer = (*gorillaWebsocketDialler)(nil)

func (g *gorillaWebsocketDialler) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (Conn, *http.Response, error) {
	return g.d.DialContext(ctx, urlStr, requestHeader)
}

func newGorillaWebsocketDialler(d *websocket.Dialer) Dialer {
	if d == nil {
		d = websocket.DefaultDialer
	}

	return &gorillaWebsocketDialler{d: d}
}
