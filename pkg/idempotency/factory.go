package idempotency

import "log/slog"

// NewStoreFromEnv returns Redis store when redisURL is set, otherwise in-memory.
func NewStoreFromEnv(redisURL string, logger *slog.Logger) Store {
	if redisURL == "" {
		return NewMemoryStore()
	}
	store, err := NewRedisStore(redisURL)
	if err != nil {
		if logger != nil {
			logger.Warn("redis idempotency unavailable, using memory", "error", err)
		}
		return NewMemoryStore()
	}
	if logger != nil {
		logger.Info("idempotency using redis")
	}
	return store
}