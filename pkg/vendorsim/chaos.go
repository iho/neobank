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

package vendorsim

import (
	"math/rand"
	"os"
	"strconv"
	"time"
)

// ChaosConfig tunes simulated delivery imperfections for integration tests.
// All knobs default to zero (off), so simulators are deterministic unless a
// test opts in.
type ChaosConfig struct {
	// MinDelay/MaxDelay bound the randomized delay before a webhook is first
	// attempted. A wide range naturally produces out-of-order delivery across
	// concurrently enqueued events.
	MinDelay time.Duration
	MaxDelay time.Duration
	// DuplicateProb is the probability [0,1] that a delivery is sent twice.
	DuplicateProb float64
	// ReorderProb is the probability [0,1] that an additional large jitter is
	// added on top of the base delay, letting a single delivery arrive
	// noticeably out of order even when MinDelay/MaxDelay are tight.
	ReorderProb float64
}

// ChaosConfigFromEnv reads "<prefix>_MIN_DELAY_MS", "<prefix>_MAX_DELAY_MS",
// "<prefix>_DUPLICATE_PROB", and "<prefix>_REORDER_PROB". Any unset or
// unparsable value defaults to off.
func ChaosConfigFromEnv(prefix string) ChaosConfig {
	return ChaosConfig{
		MinDelay:      envDurationMS(prefix+"_MIN_DELAY_MS", 0),
		MaxDelay:      envDurationMS(prefix+"_MAX_DELAY_MS", 0),
		DuplicateProb: envFloat(prefix+"_DUPLICATE_PROB", 0),
		ReorderProb:   envFloat(prefix+"_REORDER_PROB", 0),
	}
}

// Delay returns a randomized delivery delay within [MinDelay, MaxDelay],
// plus an occasional extra reorder jitter per ReorderProb.
func (c ChaosConfig) Delay() time.Duration {
	base := c.MinDelay
	if c.MaxDelay > c.MinDelay {
		span := c.MaxDelay - c.MinDelay
		base += time.Duration(rand.Int63n(int64(span)))
	}

	return base + c.reorderJitter()
}

func (c ChaosConfig) reorderJitter() time.Duration {
	if c.ReorderProb <= 0 || c.MaxDelay <= 0 || rand.Float64() >= c.ReorderProb {
		return 0
	}

	return time.Duration(rand.Int63n(int64(2 * c.MaxDelay)))
}

// ShouldDuplicate reports whether this delivery should be sent a second time.
func (c ChaosConfig) ShouldDuplicate() bool {
	return c.DuplicateProb > 0 && rand.Float64() < c.DuplicateProb
}

func envDurationMS(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	ms, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return time.Duration(ms) * time.Millisecond
}

func envFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}

	return f
}
