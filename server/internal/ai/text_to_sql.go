package ai

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"edge-insights/internal/types"

	"github.com/sashabaranov/go-openai"
)

// TextToSQLService handles natural language to SQL conversion
type TextToSQLService struct {
	db     *sql.DB
	openai *openai.Client
}

// NewTextToSQLService creates a new text-to-SQL service
func NewTextToSQLService(db *sql.DB) *TextToSQLService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	return &TextToSQLService{
		db:     db,
		openai: openai.NewClient(apiKey),
	}
}

// SQLQueryRequest represents a text-to-SQL query request
type SQLQueryRequest struct {
	Query string `json:"query"`
}

// SQLQueryResponse represents the text-to-SQL response
type SQLQueryResponse struct {
	SQL         string        `json:"sql"`
	Result      []interface{} `json:"result"`
	RowCount    int           `json:"row_count"`
	QueryType   string        `json:"query_type"`
	Explanation string        `json:"explanation"`
	Error       string        `json:"error,omitempty"`
}

// ConvertToSQL converts natural language to SQL and executes it
func (s *TextToSQLService) ConvertToSQL(query string) (*types.QueryResponse, error) {

	// Step 1: Generate SQL from natural language
	sqlQuery, queryType, explanation, err := s.generateSQL(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Step 2: Execute the SQL query
	results, rowCount, err := s.executeSQL(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL: %w", err)
	}

	sqlResponse := SQLQueryResponse{
		SQL:         sqlQuery,
		Result:      results,
		RowCount:    rowCount,
		QueryType:   queryType,
		Explanation: explanation,
	}

	return &types.QueryResponse{
		Success: true,
		Result:  sqlResponse,
		Query:   query,
		Time:    time.Now(),
	}, nil
}

// generateSQL uses OpenAI to convert natural language to SQL
func (s *TextToSQLService) generateSQL(query string) (string, string, string, error) {
	// Define the database schema for the AI
	schema := `
		Tables:
		
		sensor_readings (raw data):
		- time (TIMESTAMPTZ): When the reading was taken
		- device_id (TEXT): Unique device identifier
		- device_type (TEXT): Type of sensor (temperature_sensor, humidity_sensor, motion_detector, camera, controller)
		- location (TEXT): Location of the device (warehouse_a, warehouse_b, office_floor_1, parking_lot, server_room)
		- raw_value (NUMERIC): The sensor reading value
		- unit (TEXT): Unit of measurement (celsius, percent, boolean)
		- log_type (TEXT): Log level (INFO, WARNING, ERROR, CRITICAL, SECURITY)
		- message (TEXT): Human-readable log message

		five_min_sensor_averages (continuous aggregate - Level 1):
		- five_min_bucket (TIMESTAMPTZ): 5-minute bucket
		- device_type (TEXT): Type of sensor
		- location (TEXT): Location of device
		- avg_value (NUMERIC): Average reading for 5 minutes
		- min_value (NUMERIC): Minimum reading for 5 minutes
		- max_value (NUMERIC): Maximum reading for 5 minutes
		- reading_count (INTEGER): Number of readings in 5 minutes

		hourly_sensor_averages (continuous aggregate - Level 2):
		- hour (TIMESTAMPTZ): Hour bucket (built on five_min_sensor_averages)
		- device_type (TEXT): Type of sensor
		- location (TEXT): Location of device
		- avg_value (NUMERIC): Average of 5-min averages for the hour
		- min_value (NUMERIC): Minimum of 5-min minimums for the hour
		- max_value (NUMERIC): Maximum of 5-min maximums for the hour
		- reading_count (INTEGER): Sum of 5-min reading counts for the hour

		daily_sensor_averages (continuous aggregate - Level 3):
		- day (TIMESTAMPTZ): Day bucket (built on hourly_sensor_averages)
		- device_type (TEXT): Type of sensor
		- location (TEXT): Location of device
		- avg_value (NUMERIC): Average of hourly averages for the day
		- min_value (NUMERIC): Minimum of hourly minimums for the day
		- max_value (NUMERIC): Maximum of hourly maximums for the day
		- reading_count (INTEGER): Sum of hourly reading counts for the day

		daily_device_activity (continuous aggregate):
		- day (TIMESTAMPTZ): Day bucket
		- device_type (TEXT): Type of sensor
		- location (TEXT): Location of device
		- total_readings (INTEGER): Total readings for the day
		- error_count (INTEGER): Number of errors for the day
		- warning_count (INTEGER): Number of warnings for the day
		- info_count (INTEGER): Number of info logs for the day

		TimescaleDB Functions Available:
		- time_bucket(interval, time_column): Group by time intervals
		- NOW(): Current timestamp
		- INTERVAL: Time intervals like '1 hour', '24 hours', '7 days'
	`

	systemPrompt := fmt.Sprintf(`You are a SQL expert for a TimescaleDB database containing IoT sensor data with continuous aggregates for optimal performance. 
	
	Database Schema:
	%s
	
	Rules:
	1. PREFER hierarchical continuous aggregates for optimal performance:
	   - Use five_min_sensor_averages for real-time monitoring (5-min intervals)
	   - Use hourly_sensor_averages for hourly trends (built on 5-min data)
	   - Use daily_sensor_averages for daily summaries (built on hourly data)
	   - Use daily_device_activity for log analysis and error counts
	   - Only use sensor_readings for specific data or when aggregates don't fit
	
	2. Hierarchical Query Optimization Guidelines:
	   - For "recent 5-minute trends" → use five_min_sensor_averages (fastest)
	   - For "hourly averages" → use hourly_sensor_averages (reuses 5-min calculations)
	   - For "daily averages" → use daily_sensor_averages (reuses hourly calculations)
	   - For "daily error counts" → use daily_device_activity
	   - For "specific device readings" → use sensor_readings
	   
	3. Time-Series Query Rules:
	   - ALWAYS include time column (five_min_bucket, hour, day) for charting
	   - NEVER return just a single average without time buckets
	   - For "over last X hours" → use hourly_sensor_averages with time filter
	   - For "over last X days" → use daily_sensor_averages with time filter
	   - For "recent trends" → use five_min_sensor_averages
	
	4. Return only the SQL query, no explanations
	5. Use proper PostgreSQL syntax
	6. For time ranges, use NOW() - INTERVAL 'X hours/days'
	7. Always include ORDER BY time DESC for recent data
	8. Limit results to reasonable amounts (max 100 rows unless specifically asked for more)
	9. For filtering by temperature/humidity values, use raw_value column (sensor_readings) or avg_value (aggregates)
	10. For device filtering, use device_id or device_type columns
	11. For date filtering, use time::date = CURRENT_DATE for today
	
	Common query patterns:
	- "Show me temperature readings" → SELECT * FROM sensor_readings WHERE device_type = 'temperature_sensor' ORDER BY time DESC LIMIT 50
	- "Recent 5-minute trends" → SELECT five_min_bucket, avg_value, min_value, max_value FROM five_min_sensor_averages WHERE device_type = 'temperature_sensor' ORDER BY five_min_bucket DESC LIMIT 12
	- "Hourly averages" → SELECT hour, avg_value, min_value, max_value FROM hourly_sensor_averages WHERE device_type = 'temperature_sensor' ORDER BY hour DESC LIMIT 24
	- "Daily averages" → SELECT day, avg_value, min_value, max_value FROM daily_sensor_averages WHERE device_type = 'temperature_sensor' ORDER BY day DESC LIMIT 7
	- "Daily error summary" → SELECT day, device_type, location, error_count, warning_count FROM daily_device_activity ORDER BY day DESC LIMIT 7
	- "Today's readings" → SELECT * FROM sensor_readings WHERE time::date = CURRENT_DATE ORDER BY time DESC LIMIT 50
	
	IMPORTANT: For time-series queries like "average over last 24 hours", ALWAYS use time buckets:
	- "What's the average humidity over the last 24 hours?" → SELECT hour, avg_value FROM hourly_sensor_averages WHERE device_type = 'humidity_sensor' AND hour >= NOW() - INTERVAL '24 hours' ORDER BY hour DESC
	- "Average humidity over last 24 hours" → SELECT hour, avg_value FROM hourly_sensor_averages WHERE device_type = 'humidity_sensor' AND hour >= NOW() - INTERVAL '24 hours' ORDER BY hour DESC
	- "Temperature trends last week" → SELECT day, avg_value FROM daily_sensor_averages WHERE device_type = 'temperature_sensor' AND day >= NOW() - INTERVAL '7 days' ORDER BY day DESC
	- "Recent humidity data" → SELECT five_min_bucket, avg_value FROM five_min_sensor_averages WHERE device_type = 'humidity_sensor' AND five_min_bucket >= NOW() - INTERVAL '1 hour' ORDER BY five_min_bucket DESC
	`, schema)

	userPrompt := fmt.Sprintf("Convert this natural language query to SQL: %s", query)

	resp, err := s.openai.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: systemPrompt,
				},
				{
					Role:    "user",
					Content: userPrompt,
				},
			},
			Temperature: 0.1, // Low temperature for consistent SQL generation
		},
	)

	if err != nil {
		return "", "", "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", "", "", fmt.Errorf("no response from OpenAI")
	}

	sqlQuery := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Determine query type
	queryType := s.determineQueryType(sqlQuery)

	// Generate explanation
	explanation := s.generateExplanation(query, sqlQuery, queryType)

	return sqlQuery, queryType, explanation, nil
}

