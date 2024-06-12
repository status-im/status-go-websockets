package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net"
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
	// Perform WebSocket handshake
	if r.Header.Get(HeaderUpgrade) != websocket || r.Header.Get(HeaderConnection) != HeaderUpgrade {
		http.Error(w, "Not a WebSocket handshake", http.StatusBadRequest)
		return
	}

	secWebSocketKey := r.Header.Get(HeaderSecWebSocketKey)
	secWebSocketAccept := computeAcceptKey(secWebSocketKey)

	headers := http.Header{}
	headers.Set(HeaderUpgrade, websocket)
	headers.Set(HeaderConnection, HeaderUpgrade)
	headers.Set(HeaderSecWebSocketKey, secWebSocketAccept)

	for k, v := range headers {
		w.Header()[k] = v
	}

	// Hijacking things
	w.WriteHeader(http.StatusSwitchingProtocols)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, buf, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Main WebSocket loop
	// Handle WebSocket frames
	for {
		// Websocket read data
		messageType, payload, err := handleReadWebSocketData(buf)
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}
		log.Printf("Received message: %s", payload)

		// Websocket write data
		// in a real case the handler should process read data and return something useful to the client
		if err := handleWriteWebSocketData(conn, messageType, payload); err != nil {
			log.Println("Error writing WebSocket message:", err)
			break
		}
	}
}

// computeAcceptKey is used to generate the `secWebSocketAccept`
// Part of the websockets protocol see https://datatracker.ietf.org/doc/html/rfc6455#section-1.3
func computeAcceptKey(key string) string {
	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(key + WebsocketsMagicString))
	return base64.StdEncoding.EncodeToString(sha1Hash.Sum(nil))
}

func handleReadWebSocketData(buf *bufio.ReadWriter) (messageType int, payload []byte, err error) {
	// todo
	return 0, nil, nil
}

// handleWriteWebSocketData takes a WebSocket connection and writes data to it.
func handleWriteWebSocketData(conn net.Conn, messageType int, payload []byte) error {
	finAndOpcode := byte(0x80 | byte(messageType))
	maskAndPayloadLen := byte(len(payload))

	if _, err := conn.Write([]byte{finAndOpcode, maskAndPayloadLen}); err != nil {
		return err
	}

	if _, err := conn.Write(payload); err != nil {
		return err
	}

	return nil
}
