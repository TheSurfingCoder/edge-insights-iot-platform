package db

import (
	"database/sql"
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