// executeSQL executes the generated SQL query
func (s *TextToSQLService) executeSQL(sqlQuery string) ([]interface{}, int, error) {
	// Log the SQL query and analyze which tables are being used
	s.logQueryAnalysis(sqlQuery)

	rows, err := s.db.Query(sqlQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("SQL execution error: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get columns: %w", err)
	}

	// Filter out embeddings and message columns
	var filteredColumns []string
	var columnIndexes []int
	for i, col := range columns {
		if col != "embedding" && col != "message" {
			filteredColumns = append(filteredColumns, col)
			columnIndexes = append(columnIndexes, i)
		}
	}

	var results []interface{}
	rowCount := 0

	for rows.Next() {
		// Create a slice to hold all the values (including filtered ones)
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Create a map for this row, excluding embeddings and message
		row := make(map[string]interface{})
		for i, colIndex := range columnIndexes {
			col := filteredColumns[i]
			val := values[colIndex]

			// Handle time.Time conversion
			if timestamp, ok := val.(time.Time); ok {
				row[col] = timestamp.Format(time.RFC3339)
			} else {
				// Handle numeric conversion for charting
				switch v := val.(type) {
				case []byte:
					// PostgreSQL numeric types come as []byte, convert to float
					if strVal := string(v); strVal != "" {
						if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
							row[col] = floatVal
						} else {
							row[col] = strVal
						}
					} else {
						row[col] = nil
					}
				case string:
					// Try to convert string to number if it looks numeric
					if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
						row[col] = floatVal
					} else {
						row[col] = v
					}
				default:
					row[col] = val
				}
			}
		}

		results = append(results, row)
		rowCount++
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating results: %w", err)
	}

	return results, rowCount, nil
}

