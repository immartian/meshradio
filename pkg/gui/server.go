package gui

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/meshradio/meshradio/internal/broadcaster"
	"github.com/meshradio/meshradio/internal/listener"
	"github.com/meshradio/meshradio/pkg/audio"
)

//go:embed web/*
var webFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local use
	},
}

// Server manages the web GUI
type Server struct {
	webPort     int
	audioPort   int
	callsign    string
	ipv6        net.IP
	broadcaster *broadcaster.Broadcaster
	listener    *listener.Listener
	targetIPv6  net.IP // Target IPv6 when listening
	clients     map[*websocket.Conn]bool
	clientsMu   sync.Mutex
	broadcast   chan StatusUpdate
}

// StatusUpdate represents the current system status
type StatusUpdate struct {
	Timestamp   int64  `json:"timestamp"`
	Callsign    string `json:"callsign"`
	IPv6        string `json:"ipv6"`
	Mode        string `json:"mode"` // idle, broadcasting, listening
	Station     string `json:"station,omitempty"`
	PacketCount uint64 `json:"packetCount"`
	SignalQuality uint8 `json:"signalQuality"`
}

// NewServer creates a new web GUI server
func NewServer(webPort int, callsign string, ipv6 net.IP) *Server {
	return &Server{
		webPort:   webPort,
		audioPort: 8799, // Default audio port
		callsign:  callsign,
		ipv6:      ipv6,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan StatusUpdate, 10),
	}
}

// SetAudioPort sets the audio port for broadcasting/listening
func (s *Server) SetAudioPort(port int) {
	s.audioPort = port
}

// Start starts the web server
func (s *Server) Start() error {
	// Serve embedded static files
	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("failed to get web root: %w", err)
	}

	http.Handle("/", http.FileServer(http.FS(webRoot)))
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/api/broadcast/start", s.handleBroadcastStart)
	http.HandleFunc("/api/broadcast/stop", s.handleBroadcastStop)
	http.HandleFunc("/api/listen/start", s.handleListenStart)
	http.HandleFunc("/api/listen/stop", s.handleListenStop)
	http.HandleFunc("/api/status", s.handleStatus)

	// Start status broadcaster
	go s.statusBroadcaster()

	addr := fmt.Sprintf(":%d", s.webPort)
	log.Printf("🌐 Web GUI available at http://localhost:%d", s.webPort)

	return http.ListenAndServe(addr, nil)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
	}()

	// Send initial status
	s.sendStatus(conn)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// statusBroadcaster sends periodic updates to all clients
func (s *Server) statusBroadcaster() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := s.getStatus()

		s.clientsMu.Lock()
		for client := range s.clients {
			err := client.WriteJSON(status)
			if err != nil {
				client.Close()
				delete(s.clients, client)
			}
		}
		s.clientsMu.Unlock()
	}
}

// getStatus returns the current system status
func (s *Server) getStatus() StatusUpdate {
	status := StatusUpdate{
		Timestamp: time.Now().Unix(),
		Callsign:  s.callsign,
		IPv6:      s.ipv6.String(),
		Mode:      "idle",
	}

	if s.broadcaster != nil && s.broadcaster.IsRunning() {
		status.Mode = "broadcasting"
	} else if s.listener != nil && s.listener.IsRunning() {
		status.Mode = "listening"
		packets, _, station := s.listener.GetStats()
		// If no station callsign, show target IPv6 instead
		if station == "" || station == "unknown" {
			if s.targetIPv6 != nil {
				status.Station = s.targetIPv6.String()
			} else {
				status.Station = "unknown"
			}
		} else {
			status.Station = station
		}
		status.PacketCount = packets
	}

	return status
}

// sendStatus sends current status to a client
func (s *Server) sendStatus(conn *websocket.Conn) {
	status := s.getStatus()
	conn.WriteJSON(status)
}

// handleBroadcastStart starts broadcasting
func (s *Server) handleBroadcastStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.broadcaster != nil && s.broadcaster.IsRunning() {
		json.NewEncoder(w).Encode(map[string]string{"error": "Already broadcasting"})
		return
	}

	cfg := broadcaster.Config{
		Callsign:    s.callsign,
		IPv6:        s.ipv6,
		Port:        s.audioPort,
		Group:       "default", // TODO: Allow user to select group via API
		AudioConfig: audio.DefaultConfig(),
	}

	b, err := broadcaster.New(cfg)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := b.Start(); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	s.broadcaster = b
	json.NewEncoder(w).Encode(map[string]string{"status": "broadcasting"})
}

// handleBroadcastStop stops broadcasting
func (s *Server) handleBroadcastStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.broadcaster != nil {
		s.broadcaster.Stop()
		s.broadcaster = nil
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// handleListenStart starts listening
func (s *Server) handleListenStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IPv6 string `json:"ipv6"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	targetIPv6 := net.ParseIP(req.IPv6)
	if targetIPv6 == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid IPv6 address"})
		return
	}

	cfg := listener.Config{
		Callsign:    s.callsign,
		LocalIPv6:   s.ipv6,
		TargetIPv6:  targetIPv6,
		TargetPort:  s.audioPort,
		LocalPort:   s.audioPort,
		Group:       "default", // TODO: Allow user to select group via API
		SSMSource:   nil, // Regular multicast (receive from all sources)
		AudioConfig: audio.DefaultConfig(),
	}

	l, err := listener.New(cfg)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := l.Start(); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	s.listener = l
	s.targetIPv6 = targetIPv6 // Store target IPv6 for display
	json.NewEncoder(w).Encode(map[string]string{"status": "listening", "target": req.IPv6})
}

// handleListenStop stops listening
func (s *Server) handleListenStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.listener != nil {
		s.listener.Stop()
		s.listener = nil
		s.targetIPv6 = nil // Clear target IPv6
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// handleStatus returns current status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := s.getStatus()
	json.NewEncoder(w).Encode(status)
}
