package pubsub

// Config controls the PubSub backend via a URL (e.g., "memory://", "redis://...").
type Config struct {
	URL        string
	BufferSize uint
}
