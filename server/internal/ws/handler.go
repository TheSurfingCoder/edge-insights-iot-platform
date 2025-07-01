//listens for websocket messages. handles what happens next like logging it, validating it, storing it in the database
//websocket connections start as http. then it upgrades that to a websocket connection

package ws

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"edge-insights/internal/types"

	"edge-insights/internal/db"

	"github.com/gorilla/websocket"
)

// upgrader is a WebSocket upgrader that converts HTTP connections to WebSocket connections
// CheckOrigin: true allows all origins (useful for development, should be restricted in production)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Handler manages WebSocket connections and processes IoT log messages
type Handler struct {
	db           *sql.DB
	clients      map[*websocket.Conn]bool
	clientsMutex sync.RWMutex
}

// NewHandler creates a new WebSocket handler with database connection
func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db:      db,
		clients: make(map[*websocket.Conn]bool),
	}
}

// broadcastToClients sends a message to all connected clients
func (h *Handler) broadcastToClients(message interface{}) {
	h.clientsMutex.RLock()

	// Collect clients to remove
	var clientsToRemove []*websocket.Conn

	for client := range h.clients {
		if err := client.WriteJSON(message); err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			clientsToRemove = append(clientsToRemove, client)
		}
	}

	h.clientsMutex.RUnlock()

	// Remove failed clients
	if len(clientsToRemove) > 0 {
		h.clientsMutex.Lock()
		for _, client := range clientsToRemove {
			delete(h.clients, client)
		}
		h.clientsMutex.Unlock()
	}
}

// HandleWebSocket manages the WebSocket connection lifecycle:
// 1. Upgrades HTTP connection to WebSocket
// 2. Listens for incoming log messages
// 3. Validates and stores logs in database
// 4. Sends responses back to client
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Add client to the list of connected clients
	h.clientsMutex.Lock()
	h.clients[conn] = true
	h.clientsMutex.Unlock()

	// Remove client when connection closes
	defer func() {
		h.clientsMutex.Lock()
		delete(h.clients, conn)
		h.clientsMutex.Unlock()
		conn.Close()
	}()

	log.Printf("New WebSocket connection established. Total clients: %d", len(h.clients))

	// Main message processing loop
	for {
		// Read message from WebSocket client
		// messageType: type of message (text, binary, etc.)
		// message: the actual message content
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break // Exit loop if connection is closed or error occurs
		}

		// Parse JSON message into LogMessage struct (this is from types.go)
		var logMsg types.LogMessage
		if err := json.Unmarshal(message, &logMsg); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			sendError(conn, "Invalid JSON format")
			continue // Continue to next message instead of breaking
		}

		// Validate the log message (check required fields)
		if err := validateLogMessage(logMsg); err != nil {
			log.Printf("Validation error: %v", err)
			sendError(conn, err.Error())
			continue
		}

		// Store the validated log in TimescaleDB
		if err := h.storeLog(logMsg); err != nil {
			log.Printf("Error storing log: %v", err)
			sendError(conn, "Failed to store log")
			continue
		}

		// Send success response back to the sender
		sendSuccess(conn, "Log stored successfully")

		// Broadcast the log data to all connected clients for live feed
		h.broadcastToClients(map[string]interface{}{
			"type": "log_entry",
			"data": logMsg,
		})
	}
}

// validateLogMessage checks if all required fields are present and valid
func validateLogMessage(log types.LogMessage) error {
	if log.DeviceID == "" {
		return fmt.Errorf("device_id is required")
	}
	if log.LogType == "" {
		return fmt.Errorf("log_type is required")
	}
	if log.LogType == "" {
		return fmt.Errorf("LogType is required")
	}
	// If time is not provided, use current time
	if log.Time.IsZero() {
		log.Time = time.Now()
	}
	return nil
}

// storeLog inserts a log message into the TimescaleDB device_logs table
func (h *Handler) storeLog(log types.LogMessage) error {
	return db.StoreSensorReading(h.db, log)
}

// sendSuccess sends a success response to the WebSocket client
// log response is from types.go
func sendSuccess(conn *websocket.Conn, message string) {
	response := types.LogResponse{
		Success: true,
		Message: message,
	}

	// Convert response struct to JSON and send
	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending success response: %v", err)
	}
}

// sendError sends an error response to the WebSocket client
func sendError(conn *websocket.Conn, errorMsg string) {
	response := types.LogResponse{
		Success: false,
		Message: "Error processing log",
		Error:   errorMsg,
	}

	// Convert response struct to JSON and send
	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending error response: %v", err)
	}
}
