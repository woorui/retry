// package retry provides a wrapper to retry function.
package retry

import (
	"context"
	"math/rand"
	"time"
)

// Function defines how to handle data from Reader.
type Function interface {
	// Do this the function to retry.
	Do(context.Context) error
}

// The FunctionFunc is an adapter to allow the use of ordinary functions
// as Function. FunctionFunc(fn) is a Function that calls fn.
type FunctionFunc func(ctx context.Context) error

// Do calls FunctionFunc itself.
func (f FunctionFunc) Do(ctx context.Context) error { return f(ctx) }

// DefaultMaxRetries used when maxRetries passed less or equal 0.
const DefaultMaxRetries = 10

// DefaultRetryStrategy used when strategy passed is nil.
var DefaultRetryStrategy = FixedRetryStrategy(time.Second)

// Call calls function with retry,
// If fn returns an error, retry it.
//
// the defaults maxRetries is 10, maxRetries must large than 0, if not, set to default,
// the defaults strategy is 1 second.
func Call(ctx context.Context, fn Function, maxRetries int, strategy RetryStrategy) error {
	// call fn first, if return nil error, return directly.
	if err := fn.Do(ctx); err == nil {
		return nil
	}

	if maxRetries <= 0 {
		maxRetries = DefaultMaxRetries
	}
	if strategy == nil {
		strategy = FixedRetryStrategy(time.Second)
	}

	// goto retry !
	errch := make(chan error)

	go func() {
		var (
			i    = 1
			d    = strategy(i)
			tick = time.NewTicker(d)
		)
		var gerr error
		for {
			select {
			case <-ctx.Done():
				gerr = ctx.Err()
				goto END
			case <-tick.C:
				err := fn.Do(ctx)
				if err == nil {
					close(errch)
					return
				}
				gerr = err
				i++
				if i >= maxRetries {
					goto END
				}
				tick.Reset(strategy(i))
			}
		}
	END:
		errch <- gerr
		tick.Stop()
	}()

	err, ok := <-errch
	if !ok {
		return nil
	}
	return err
}

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
func FixedJitterRetryStrategy(d, maxJitter time.Duration) RetryStrategy {
	return func(_ int) time.Duration {
		return d + jitter(d, maxJitter)
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

		return time.Duration(i)*d + time.Duration(j)
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

	randSource := rand.NewSource(time.Now().UnixNano())

	d := time.Duration(rand.New(randSource).Int63n(max-min+1)+min) * time.Millisecond
	if d == 0 {
		return base
	}
	return d - base
}
