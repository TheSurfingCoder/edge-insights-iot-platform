package db

import (
	"database/sql"
	"edge-insights/internal/types"
	"time"
)

// LogEntry represents a log entry from the database
type LogEntry struct {
	Time     time.Time `json:"time"`
	DeviceID string    `json:"device_id"`
	LogType  string    `json:"log_type"`
	Message  string    `json:"message"`
}

// GetRecentLogs retrieves the most recent logs from the database
func GetRecentLogs(db *sql.DB, limit int) ([]LogEntry, error) {
	query := `
        SELECT time, device_id, log_type, message 
        FROM device_logs 
        ORDER BY time DESC 
        LIMIT $1
    `

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.Time, &log.DeviceID, &log.LogType, &log.Message); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetLogsByDevice retrieves logs for a specific device
func GetLogsByDevice(db *sql.DB, deviceID string, limit int) ([]LogEntry, error) {
	query := `
        SELECT time, device_id, log_type, message 
        FROM device_logs 
        WHERE device_id = $1
        ORDER BY time DESC 
        LIMIT $2
    `

	rows, err := db.Query(query, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.Time, &log.DeviceID, &log.LogType, &log.Message); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// Add new function for sensor readings
func StoreSensorReading(db *sql.DB, reading types.LogMessage) error {
	query := `
        INSERT INTO sensor_readings (time, device_id, device_type, location, raw_value, unit, log_type, message)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := db.Exec(query, reading.Time, reading.DeviceID, reading.DeviceType,
		reading.Location, reading.RawValue, reading.Unit, reading.LogType, reading.Message)
	return err
}

// Update GetRecentLogs to use new table
func GetRecentSensorReadings(db *sql.DB, limit int) ([]types.LogMessage, error) {
	query := `
        SELECT time, device_id, device_type, location, raw_value, unit, log_type, message 
        FROM sensor_readings 
        ORDER BY time DESC 
        LIMIT $1
    `

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []types.LogMessage
	for rows.Next() {
		var reading types.LogMessage
		if err := rows.Scan(&reading.Time, &reading.DeviceID, &reading.DeviceType,
			&reading.Location, &reading.RawValue, &reading.Unit, &reading.LogType, &reading.Message); err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	return readings, nil
}
