//
// Copyright (c) 2026 Sumicare
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
