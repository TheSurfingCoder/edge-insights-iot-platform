/*
EDGE INSIGHTS - Real-Time IoT Log Intelligence System with AI-Powered Analytics

PROJECT STRUCTURE AND PURPOSE:

DIRECTORIES:
/cmd/server/          - Application entry point and main server logic
/internal/            - Core application logic and business rules
  ├── /db/            - Database connection, migrations, and query functions
  ├── /ws/            - WebSocket server, handlers, and HTTP endpoints
  └── /ai/            - AI service integration with pgAI and OpenAI
/migrations/          - SQL migration files for database schema management
/scripts/             - Utility scripts and tools
  └── /simulator/     - IoT device log simulator for testing

OPERATIONAL FILES:

/cmd/server/main.go
    PURPOSE: Application entry point that initializes the entire system
    RESPONSIBILITIES:
    - Load environment configuration
    - Establish database connection to TimescaleDB Cloud
    - Run database migrations
    - Start WebSocket server
    - Orchestrate all system components

/internal/db/connection.go
    PURPOSE: Database connection management and configuration
    RESPONSIBILITIES:
    - Load database credentials from environment variables
    - Establish connection to TimescaleDB Cloud
    - Handle connection pooling and timeouts
    - Provide database configuration struct

/internal/db/migrations.go
    PURPOSE: Database schema management and migration execution
    RESPONSIBILITIES:
    - Execute SQL migration files in order
    - Create device_logs hypertable with TimescaleDB features
    - Add indexes and compression policies
    - Ensure database schema consistency

/internal/db/queries.go
    PURPOSE: Database query functions for log retrieval
    RESPONSIBILITIES:
    - Get recent logs with pagination
    - Retrieve device-specific logs
    - Format query results as structured data
    - Handle database query errors

/internal/ws/types.go
    PURPOSE: Data structures and types for WebSocket communication
    RESPONSIBILITIES:
    - Define LogMessage struct for IoT device logs
    - Define LogResponse struct for server responses
    - Define QueryRequest/QueryResponse for AI queries
    - Provide consistent data structures across the application

/internal/ws/handler.go
    PURPOSE: WebSocket connection handling and log processing
    RESPONSIBILITIES:
    - Accept WebSocket connections from IoT devices
    - Parse and validate incoming JSON log messages
    - Store validated logs in TimescaleDB
    - Send success/error responses to clients
    - Handle connection lifecycle and error management

/internal/ws/server.go
    PURPOSE: Creates and manages a comprehensive HTTP server with WebSocket and REST API capabilities
    RESPONSIBILITIES:
    - Initialize and configure HTTP server with custom port settings
    - Create and register WebSocket handler for real-time IoT device connections (/ws)
    - Set up REST API endpoints for log retrieval (/api/logs, /api/logs/device/*)
    - Implement AI analysis endpoints (/api/ai/query, /api/ai/summarize, /api/ai/anomalies)
    - Provide health check endpoint for monitoring and load balancers (/health)
    - Handle HTTP routing and request/response processing
    - Manage server lifecycle and graceful shutdown
    - Configure CORS and content-type headers for API responses

/internal/ai/service.go
    PURPOSE: Provides AI analysis functions for the WebSocket server
    RESPONSIBILITIES:
    - Exports QueryLogs() function for natural language log queries
    - Exports SummarizeLogs() function for AI-powered log summaries
    - Exports DetectAnomalies() function for identifying unusual patterns
    - Exports SearchSimilarLogs() function for semantic log search
    - Provides AIService struct that other packages can instantiate
    - Handles database queries using TimescaleDB's pgAI functions
    - Returns structured QueryResponse objects for API responses

/migrations/001_create_device_logs_table.sql
    PURPOSE: Initial database schema creation
    RESPONSIBILITIES:
    - Create device_logs table with time-series columns
    - Convert table to TimescaleDB hypertable
    - Add performance indexes
    - Set up compression policies

/scripts/main.go
    PURPOSE: IoT device log simulator for testing and demonstration
    RESPONSIBILITIES:
    - Simulate realistic IoT devices (sensors, cameras, controllers)
    - Generate diverse log types (INFO, WARN, ERROR, DEBUG)
    - Send logs via WebSocket to the server
    - Provide real-time feedback on log transmission
    - Create test data for AI analysis and vectorization

DATA FLOW:
IoT Simulator → WebSocket Server → Database Storage → AI Analysis → Vector Embeddings

TECHNOLOGIES:
- Go (Backend)
- TimescaleDB Cloud (Time-series database)
- WebSocket (Real-time communication)
- pgAI (Natural language processing)
- OpenAI Embeddings (Vector search)
- Docker (pgAI installation)

AUTHOR: Edge Insights Project
VERSION: 1.0
*/

package main

import (
	"log"

	"edge-insights/internal/db"
	"edge-insights/internal/ws"

	"edge-insights/internal/ai"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load database configuration
	config := db.LoadConfig()

	// Connect to database
	database, err := db.Connect(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Testing database connection...")
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM device_logs").Scan(&count)
	if err != nil {
		log.Printf("Error querying device_logs table: %v", err)
	} else {
		log.Printf("Current log count in database: %d", count)
	}

	// Test OpenAI embedding generation
	log.Println("Testing OpenAI embedding generation...")
	aiService := ai.NewAIService(database)
	if err := aiService.TestEmbeddingGeneration(); err != nil {
		log.Printf("OpenAI embedding test failed: %v", err)
	} else {
		log.Println("✅ OpenAI embedding generation test passed!")
	}

	log.Println("Edge Insights server initialized successfully")

	// Start WebSocket server
	server := ws.NewServer(database)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
