package main

import (
	"crypto/sha1"
	"encoding/base64"
)

// computeAcceptKey is used to generate the `secWebSocketAccept`
// Part of the websockets protocol see https://datatracker.ietf.org/doc/html/rfc6455#section-1.3
func computeAcceptKey(key string) string {
	magicString := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(key + magicString))
	return base64.StdEncoding.EncodeToString(sha1Hash.Sum(nil))
}
