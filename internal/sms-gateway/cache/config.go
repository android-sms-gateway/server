package cache

// Config controls the cache backend via a URL (e.g., "memory://", "redis://...").
type Config struct {
	URL string
}
