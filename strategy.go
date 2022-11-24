package retry

import (
	"math/rand"
	"time"
)

// RetryStrategy is retry strategy.
//
// we provide six RetryStrategy below:
// 1. FixedRetryStrategy,
// 2. FixedJitterRetryStrategy,
// 3. LinearRetryStrategy,
// 4. LinearJitterRetryStrategy,
// 5. ExponentialRetryStrategy,
// 6. ExponentialJitterRetryStrategy,
type RetryStrategy func(i int) time.Duration

// FixedRetryStrategy returns fixed durations.
func FixedRetryStrategy(d time.Duration) RetryStrategy {
	return func(_ int) time.Duration {
		return d
	}
}

// FixedJitterRetryStrategy returns fixed durations with jitter.
//
// maxJitter defines the max jitter, maxJitter must be less than or equal to d,
// the returned durations range is between `d - jitter` and `d + jitter`.
func FixedJitterRetryStrategy(d time.Duration) RetryStrategy {
	return func(_ int) time.Duration {
		return d
	}
}

// LinearRetryStrategy returns increasing durations, each duration longer than the last.
func LinearRetryStrategy(d time.Duration) RetryStrategy {
	return func(i int) time.Duration {
		return time.Duration(i) * d
	}
}

// LinearJitterRetryStrategy returns increasing durations with a jitter, each duration longer than the last.
//
// maxJitter defines the max jitter, maxJitter must be less than or equal to d,
// the returned durations range is between `d - jitter` and `d + jitter`.
func LinearJitterRetryStrategy(d, maxJitter time.Duration) RetryStrategy {
	return func(i int) time.Duration {
		j := jitter(d, maxJitter)

		return time.Duration(i)*d - time.Duration(j)
	}
}

// ExponentialRetryStrategy returns increasing durations, each duration is a power of 2 than last.
func ExponentialRetryStrategy(d time.Duration) RetryStrategy {
	return func(i int) time.Duration {
		return time.Duration(1<<int64(i)) * d
	}
}

// ExponentialJitterRetryStrategy returns increasing durations with a jitter, each duration is a power of 2 than last.
//
// maxJitter defines the max jitter, maxJitter must be less than or equal to d,
// the returned durations range is between `d - jitter` and `d + jitter`.
func ExponentialJitterRetryStrategy(d, maxJitter time.Duration) RetryStrategy {
	return func(i int) time.Duration {
		j := jitter(d, maxJitter)

		return time.Duration(1<<int64(i))*d + j
	}
}

func jitter(base, maxJitter time.Duration) time.Duration {
	if maxJitter > base {
		panic("maxJitter must be less than or equal to base")
	}

	var (
		baseInt64   = int64(base / time.Millisecond)
		jitterInt64 = int64(maxJitter / time.Millisecond)
	)

	var (
		min = baseInt64 - jitterInt64
		max = baseInt64 + jitterInt64
	)

	d := time.Duration(rand.Int63n(max-min+1)+min) * time.Millisecond
	if d == 0 {
		return base
	}
	return d
}
