package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teerarak/mcp-postgresql-report/internal/config"
)

// Manager manages a pgxpool connection pool.
type Manager struct {
	mu   sync.Mutex
	pool *pgxpool.Pool
}

// NewManager creates a new Manager with no active connection.
func NewManager() *Manager {
	return &Manager{}
}

// Connect initializes or replaces the connection pool using the given config.
func (m *Manager) Connect(ctx context.Context, cfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pool != nil {
		m.pool.Close()
		m.pool = nil
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("ping failed: %w", err)
	}

	m.pool = pool
	return nil
}

// Pool returns the active connection pool, or an error if not connected.
func (m *Manager) Pool() (*pgxpool.Pool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pool == nil {
		return nil, fmt.Errorf("not connected to database")
	}
	return m.pool, nil
}

// EnsureConnected auto-connects from env vars if no pool exists.
func (m *Manager) EnsureConnected(ctx context.Context) error {
	m.mu.Lock()
	hasPool := m.pool != nil
	m.mu.Unlock()

	if hasPool {
		return nil
	}

	cfg := config.Load()
	if cfg.DBName == "" {
		return fmt.Errorf("not connected: call connect_database first or set POSTGRES_DB env var")
	}
	return m.Connect(ctx, cfg)
}

// Close closes the underlying connection pool.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pool != nil {
		m.pool.Close()
		m.pool = nil
	}
}
