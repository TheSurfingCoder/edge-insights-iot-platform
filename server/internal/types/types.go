package types

import (
	"time"
)

// LogMessage represents an IoT device log entry
// Update LogMessage struct
type LogMessage struct {
	Time       time.Time `json:"time"`
	DeviceID   string    `json:"device_id"`
	DeviceType string    `json:"device_type"`
	Location   string    `json:"location"`
	RawValue   *float64  `json:"raw_value,omitempty"` // Pointer so it can be nil
	Unit       string    `json:"unit,omitempty"`
	LogType    string    `json:"log_type"`
	Message    string    `json:"message"`
}

// LogResponse represents the response after processing a log
type LogResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// QueryRequest represents a natural language query request
type QueryRequest struct {
	Query string `json:"query"`
}

// QueryResponse represents the AI query response
type QueryResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error,omitempty"`
	Query   string      `json:"query"`
	Time    time.Time   `json:"time"`
}

// SearchResult represents a single search result with distance score
type SearchResult struct {
	EmbeddingUUID string   `json:"embedding_uuid"`
	Time          string   `json:"time"`
	DeviceID      string   `json:"device_id"`
	DeviceType    string   `json:"device_type"`
	Location      string   `json:"location"`
	LogType       string   `json:"log_type"`
	ChunkSeq      int      `json:"chunk_seq"`
	Chunk         string   `json:"chunk"`
	Distance      float64  `json:"distance"`
	RawValue      *float64 `json:"raw_value,omitempty"`
	Unit          string   `json:"unit,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
	Query   string         `json:"query"`
}

type SummaryResponse struct {
	Summary     string   `json:"summary"`
	TimeRange   string   `json:"time_range"`
	LogCount    int      `json:"log_count"`
	KeyInsights []string `json:"key_insights"`
}

// AnomalyResponse represents detected anomalies
type AnomalyResponse struct {
	Anomalies  []Anomaly `json:"anomalies"`
	TotalFound int       `json:"total_found"`
	TimeRange  string    `json:"time_range"`
}

// Anomaly represents a single detected anomaly
type Anomaly struct {
	Time       time.Time `json:"time"`
	DeviceID   string    `json:"device_id"`
	Type       string    `json:"type"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	Confidence float64   `json:"confidence"`
}
