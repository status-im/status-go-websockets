package main

import (
	"crypto/sha1"
	"encoding/base64"
	"net/http"
)

const (
	HeaderConnection      = "Connection"
	HeaderSecWebSocketKey = "Sec-WebSocket-Key"
	HeaderUpgrade         = "Upgrade"

	websocket = "websocket"

	WebsocketsMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

// handleWebSocket is the handler passed into a http server
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(HeaderUpgrade) != websocket || r.Header.Get(HeaderConnection) != HeaderUpgrade {
		http.Error(w, "Not a WebSocket handshake", http.StatusBadRequest)
		return
	}

	// Perform WebSocket handshake
	secWebSocketKey := r.Header.Get(HeaderSecWebSocketKey)
	secWebSocketAccept := computeAcceptKey(secWebSocketKey)

	headers := http.Header{}
	headers.Set(HeaderUpgrade, websocket)
	headers.Set(HeaderConnection, HeaderUpgrade)
	headers.Set(HeaderSecWebSocketKey, secWebSocketAccept)

	for k, v := range headers {
		w.Header()[k] = v
	}
}

// computeAcceptKey is used to generate the `secWebSocketAccept`
// Part of the websockets protocol see https://datatracker.ietf.org/doc/html/rfc6455#section-1.3
func computeAcceptKey(key string) string {
	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(key + WebsocketsMagicString))
	return base64.StdEncoding.EncodeToString(sha1Hash.Sum(nil))
}
