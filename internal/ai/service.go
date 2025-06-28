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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"edge-insights/internal/types"

	"github.com/sashabaranov/go-openai"
	"github.com/pgvector/pgvector-go"
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

	// Step 2: Perform vector similarity search using pgvector
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

	// Step 3: Collect results with distance scores
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

	// Step 4: Format results as JSON string for the response
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		log.Printf("Failed to marshal results: %v", err)
		return nil, fmt.Errorf("failed to marshal results: %w", err)
	}

	log.Printf("Semantic search completed successfully, found %d results", len(results))
	return &types.QueryResponse{
		Success: true,
		Result:  string(resultsJSON),
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
// This function uses TimescaleDB's AI capabilities to understand and respond to
// natural language questions about the device logs
func (s *AIService) QueryLogs(query string) (*types.QueryResponse, error) {
	log.Printf("Processing AI query: %s", query)

	// TODO: Implement proper AI querying using vector search + LLM
	// For now, return a placeholder response
	return &types.QueryResponse{
		Success: true,
		Result:  fmt.Sprintf("AI query functionality not yet implemented. Query: %s", query),
		Query:   query,
		Time:    time.Now(),
	}, nil
}

// SummarizeLogs generates AI-powered summaries of recent logs
// This function provides intelligent summaries of log activity over different time periods
func (s *AIService) SummarizeLogs(timeRange string) (*types.QueryResponse, error) {
	log.Printf("Generating log summary for time range: %s", timeRange)

	// TODO: Implement proper AI summarization using vector search + LLM
	// For now, return a placeholder response
	return &types.QueryResponse{
		Success: true,
		Result:  fmt.Sprintf("AI summarization functionality not yet implemented. Time range: %s", timeRange),
		Query:   fmt.Sprintf("Summarize logs from last %s", timeRange),
		Time:    time.Now(),
	}, nil
}

// DetectAnomalies uses AI to identify unusual patterns in device logs
// This function analyzes logs to find potential issues or abnormal behavior
func (s *AIService) DetectAnomalies() (*types.QueryResponse, error) {
	log.Printf("Detecting anomalies in recent logs")

	// TODO: Implement proper AI anomaly detection using vector search + LLM
	// For now, return a placeholder response
	return &types.QueryResponse{
		Success: true,
		Result:  "AI anomaly detection functionality not yet implemented.",
		Query:   "Detect anomalies in recent logs",
		Time:    time.Now(),
	}, nil
}
