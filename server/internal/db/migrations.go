// automates the maintenance of my schema beacuse we run this runner each time
// during intial set up of my application or when i make changes to my schema to keep my database up to date with my latest changes
// makes managing my schema a lot smoother over time

package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// List of migration files in order
	migrations := []string{
		"migrations/001_create_device_logs_table.sql",
		"migrations/002_create_embeddings_table.sql",
		"migrations/003_create_sensor_readings_table.sql",
		"migrations/005_add_log_type_to_sensor_readings.sql",
		"migrations/008_add_message_to_sensor_readings.sql",
	}

	for _, migrationPath := range migrations {
		log.Printf("Running migration: %s", migrationPath)

		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
		}

		// Split by semicolon and execute each statement
		statements := strings.Split(string(content), ";")

		for _, statement := range statements {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}

			if _, err := db.Exec(statement); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migrationPath, err)
			}
		}

		log.Printf("Migration %s completed", migrationPath)
	}

	log.Println("All database migrations completed successfully")
	return nil
}
