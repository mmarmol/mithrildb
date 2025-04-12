package db

// ReadTier values used for SetReadTier()
const (
	// Read from all tiers: memory + disk
	ReadTierAll = 0

	// Read only from cache (no disk access)
	ReadTierCacheOnly = 1
)
