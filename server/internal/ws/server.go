package ws

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"edge-insights/internal/ai"
	"edge-insights/internal/db"
	"edge-insights/internal/types"
)

type Server struct {
	db      *sql.DB
	port    string
	handler *Handler
	ai      *ai.AIService
}

func NewServer(db *sql.DB) *Server {
	port := getEnv("SERVER_PORT", "8080")
	return &Server{
		db:      db,
		port:    port,
		handler: NewHandler(db),
		ai:      ai.NewAIService(db),
	}
}


func enableCORS(w http.ResponseWriter, r *http.Request) {
    // Get allowed origins from environment variable
    allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
    
    if allowedOrigins == "" {
        // Default to localhost for development
        allowedOrigins = "http://localhost:3000,http://localhost:3001"
    }
    
    // Parse the origins string (comma-separated)
    origins := strings.Split(allowedOrigins, ",")
    
    // Get the requesting origin
    origin := r.Header.Get("Origin")
    
    // Check if the requesting origin is in our allowed list
    for _, allowedOrigin := range origins {
        if strings.TrimSpace(allowedOrigin) == origin {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            break
        }
    }
    
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}

//CORS middleware wrapper - handles all requests
func corsMiddleware(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Handle preflight OPTIONS request
        if r.Method == "OPTIONS" {
            enableCORS(w, r)
            w.WriteHeader(http.StatusOK)
            return
        }

        // Enable CORS for all requests (GET, POST, etc.)
        enableCORS(w, r)
        
        // Call the actual handler
        handler(w, r)
    }
}


func (s *Server) Start() error {
	// WebSocket endpoint
	http.HandleFunc("/ws", s.handler.HandleWebSocket)

	 // Health check endpoint
	 http.HandleFunc("/health", corsMiddleware(s.healthHandler))


 // Log viewing endpoints (GET requests)
 http.HandleFunc("/api/logs", corsMiddleware(s.logsHandler))
 http.HandleFunc("/api/logs/device/", corsMiddleware(s.deviceLogsHandler))


	log.Printf("Starting WebSocket server on port %s", s.port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", s.port)
	log.Printf("Health check: http://localhost:%s/health", s.port)
	log.Printf("View logs: http://localhost:%s/api/logs", s.port)

	http.HandleFunc("/api/ai/query", corsMiddleware(s.aiQueryHandler))
    http.HandleFunc("/api/ai/summarize", corsMiddleware(s.aiSummarizeHandler))
    http.HandleFunc("/api/ai/anomalies", corsMiddleware(s.aiAnomaliesHandler))
    http.HandleFunc("/api/ai/search", corsMiddleware(s.aiSearchHandler))
	log.Printf("Starting WebSocket server on port %s", s.port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", s.port)
	log.Printf("Health check: http://localhost:%s/health", s.port)
	log.Printf("View logs: http://localhost:%s/api/logs", s.port)
	log.Printf("AI Query: http://localhost:%s/api/ai/query", s.port)

	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "edge-insights"}`))
}



func (s *Server) logsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := db.GetRecentSensorReadings(s.db, limit)
	if err != nil {
		log.Printf("Error fetching logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

func (s *Server) deviceLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract device ID from URL path
	deviceID := r.URL.Path[len("/api/logs/device/"):]
	if deviceID == "" {
		http.Error(w, "Device ID required", http.StatusBadRequest)
		return
	}

	limit := 20 // Default limit for device logs
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := db.GetLogsByDevice(s.db, deviceID, limit)
	if err != nil {
		log.Printf("Error fetching device logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_id": deviceID,
		"logs":      logs,
		"count":     len(logs),
	})
}

func (s *Server) aiQueryHandler(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body into QueryRequest struct
	var req types.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	//  Validate query is not empty
	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Call AI service (in service.go) with the query
	response, err := s.ai.QueryLogs(req.Query)
	if err != nil {
		log.Printf("AI query error: %v", err)
		http.Error(w, "AI query failed", http.StatusInternalServerError)
		return
	}

	//  Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) aiSummarizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "1h" // Default to 1 hour
	}

	response, err := s.ai.SummarizeLogs(timeRange)
	if err != nil {
		log.Printf("AI summary error: %v", err)
		http.Error(w, "AI summary failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) aiAnomaliesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response, err := s.ai.DetectAnomalies()
	if err != nil {
		log.Printf("AI anomaly detection error: %v", err)
		http.Error(w, "AI anomaly detection failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ... existing code ...
func (s *Server) aiSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON body
	var req struct {
		SearchText string `json:"search_text"`
		Limit      int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.SearchText == "" {
		http.Error(w, "Search text is required", http.StatusBadRequest)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10 // Default limit
	}

	response, err := s.ai.SearchSimilarLogs(req.SearchText, req.Limit)
	if err != nil {
		log.Printf("AI search error: %v", err)
		http.Error(w, "AI search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ... existing code ...
