package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

// Event represents a live cluster event pushed over WebSocket.
type Event struct {
	Type      string      `json:"type"`      // e.g. "POD_UPDATED", "NODE_METRICS", "HUBBLE_FLOW", "TUNNEL_STATUS", "ALERT"
	Namespace string      `json:"namespace"` // namespace scope or ""
	Resource  string      `json:"resource"`  // resource kind
	Data      interface{} `json:"data"`      // payload
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a connected WebSocket subscriber.
type Client struct {
	hub  *Hub
	conn net.Conn
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts cluster events.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	log        *zap.Logger
	stopCh     chan struct{}
}

// NewHub creates a new WebSocket Hub.
func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		log:        log.Named("websocket"),
		stopCh:     make(chan struct{}),
	}
}

// Start runs the event broadcasting loop.
func (h *Hub) Start() {
	go h.run()
	go h.heartbeatLoop()
}

// Stop stops the WebSocket hub.
func (h *Hub) Stop() {
	close(h.stopCh)
}

func (h *Hub) run() {
	for {
		select {
		case <-h.stopCh:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				_ = client.conn.Close()
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.log.Debug("websocket client connected", zap.Int("totalClients", len(h.clients)))

			// Send welcome event
			welcome := Event{
				Type:      "CONNECTED",
				Resource:  "Cluster",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"cluster": "tarak-cluster-prod",
					"version": "v1.0.6",
					"status":  "Healthy",
					"mTLS":    "Active",
				},
			}
			raw, _ := json.Marshal(welcome)
			h.sendFrame(client.conn, raw)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				_ = client.conn.Close()
			}
			h.mu.Unlock()
			h.log.Debug("websocket client disconnected", zap.Int("totalClients", len(h.clients)))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Drop slow client
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent publishes an event to all connected dashboard clients.
func (h *Hub) BroadcastEvent(eventType, namespace, resource string, data interface{}) {
	evt := Event{
		Type:      eventType,
		Namespace: namespace,
		Resource:  resource,
		Data:      data,
		Timestamp: time.Now(),
	}
	raw, err := json.Marshal(evt)
	if err != nil {
		return
	}

	select {
	case h.broadcast <- raw:
	default:
	}
}

// ServeHTTP upgrades HTTP requests to WebSocket connection.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Verify WebSocket upgrade headers
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") ||
		!strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		http.Error(w, "Expected WebSocket Upgrade request", http.StatusBadRequest)
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "Missing Sec-WebSocket-Key", http.StatusBadRequest)
		return
	}

	// Calculate accept key
	hSha := sha1.New() //nolint:gosec
	hSha.Write([]byte(key + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(hSha.Sum(nil))

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write WebSocket handshake response
	handshake := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n\r\n", acceptKey)

	if _, err := bufrw.WriteString(handshake); err != nil {
		_ = conn.Close()
		return
	}
	if err := bufrw.Flush(); err != nil {
		_ = conn.Close()
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 128),
	}

	h.register <- client

	// Start write and read pumps
	go client.writePump(h)
	go client.readPump(h, bufrw.Reader)
}

func (c *Client) writePump(h *Hub) {
	ticker := time.NewTicker(20 * time.Second)
	defer func() {
		ticker.Stop()
		h.unregister <- c
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := h.sendFrame(c.conn, msg); err != nil {
				return
			}

		case <-ticker.C:
			// Send ping frame
			if err := h.sendPing(c.conn); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump(h *Hub, r *bufio.Reader) {
	defer func() {
		h.unregister <- c
	}()

	for {
		// Read frame header
		b1, err := r.ReadByte()
		if err != nil {
			return
		}
		opcode := b1 & 0x0F

		// Close frame
		if opcode == 0x8 {
			return
		}

		b2, err := r.ReadByte()
		if err != nil {
			return
		}

		isMasked := (b2 & 0x80) != 0
		length := int(b2 & 0x7F)

		if length == 126 {
			var l uint16
			var b [2]byte
			if _, err := io.ReadFull(r, b[:]); err != nil {
				return
			}
			l = uint16(b[0])<<8 | uint16(b[1])
			length = int(l)
		} else if length == 127 {
			var b [8]byte
			if _, err := io.ReadFull(r, b[:]); err != nil {
				return
			}
			length = int(b[7])
		}

		var mask [4]byte
		if isMasked {
			if _, err := io.ReadFull(r, mask[:]); err != nil {
				return
			}
		}

		payload := make([]byte, length)
		if _, err := io.ReadFull(r, payload); err != nil {
			return
		}

		if isMasked {
			for i := 0; i < length; i++ {
				payload[i] ^= mask[i%4]
			}
		}

		// Ping -> Respond with Pong
		if opcode == 0x9 {
			_ = h.sendPong(c.conn, payload)
		}
	}
}

func (h *Hub) sendFrame(conn net.Conn, payload []byte) error {
	length := len(payload)
	var header []byte

	if length <= 125 {
		header = []byte{0x81, byte(length)}
	} else if length <= 65535 {
		header = []byte{0x81, 126, byte(length >> 8), byte(length & 0xFF)}
	} else {
		header = []byte{0x81, 127, 0, 0, 0, 0,
			byte((length >> 24) & 0xFF),
			byte((length >> 16) & 0xFF),
			byte((length >> 8) & 0xFF),
			byte(length & 0xFF),
		}
	}

	if _, err := conn.Write(header); err != nil {
		return err
	}
	_, err := conn.Write(payload)
	return err
}

func (h *Hub) sendPing(conn net.Conn) error {
	_, err := conn.Write([]byte{0x89, 0x00})
	return err
}

func (h *Hub) sendPong(conn net.Conn, payload []byte) error {
	length := len(payload)
	if length > 125 {
		length = 125
	}
	header := []byte{0x8A, byte(length)}
	if _, err := conn.Write(header); err != nil {
		return err
	}
	if length > 0 {
		_, err := conn.Write(payload[:length])
		return err
	}
	return nil
}

func (h *Hub) heartbeatLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.mu.RLock()
			clientCount := len(h.clients)
			h.mu.RUnlock()

			if clientCount > 0 {
				h.BroadcastEvent("TELEMETRY_PULSE", "", "Cluster", map[string]interface{}{
					"status":      "HEALTHY",
					"activeNodes": 1,
					"time":        time.Now().Format(time.RFC3339),
				})
			}
		}
	}
}
