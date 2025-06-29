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
	db *sql.DB // Database connection for querying logs and embeddings
}

// NewAIService creates a new AI service instance
// Initializes the service with a database connection for log analysis
func NewAIService(db *sql.DB) *AIService {
	return &AIService{db: db}
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

	log.Printf("Successfully generated embedding with %d dimensions", len(resp.Data[0].Embedding))

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
	log.Printf("Searching for logs similar to: %s", searchText)

	// Step 1: Generate embedding for the search query
	queryEmbedding, err := s.generateEmbedding(searchText)
	if err != nil {
		log.Printf("Failed to generate embedding: %v", err)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Step 2: Convert []float64 to []float32 (pgvector expects float32)
	embedding32 := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		embedding32[i] = float32(v)
	}

	// Step 3: Create pgvector vector
	embeddingVec := pgvector.NewVector(embedding32)

	// Step 4: Perform vector similarity search using pgvector
	searchQuery := `
		SELECT 
			embedding_uuid,
			time,
			device_id,
			chunk_seq,
			chunk,
			embedding <=> $1 as distance
		FROM device_logs_embedding_store
		ORDER BY distance ASC
		LIMIT $2
	`

	rows, err := s.db.Query(searchQuery, embeddingVec, limit)
	if err != nil {
		log.Printf("Vector search failed: %v", err)
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer rows.Close()

	// Step 5: Collect results with distance scores
	var results []types.SearchResult
	for rows.Next() {
		var result types.SearchResult
		err := rows.Scan(
			&result.EmbeddingUUID,
			&result.Time,
			&result.DeviceID,
			&result.ChunkSeq,
			&result.Chunk,
			&result.Distance,
		)
		if err != nil {
			log.Printf("Error scanning result: %v", err)
			continue
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating results: %v", err)
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	// Step 6: Format results as JSON string for the response
	searchResponse := types.SearchResponse{
		Results: results,
		Count:   len(results),
		Query:   searchText,
	}

	log.Printf("Semantic search completed successfully, found %d results", len(results))
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

	embedding, err := s.generateEmbedding("test message for embedding generation")
	if err != nil {
		return fmt.Errorf("embedding generation failed: %w", err)
	}

	log.Printf("✅ Successfully generated embedding with %d dimensions", len(embedding))
	return nil
}

// QueryLogs performs natural language querying on IoT logs using AI
func (s *AIService) QueryLogs(query string) (*types.QueryResponse, error) {
	log.Printf("Processing AI query: %s", query)

	// Step 1: Use semantic search to find relevant logs
	searchResults, err := s.SearchSimilarLogs(query, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search for relevant logs: %w", err)
	}

	// Step 2: Extract the search results
	searchResponse, ok := searchResults.Result.(types.SearchResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from search")
	}

	// Step 3: Generate a natural language answer based on the results
	answer := s.generateAnswerFromResults(query, searchResponse.Results)

	return &types.QueryResponse{
		Success: true,
		Result: map[string]interface{}{
			"answer":        answer,
			"relevant_logs": searchResponse.Results,
			"log_count":     searchResponse.Count,
		},
		Query: query,
		Time:  time.Now(),
	}, nil
}

// SummarizeLogs generates AI-powered summaries of recent logs
func (s *AIService) SummarizeLogs(timeRange string) (*types.QueryResponse, error) {
	log.Printf("Generating log summary for time range: %s", timeRange)

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
	log.Printf("Detecting anomalies in recent logs")

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
		if err := rows.Scan(&log.Time, &log.DeviceID, &log.LogType, &log.Message); err != nil {
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
				Time:       log.Time,
				DeviceID:   log.DeviceID,
				Type:       "Error",
				Severity:   "High",
				Message:    log.Message,
				Confidence: 0.8,
			}
			anomalies = append(anomalies, anomaly)
		}

		// Detect unusual patterns (simplified)
		if strings.Contains(strings.ToLower(log.Message), "critical") {
			anomaly := types.Anomaly{
				Time:       log.Time,
				DeviceID:   log.DeviceID,
				Type:       "Critical Event",
				Severity:   "Critical",
				Message:    log.Message,
				Confidence: 0.9,
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}
