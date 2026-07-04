package usecase

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// seedMids are approximate real-world mids for the currency pairs this
// simulator supports; each direction is seeded independently so a round
// trip (EUR->USD->EUR) isn't perfectly lossless, same as a real market.
var seedMids = map[string]float64{
	"EUR/USD": 1.08,
	"USD/EUR": 0.926,
	"GBP/USD": 1.27,
	"USD/GBP": 0.787,
	"EUR/GBP": 0.851,
	"GBP/EUR": 1.175,
}

// bucketWidth controls how often the simulated mid rate moves; within one
// bucket the rate is stable, so a quote and a same-instant re-quote agree.
const bucketWidth = 30 * time.Second

// MidRate returns the simulated mid-market rate for converting 1 unit of
// from into to, as of now. It is a deterministic pseudo-random walk around
// a seed value — the same (pair, time bucket) always yields the same rate,
// but the rate drifts bucket to bucket so a rates chart looks alive.
func MidRate(from, to string, now time.Time) (decimal.Decimal, error) {
	if from == to {
		return decimal.NewFromInt(1), nil
	}

	seed, ok := seedMids[from+"/"+to]
	if !ok {
		return decimal.Zero, fmt.Errorf("unsupported currency pair %s/%s", from, to)
	}

	bucket := now.UTC().Unix() / int64(bucketWidth.Seconds())

	h := fnv.New32a()
	_, _ = h.Write([]byte(from + to + strconv.FormatInt(bucket, 10)))
	// Map the hash to [-1, 1], then scale to ±0.3% of the seed mid.
	normalized := float64(h.Sum32()%2000)/1000 - 1
	offset := seed * 0.003 * normalized

	return decimal.NewFromFloat(seed + offset).Round(8), nil
}

// ApplyClientSpread widens the mid rate against the customer by spreadBps
// (hundredths of a percent), the way a real FX desk prices retail quotes.
func ApplyClientSpread(mid decimal.Decimal, spreadBps int) decimal.Decimal {
	factor := decimal.NewFromInt(10000 - int64(spreadBps)).Div(decimal.NewFromInt(10000))
	return mid.Mul(factor).Round(8)
}