// logQueryAnalysis analyzes and logs which tables are being queried
func (s *TextToSQLService) logQueryAnalysis(sqlQuery string) {
	queryLower := strings.ToLower(sqlQuery)

	// Define table categories
	rawTables := []string{"sensor_readings"}
	continuousAggregates := []string{
		"five_min_sensor_averages",
		"hourly_sensor_averages",
		"daily_sensor_averages",
		"daily_device_activity",
	}

	// Check which tables are being used
	var tablesUsed []string
	var tableTypes []string

	// Check for raw tables
	for _, table := range rawTables {
		if strings.Contains(queryLower, table) {
			tablesUsed = append(tablesUsed, table)
			tableTypes = append(tableTypes, "RAW_DATA")
		}
	}

	// Check for continuous aggregates
	for _, table := range continuousAggregates {
		if strings.Contains(queryLower, table) {
			tablesUsed = append(tablesUsed, table)
			tableTypes = append(tableTypes, "CONTINUOUS_AGGREGATE")
		}
	}

	// Determine query type
	queryType := "UNKNOWN"
	if strings.Contains(queryLower, "time_bucket") {
		queryType = "TIME_SERIES"
	} else if strings.Contains(queryLower, "avg(") || strings.Contains(queryLower, "count(") ||
		strings.Contains(queryLower, "sum(") || strings.Contains(queryLower, "min(") ||
		strings.Contains(queryLower, "max(") {
		queryType = "AGGREGATION"
	} else if strings.Contains(queryLower, "where") {
		queryType = "FILTERED_QUERY"
	} else {
		queryType = "SIMPLE_SELECT"
	}

	// Log the analysis
	log.Printf("🔍 QUERY ANALYSIS:")
	log.Printf("   SQL Query: %s", sqlQuery)
	log.Printf("   Query Type: %s", queryType)
	log.Printf("   Tables Used: %v", tablesUsed)
	log.Printf("   Table Types: %v", tableTypes)

	// Performance recommendations
	if len(tablesUsed) > 0 {
		hasRawData := false
		hasAggregates := false

		for _, tableType := range tableTypes {
			if tableType == "RAW_DATA" {
				hasRawData = true
			} else if tableType == "CONTINUOUS_AGGREGATE" {
				hasAggregates = true
			}
		}

		if hasRawData && !hasAggregates {
			log.Printf("   ⚠️  PERFORMANCE: Query uses raw data - consider using continuous aggregates for better performance")
		} else if hasAggregates {
			log.Printf("   ✅ PERFORMANCE: Query uses continuous aggregates - optimal performance")
		}
	}

	log.Printf("   ---")
}

// determineQueryType categorizes the SQL query
func (s *TextToSQLService) determineQueryType(sqlQuery string) string {
	sqlLower := strings.ToLower(sqlQuery)

	if strings.Contains(sqlLower, "avg(") || strings.Contains(sqlLower, "count(") ||
		strings.Contains(sqlLower, "sum(") || strings.Contains(sqlLower, "min(") ||
		strings.Contains(sqlLower, "max(") {
		return "aggregation"
	}

	if strings.Contains(sqlLower, "time_bucket") {
		return "time_series"
	}

	if strings.Contains(sqlLower, "where") && (strings.Contains(sqlLower, "error") ||
		strings.Contains(sqlLower, "critical") || strings.Contains(sqlLower, "warning")) {
		return "alert_filter"
	}

	return "data_query"
}

// generateExplanation provides a human-readable explanation
func (s *TextToSQLService) generateExplanation(query, sqlQuery, queryType string) string {
	switch queryType {
	case "aggregation":
		return fmt.Sprintf("This query calculates aggregated statistics from your sensor data based on: '%s'", query)
	case "time_series":
		return fmt.Sprintf("This query shows time-based trends and patterns from your sensor data based on: '%s'", query)
	case "alert_filter":
		return fmt.Sprintf("This query filters for alerts and issues in your sensor data based on: '%s'", query)
	default:
		return fmt.Sprintf("This query retrieves specific sensor data based on: '%s'", query)
	}
}
