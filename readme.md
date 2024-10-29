# Golang WebSocket Chat Server

This project is a simple WebSocket chat server implemented in Go. It allows multiple clients to connect, set their usernames, and exchange messages in real-time.

Got the idea to do this thing from [tsoding](https://github.com/rexim) after he created one in [c3lang](https://c3-lang.org/)

Really fun project, mostly inspired from watching [this](https://www.youtube.com/watch?v=-PfG87485Po) video. Decided to go with golang, cause I wanted to do something with this language for a while as it seems really cool. Not sure if this is the best lang to implemetn a wss in, but it was super fun.

Will probably use in some future project

## Features

- WebSocket connection handling
- Username setting
- Broadcasting messages to all connected clients
- Handling of text, ping, and close frames

## Getting Started

### Prerequisites

- Go 1.16 or later

### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/MarcelOlsen/golang-chat-wss.git
   cd golang-chat-wss
   ```

2. Build the server:

   ```sh
   go build -o chat-server
   ```

3. Run the server:

   ```sh
   ./chat-server
   ```

   The server will start on port `8080`.

### Or simply:

```sh
go run server.go
```

### Usage

1. Open a WebSocket client (e.g., a browser or a WebSocket client tool) and connect to `ws://localhost:8080/ws`.

2. Send your username as the first message.

3. Start chatting! All messages will be broadcast to all connected clients.

## Code Overview

- `main.go`: Contains the main server logic, including WebSocket connection handling and message broadcasting.
- `Client`: Represents a connected client.
- `handleWebSocket`: Handles the WebSocket upgrade and initializes the client connection.
- `handleFrames`: Reads and processes WebSocket frames from the client.
- `broadcastMessage`: Sends a message to all connected clients.

## Acknowledgments

- Got the idea to do this thing from [tsoding](https://github.com/rexim) after he created one in [c3lang](https://c3-lang.org/)
- Inspired by various WebSocket tutorials and examples available online.
