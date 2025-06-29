

package db

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
    Host     string
    Port     string
    Database string
    User     string
    Password string
    SSLMode  string
}

func LoadConfig() *Config {
    return &Config{
        Host:     getEnv("TIMESCALE_HOST", "localhost"),
        Port:     getEnv("TIMESCALE_PORT", "5432"),
        Database: getEnv("TIMESCALE_DB", "postgres"),
        User:     getEnv("TIMESCALE_USER", "postgres"),
        Password: getEnv("TIMESCALE_PASSWORD", ""),
        SSLMode:  getEnv("TIMESCALE_SSL_MODE", "require"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func Connect(config *Config) (*sql.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
        config.Host, config.Port, config.Database, config.User, config.Password, config.SSLMode)
    
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Test the connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Println("Successfully connected to TimescaleDB")
    return db, nil
}