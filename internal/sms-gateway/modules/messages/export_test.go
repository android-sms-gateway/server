package messages

// NewStateCacheForTest exposes newCache to the external test package.
var NewStateCacheForTest = newCache

// NewCountCacheForTest exposes newCountCache to the external test package.
var NewCountCacheForTest = newCountCache

// StateCacheKeyForTest exposes the state cache key layout to the external test package.
var StateCacheKeyForTest = stateCacheKey

// CountCacheKeyVersionForTest exposes the list-total cache key namespace to the external test package.
const CountCacheKeyVersionForTest = countCacheKeyVersion

// CountCacheKeyForTest exposes the storage key Count uses for this filter to the external test package.
func (f *SelectFilter) CountCacheKeyForTest() string {
	return countCacheKey(f.hash())
}
