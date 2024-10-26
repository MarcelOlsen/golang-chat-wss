package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

const (
	TextFrame  = 0x1
	CloseFrame = 0x8
	PingFrame  = 0x9
	PongFrame  = 0xA
)

type Client struct {
	Conn     net.Conn
	Username string
}

var (
	connections = make(map[net.Conn]*Client)
	connMutex   = sync.Mutex{}
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" || r.Header.Get("Connection") != "Upgrade" {
		http.Error(w, "Invalid WebSocket request", http.StatusBadRequest)
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "Missing WebSocket key", http.StatusBadRequest)
		return
	}
	acceptKey := generateAcceptKey(key)
	w.Header().Set("Upgrade", "websocket")
	w.Header().Set("Connection", "Upgrade")
	w.Header().Set("Sec-WebSocket-Accept", acceptKey)
	w.WriteHeader(http.StatusSwitchingProtocols)

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "HTTPS hijacking is not supported", http.StatusInternalServerError)
		return
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "HTTP hijacking has failed", http.StatusInternalServerError)
		return
	}

	log.Println("WebSocket connection has been established")

	client := &Client{Conn: conn}

	connMutex.Lock()
	connections[conn] = client
	connMutex.Unlock()

	// Handle messages for this client
	go handleFrames(client)
}

func generateAcceptKey(key string) string {
	const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func handleFrames(client *Client) {
	conn := client.Conn
	defer func() {
		connMutex.Lock()
		delete(connections, conn)
		connMutex.Unlock()
		conn.Close()
		log.Printf("Connection closed for %s\n", client.Username)
	}()

	usernameSet := false

	for {
		frame, err := readFrame(conn)
		if err != nil {
			log.Println("Error reading frame:", err)
			break
		}

		switch frame.OpCode {
		case TextFrame:
			message := string(frame.Payload)
			if !usernameSet {
				// Set the username from the first message
				client.Username = message
				usernameSet = true
				log.Printf("Username set for connection: %s\n", client.Username)
				broadcastMessage(fmt.Sprintf("%s joined the chat", client.Username), "Server")
			} else {
				log.Printf("[%s]: %s\n", client.Username, message)
				broadcastMessage(message, client.Username)
			}
		case PingFrame:
			log.Println("Received ping")
			if err := writeFrame(conn, PongFrame, frame.Payload); err != nil {
				log.Println("Error writing pong frame:", err)
				break
			}
		case CloseFrame:
			log.Printf("%s disconnected", client.Username)
			broadcastMessage(fmt.Sprintf("%s left the chat", client.Username), "Server")
			return
		default:
			log.Printf("Unhandled frame type: %x\n", frame.OpCode)
		}
	}
}

// broadcastMessage sends the message to all connected clients with a username prefix
func broadcastMessage(message, username string) {
	formattedMessage := fmt.Sprintf("[%s]: %s", username, message)

	connMutex.Lock()
	defer connMutex.Unlock()

	for _, client := range connections {
		if err := writeFrame(client.Conn, TextFrame, []byte(formattedMessage)); err != nil {
			log.Printf("Error sending message to %s: %v\n", client.Username, err)
			client.Conn.Close()
			delete(connections, client.Conn)
		}
	}
}

type Frame struct {
	OpCode  byte
	Payload []byte
}

func readFrame(conn net.Conn) (Frame, error) {
	reader := bufio.NewReader(conn)
	frame := Frame{}

	firstByte, err := reader.ReadByte()
	if err != nil {
		return frame, err
	}
	frame.OpCode = firstByte & 0x0F

	secondByte, err := reader.ReadByte()
	if err != nil {
		return frame, err
	}
	payloadLen := int(secondByte & 0x7F)

	if payloadLen == 126 {
		var extendedLength uint16
		if err := binary.Read(reader, binary.BigEndian, &extendedLength); err != nil {
			return frame, err
		}
		payloadLen = int(extendedLength)
	} else if payloadLen == 127 {
		var extendedLength uint64
		if err := binary.Read(reader, binary.BigEndian, &extendedLength); err != nil {
			return frame, err
		}
		payloadLen = int(extendedLength)
	}

	mask := secondByte&0x80 != 0
	maskingKey := make([]byte, 4)
	if mask {
		if _, err := io.ReadFull(reader, maskingKey); err != nil {
			return frame, err
		}
	}

	frame.Payload = make([]byte, payloadLen)
	if _, err := io.ReadFull(reader, frame.Payload); err != nil {
		return frame, err
	}

	if mask {
		for i := 0; i < payloadLen; i++ {
			frame.Payload[i] ^= maskingKey[i%4]
		}
	}

	return frame, nil
}

func writeFrame(conn net.Conn, opCode byte, payload []byte) error {
	var buffer []byte

	buffer = append(buffer, 0x80|opCode)

	payloadLen := len(payload)
	if payloadLen <= 125 {
		buffer = append(buffer, byte(payloadLen))
	} else if payloadLen <= 65535 {
		buffer = append(buffer, 126)
		buffer = append(buffer, byte(payloadLen>>8), byte(payloadLen&0xFF))
	} else {
		buffer = append(buffer, 127)
		for i := 7; i >= 0; i-- {
			buffer = append(buffer, byte(payloadLen>>(i*8)))
		}
	}

	buffer = append(buffer, payload...)

	_, err := conn.Write(buffer)
	return err
}
