package coinbase

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

type Conn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
}

type Dialer interface {
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (Conn, *http.Response, error)
}

type GorillaWebsocketDialler struct {
	d *websocket.Dialer
}

func (g *GorillaWebsocketDialler) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (Conn, *http.Response, error) {
	return g.d.DialContext(ctx, urlStr, requestHeader)
}

func NewGorillaWebsocketDialler(d *websocket.Dialer) Dialer {
	if d == nil {
		d = websocket.DefaultDialer
	}

	return &GorillaWebsocketDialler{d: d}
}
