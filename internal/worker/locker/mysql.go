package locker

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type mySQLLocker struct {
	db *sql.DB

	prefix  string
	timeout uint

	mu    sync.Mutex
	conns map[string]*sql.Conn
}

func NewMySQLLocker(db *sql.DB, prefix string, timeout uint) Locker {
	return &mySQLLocker{
		db: db,

		prefix:  prefix,
		timeout: timeout,

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
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", name, m.timeout).Scan(&res); err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to get lock: %w", err)
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
		return fmt.Errorf("no held connection for key %q", key)
	}

	var result sql.NullInt64
	err := conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", name).Scan(&result)
	// Always close the pinned connection.
	_ = conn.Close()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if !result.Valid || result.Int64 != 1 {
		return fmt.Errorf("lock was not held or doesn't exist")
	}

	return nil
}

var _ Locker = (*mySQLLocker)(nil)
