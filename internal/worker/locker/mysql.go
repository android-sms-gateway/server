package locker

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

type mySQLLocker struct {
	db *sql.DB

	prefix  string
	timeout time.Duration

	mu    sync.Mutex
	conns map[string]*sql.Conn
}

// NewMySQLLocker creates a new MySQL-based distributed locker.
func NewMySQLLocker(db *sql.DB, prefix string, timeout time.Duration) Locker {
	return &mySQLLocker{
		db: db,

		prefix:  prefix,
		timeout: timeout,

		mu:    sync.Mutex{},
		conns: make(map[string]*sql.Conn),
	}
}

// AcquireLock implements Locker.
func (m *mySQLLocker) AcquireLock(ctx context.Context, key string) error {
	name := m.prefix + key

	// Pin a dedicated connection for this lock.
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get conn: %w", err)
	}

	var res sql.NullInt64
	if lockErr := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", name, m.timeout.Seconds()).
		Scan(&res); lockErr != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to get lock: %w", lockErr)
	}
	if !res.Valid || res.Int64 != 1 {
		_ = conn.Close()
		return ErrLockNotAcquired
	}

	m.mu.Lock()
	// Should not exist; if it does, close previous to avoid leaks.
	if prev, ok := m.conns[key]; ok && prev != nil {
		_ = prev.Close()
	}
	m.conns[key] = conn
	m.mu.Unlock()

	return nil
}

// ReleaseLock implements Locker.
func (m *mySQLLocker) ReleaseLock(ctx context.Context, key string) error {
	name := m.prefix + key

	m.mu.Lock()
	conn := m.conns[key]
	delete(m.conns, key)
	m.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("%w: no held connection for key %q", ErrLockNotAcquired, key)
	}

	var result sql.NullInt64
	err := conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", name).Scan(&result)
	// Always close the pinned connection.
	_ = conn.Close()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if !result.Valid || result.Int64 != 1 {
		return fmt.Errorf("%w: lock was not held or doesn't exist", ErrLockNotAcquired)
	}

	return nil
}

// Close closes all remaining pinned connections.
// Should be called during shutdown to clean up resources.
func (m *mySQLLocker) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, conn := range m.conns {
		if conn != nil {
			_ = conn.Close()
		}
		delete(m.conns, key)
	}
	return nil
}

var _ Locker = (*mySQLLocker)(nil)
