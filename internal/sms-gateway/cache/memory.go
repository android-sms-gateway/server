package cache

type memoryCache struct {
}

func NewMemory() Cache {
	return &memoryCache{}
}
