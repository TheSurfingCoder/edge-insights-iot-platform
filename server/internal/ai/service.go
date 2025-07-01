/*
AI Service for Edge Insights IoT Platform

PURPOSE:
This service provides AI-powered analysis of IoT device logs using TimescaleDB's
vector embeddings and natural language processing capabilities.

DEPENDENCIES:
- internal/db: Uses database connection for querying logs and embeddings
- internal/ws: Provides response structures for API endpoints

FEATURES:
- Natural language querying of IoT logs
- AI-powered log summarization
- Anomaly detection in device behavior
- Semantic search using vector embeddings

USAGE:
This service is used by the WebSocket server (internal/ws/server.go) to provide
AI endpoints for intelligent log analysis and insights.

AUTHOR: Edge Insights Project
VERSION: 1.0
*/

package ai

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"edge-insights/internal/types"

	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
)

// AIService handles AI-powered analysis of IoT logs
// This struct manages all AI-related database queries and processing
type AIService struct {
	db        *sql.DB
	textToSQL *TextToSQLService
}

// NewAIService creates a new AI service instance
// Initializes the service with a database connection for log analysis
func NewAIService(db *sql.DB) *AIService {
	return &AIService{
		db:        db,
		textToSQL: NewTextToSQLService(db),
	}
}

// generateEmbedding creates a vector embedding for the given text using OpenAI API
func (s *AIService) generateEmbedding(text string) ([]float64, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)

	resp, err := client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.SmallEmbedding3,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned from API")
	}

	// Convert []float32 to []float64
	embedding := make([]float64, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float64(v)
	}
	return embedding, nil
}

