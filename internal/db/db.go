package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/vibeswithkk/paylink/internal/config"
)

// DB holds database connections
type DB struct {
	SQL   *sql.DB
	Redis *RedisClient
}

// RedisClient is a minimal Redis client implementation
type RedisClient struct {
	Addr string
	// In production, use github.com/redis/go-redis/v9
	// For now, this is a stub that can be replaced
}

// Connect establishes database connections
func Connect(cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	redisClient := &RedisClient{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
	}

	return &DB{
		SQL:   sqlDB,
		Redis: redisClient,
	}, nil
}

// Close closes all database connections
func (db *DB) Close() {
	if db.SQL != nil {
		db.SQL.Close()
	}
}

// LPush adds a value to a Redis list (stub implementation)
func (r *RedisClient) LPush(ctx context.Context, key string, value []byte) error {
	// TODO: Replace with actual Redis client
	// For testing purposes, this is a no-op
	return nil
}

// BLPop blocks and pops from a Redis list (stub implementation)
func (r *RedisClient) BLPop(ctx context.Context, timeout time.Duration, key string) ([]string, error) {
	// TODO: Replace with actual Redis client
	// For now, simulate no messages
	time.Sleep(timeout)
	return nil, nil
}
