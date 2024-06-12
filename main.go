package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
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
	header := make([]byte, 2)
	if _, err := buf.Read(header); err != nil {
		return 0, nil, err
	}

	fin := header[0] & 0x80
	if fin == 0 {
		return 0, nil, fmt.Errorf("fragmented messages are not supported")
	}

	opcode := header[0] & 0x0F
	mask := header[1] & 0x80
	payloadLen := int(header[1] & 0x7F)

	if mask == 0 {
		return 0, nil, fmt.Errorf("unmasked messages from the client are not supported")
	}

	if payloadLen == 126 {
		extendedPayloadLen := make([]byte, 2)
		if _, err := buf.Read(extendedPayloadLen); err != nil {
			return 0, nil, err
		}
		payloadLen = int(extendedPayloadLen[0])<<8 | int(extendedPayloadLen[1])
	} else if payloadLen == 127 {
		extendedPayloadLen := make([]byte, 8)
		if _, err := buf.Read(extendedPayloadLen); err != nil {
			return 0, nil, err
		}
		payloadLen = int(extendedPayloadLen[0])<<56 | int(extendedPayloadLen[1])<<48 |
			int(extendedPayloadLen[2])<<40 | int(extendedPayloadLen[3])<<32 |
			int(extendedPayloadLen[4])<<24 | int(extendedPayloadLen[5])<<16 |
			int(extendedPayloadLen[6])<<8 | int(extendedPayloadLen[7])
	}

	maskingKey := make([]byte, 4)
	if _, err := buf.Read(maskingKey); err != nil {
		return 0, nil, err
	}

	payload = make([]byte, payloadLen)
	if _, err := buf.Read(payload); err != nil {
		return 0, nil, err
	}

	for i := 0; i < payloadLen; i++ {
		payload[i] ^= maskingKey[i%4]
	}

	return int(opcode), payload, nil
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