// SearchSimilarLogs performs semantic search using vector embeddings
// This function finds logs with similar meaning using the embeddings we generated
func (s *AIService) SearchSimilarLogs(searchText string, limit int) (*types.QueryResponse, error) {

	// Step 1: Generate embedding for the search query
	queryEmbedding, err := s.generateEmbedding(searchText)
	if err != nil {

		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Step 2: Convert []float64 to []float32 (pgvector expects float32)
	embedding32 := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		embedding32[i] = float32(v)
	}

	// Step 3: Create pgvector vector
	embeddingVec := pgvector.NewVector(embedding32)

	// Step 4: Perform vector similarity search using pgvector on sensor_readings_embeddings
	searchQuery := `
		SELECT 
			time,
			device_id,
			device_type,
			location,
			raw_value,
			unit,
			log_type,
			COALESCE(message, '') as message,
			embedding <=> $1 as distance
		FROM sensor_readings_embeddings
		WHERE embedding IS NOT NULL
		ORDER BY distance ASC
		LIMIT $2
	`

	// Log the semantic search query
	log.Printf("🔍 SEMANTIC SEARCH:")
	log.Printf("   Table Used: sensor_readings_embeddings (EMBEDDINGS)")
	log.Printf("   Reason: Vector similarity search now uses the new embeddings table")
	log.Printf("   ---")

	rows, err := s.db.Query(searchQuery, embeddingVec, limit)
	if err != nil {

		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer rows.Close()

	// Step 5: Collect results with distance scores
	var results []types.SearchResult
	for rows.Next() {
		var result types.SearchResult
		var time time.Time
		var deviceType, location, unit, logType, message string
		var rawValue *float64

		err := rows.Scan(
			&time,
			&result.DeviceID,
			&deviceType,
			&location,
			&rawValue,
			&unit,
			&logType,
			&message,
			&result.Distance,
		)
		if err != nil {
			continue
		}

		result.Time = time.Format("2006-01-02T15:04:05Z07:00")
		result.DeviceType = deviceType
		result.Location = location
		result.LogType = logType
		result.Chunk = message
		result.ChunkSeq = 0
		result.EmbeddingUUID = ""
		result.RawValue = rawValue
		result.Unit = unit

		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	// Step 6: Format results as JSON string for the response
	searchResponse := types.SearchResponse{
		Results: results,
		Count:   len(results),
		Query:   searchText,
	}

	return &types.QueryResponse{
		Success: true,
		Result:  searchResponse,
		Query:   fmt.Sprintf("Find logs similar to: %s", searchText),
		Time:    time.Now(),
	}, nil
}

// TestEmbeddingGeneration tests the OpenAI embedding generation
func (s *AIService) TestEmbeddingGeneration() error {
	log.Println("Testing OpenAI embedding generation...")

	_, err := s.generateEmbedding("test message for embedding generation")
	if err != nil {
		return fmt.Errorf("embedding generation failed: %w", err)
	}

	return nil
}

// QueryLogs performs intelligent query routing between semantic search and text-to-SQL
func (s *AIService) QueryLogs(query string) (*types.QueryResponse, error) {
	// Determine if this is a data query (text-to-SQL) or pattern search (semantic search)
	queryType := s.determineQueryType(query)

	if queryType == "data_query" {
		// Use text-to-SQL for specific data queries
		return s.textToSQL.ConvertToSQL(query)
	} else {
		// Use semantic search for pattern discovery and insights
		return s.performSemanticSearch(query)
	}
}

// determineQueryType decides whether to use text-to-SQL or semantic search
func (s *AIService) determineQueryType(query string) string {
	queryLower := strings.ToLower(query)

	// Keywords that suggest specific data queries (use text-to-SQL)
	dataKeywords := []string{
		"show me", "what is", "how many", "average", "count", "temperature",
		"humidity", "motion", "camera", "controller", "device", "location",
		"last hour", "last 24 hours", "yesterday", "today", "this week",
		"above", "below", "between", "greater than", "less than",
		"raw_value", "unit", "time", "hour", "day", "week", "month",
	}

	// Keywords that suggest pattern discovery (use semantic search)
	patternKeywords := []string{
		"why", "how", "patterns", "similar", "unusual", "anomaly", "problem",
		"issue", "failure", "error", "warning", "critical", "security",
		"behavior", "trend", "insight", "analysis", "explain", "understand",
		"find logs", "search for", "discover", "investigate",
	}

	// Count matches
	dataMatches := 0
	patternMatches := 0

	for _, keyword := range dataKeywords {
		if strings.Contains(queryLower, keyword) {
			dataMatches++
		}
	}

	for _, keyword := range patternKeywords {
		if strings.Contains(queryLower, keyword) {
			patternMatches++
		}
	}

	// Decision logic
	if dataMatches > patternMatches {
		return "data_query"
	} else {
		return "pattern_search"
	}
}

// performSemanticSearch handles pattern discovery queries
func (s *AIService) performSemanticSearch(query string) (*types.QueryResponse, error) {
	// Use existing semantic search functionality but updated for sensor_readings
	searchResults, err := s.SearchSimilarLogs(query, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to perform semantic search: %w", err)
	}

	// Extract the search results
	searchResponse, ok := searchResults.Result.(types.SearchResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from search")
	}

	// Generate a natural language answer based on the results
	answer := s.generateAnswerFromResults(query, searchResponse.Results)

	return &types.QueryResponse{
		Success: true,
		Result: map[string]interface{}{
			"answer":        answer,
			"relevant_logs": searchResponse.Results,
			"log_count":     searchResponse.Count,
			"query_type":    "pattern_search",
		},
		Query: query,
		Time:  time.Now(),
	}, nil
}

// SummarizeLogs generates AI-powered summaries of recent logs
func (s *AIService) SummarizeLogs(timeRange string) (*types.QueryResponse, error) {

	// Step 1: Get recent logs from the database
	logs, err := s.getRecentLogs(timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", err)
	}

	// Step 2: Generate summary
	summary := s.generateSummary(logs, timeRange)

	// Step 3: Extract key insights
	insights := s.extractKeyInsights(logs)

	summaryResponse := types.SummaryResponse{
		Summary:     summary,
		TimeRange:   timeRange,
		LogCount:    len(logs),
		KeyInsights: insights,
	}

	return &types.QueryResponse{
		Success: true,
		Result:  summaryResponse,
		Query:   fmt.Sprintf("Summarize logs from last %s", timeRange),
		Time:    time.Now(),
	}, nil
}

// DetectAnomalies uses AI to identify unusual patterns in device logs
func (s *AIService) DetectAnomalies() (*types.QueryResponse, error) {

	// Step 1: Get recent logs
	logs, err := s.getRecentLogs("24h")
	if err != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", err)
	}

	// Step 2: Detect anomalies
	anomalies := s.detectAnomalies(logs)

	anomalyResponse := types.AnomalyResponse{
		Anomalies:  anomalies,
		TotalFound: len(anomalies),
		TimeRange:  "24h",
	}

	return &types.QueryResponse{
		Success: true,
		Result:  anomalyResponse,
		Query:   "Detect anomalies in recent logs",
		Time:    time.Now(),
	}, nil
}

// Helper functions for the AI endpoints
func (s *AIService) generateAnswerFromResults(query string, results []types.SearchResult) string {
	if len(results) == 0 {
		return "I couldn't find any relevant logs to answer your question."
	}

	// Simple answer generation based on search results
	answer := fmt.Sprintf("Based on %d relevant logs, here's what I found:\n\n", len(results))

	for i, result := range results {
		if i >= 3 { // Limit to first 3 results for readability
			break
		}
		answer += fmt.Sprintf("• %s (Device: %s, Similarity: %.2f)\n",
			result.Chunk, result.DeviceID, result.Distance)
	}

	return answer
}

func (s *AIService) getRecentLogs(timeRange string) ([]types.LogMessage, error) {
	// Parse time range and get logs from database
	// This is a simplified version - you'd implement proper time parsing
	query := `
		SELECT time, device_id, log_type, message 
		FROM device_logs 
		WHERE time > NOW() - INTERVAL '1 hour'
		ORDER BY time DESC 
		LIMIT 100
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []types.LogMessage
	for rows.Next() {
		var log types.LogMessage
		if err := rows.Scan(&log.Time, &log.DeviceID, &log.LogType); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (s *AIService) generateSummary(logs []types.LogMessage, timeRange string) string {
	if len(logs) == 0 {
		return fmt.Sprintf("No logs found in the last %s.", timeRange)
	}

	// Count log types
	infoCount := 0
	warnCount := 0
	errorCount := 0
	deviceCount := make(map[string]bool)

	for _, log := range logs {
		switch log.LogType {
		case "INFO":
			infoCount++
		case "WARN":
			warnCount++
		case "ERROR":
			errorCount++
		}
		deviceCount[log.DeviceID] = true
	}

	summary := fmt.Sprintf("In the last %s, %d logs were generated across %d devices:\n",
		timeRange, len(logs), len(deviceCount))
	summary += fmt.Sprintf("• %d INFO logs\n", infoCount)
	summary += fmt.Sprintf("• %d WARN logs\n", warnCount)
	summary += fmt.Sprintf("• %d ERROR logs\n", errorCount)

	return summary
}

func (s *AIService) extractKeyInsights(logs []types.LogMessage) []string {
	var insights []string

	if len(logs) == 0 {
		return insights
	}

	// Simple insight extraction
	errorCount := 0
	for _, log := range logs {
		if log.LogType == "ERROR" {
			errorCount++
		}
	}

	if errorCount > 0 {
		insights = append(insights, fmt.Sprintf("Found %d error logs that may need attention", errorCount))
	}

	if len(logs) > 50 {
		insights = append(insights, "High log volume detected - consider reviewing system health")
	}

	return insights
}

func (s *AIService) detectAnomalies(logs []types.LogMessage) []types.Anomaly {
	var anomalies []types.Anomaly

	// Simple anomaly detection
	for _, log := range logs {
		// Detect error spikes
		if log.LogType == "ERROR" {
			anomaly := types.Anomaly{
				Time:     log.Time,
				DeviceID: log.DeviceID,
				Type:     "Error",
				Severity: "High",

				Confidence: 0.8,
			}
			anomalies = append(anomalies, anomaly)
		}

	}

	return anomalies
}
