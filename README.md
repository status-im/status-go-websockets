# WebSocket Server with TLS

This repository contains a simple WebSocket server implemented in Go that uses TLS for secure communication. This server listens for WebSocket connections on `wss://localhost:8443/ws`.

## Prerequisites

- [Go](https://golang.org/doc/install) (version 1.16 or later)
- [OpenSSL](https://www.openssl.org/)
- [Node.js](https://nodejs.org/) (with npm)
- [wscat](https://github.com/websockets/wscat)

## Setup

### Generate TLS Certificates

- Navigate to the repo's directory.
- Create a self-signed certificate and key, use the following OpenSSL command:

```sh
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 7 -nodes
```

### Install `wscat`
Install wscat globally using npm:

```sh
npm install -g wscat
```

## Running the WebSocket Server

Ensure you are in this repo's dir.

### Run the WebSocket server:

```sh
go run main.go
```

You should see output like below, any message sent via the WS client will also be printed in the server CLI:

```log
2024/06/12 14:57:06 WebSocket server started on wss://localhost:8443/ws
2024/06/12 14:57:27 Received message: hello
2024/06/12 14:57:34 Received message: I'm a goat
2024/06/12 14:57:39 Received message: Welcome
```

## Testing the WebSocket Server
To test the WebSocket server using `wscat`, run the following command:

```sh
wscat -c wss://localhost:8443/ws --no-check
```

Ensure the WebSocket server is running.

Send a message: Type a message in the wscat terminal and press Enter. You should see the message echoed back.

```log
 % wscat -c wss://localhost:8443/ws --no-check

Connected (press CTRL+C to quit)
> hello
< hello
> I'm a goat
< I'm a goat
> Welcome
< Welcome
> %
```

## Note
The `--no-check` flag for `wscat` is used to bypass certificate validation, which is necessary when using self-signed certificates.
`status-go` makes extensive use of self-signed certificates, we ensure the integrity of our local connections by sharing the cert data via side channels.

## License
This project is licensed under the MIT License.