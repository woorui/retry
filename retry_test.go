package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCall(t *testing.T) {
	type args struct {
		fn         Function
		maxRetries int
		strategy   RetryStrategy
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "retry fail",
			args: args{
				fn:         &testFunc{ok: 6},
				maxRetries: 5,
				strategy:   FixedRetryStrategy(time.Millisecond),
			},
			wantErr: errTest,
		},
		{
			name: "retry success",
			args: args{
				fn:         &testFunc{ok: 3},
				maxRetries: 5,
				strategy:   FixedRetryStrategy(time.Millisecond),
			},
			wantErr: nil,
		},
		{
			name: "retry in default args success",
			args: args{
				fn:         &testFunc{ok: 3},
				maxRetries: 0,
				strategy:   nil,
			},
			wantErr: nil,
		},
		{
			name: "retry in default args fail",
			args: args{
				fn:         &testFunc{ok: 300},
				maxRetries: 0,
				strategy:   FixedRetryStrategy(time.Millisecond),
			},
			wantErr: errTest,
		},
		{
			name: "retry in first time success",
			args: args{
				fn:         &testFunc{ok: 1},
				maxRetries: 1,
				strategy:   FixedRetryStrategy(time.Millisecond),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Call(
				context.Background(),
				tt.args.fn,
				tt.args.maxRetries,
				tt.args.strategy,
			); err != tt.wantErr {
				t.Errorf("Call() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStrategy(t *testing.T) {
	var (
		base   = 2 * time.Second
		jitter = time.Second
		n      = 2
	)
	t.Run("FixedRetryStrategy", func(t *testing.T) {
		var (
			got  = FixedRetryStrategy(base)(n)
			want = base
		)
		if got != want {
			t.Errorf("FixedRetryStrategy got %v, want %v", got, want)
		}
	})
	t.Run("FixedJitterRetryStrategy", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			var (
				got   = FixedJitterRetryStrategy(base, jitter)(n)
				left  = base - jitter
				right = base + jitter
			)
			if got <= left || got > right {
				t.Errorf("FixedRetryStrategy got %v, want between %v and %v", got, left, right)
			}
		}
	})
	t.Run("LinearRetryStrategy", func(t *testing.T) {
		var (
			got  = LinearRetryStrategy(base)(n)
			want = base * time.Duration(n)
		)
		if got != want {
			t.Errorf("LinearRetryStrategy got %v, want %v", got, want)
		}
	})
	t.Run("LinearJitterRetryStrategy", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			var (
				got   = LinearJitterRetryStrategy(base, jitter)(n)
				left  = base*time.Duration(n) - jitter
				right = base*time.Duration(n) + jitter
			)
			if got <= left || got > right {
				t.Errorf("LinearJitterRetryStrategy got %v, want between %v and %v", got, left, right)
			}
		}
	})
	t.Run("ExponentialRetryStrategy", func(t *testing.T) {
		var (
			got  = ExponentialRetryStrategy(base)(n)
			want = base * time.Duration(1<<int64(n))
		)
		if got != want {
			t.Errorf("ExponentialRetryStrategy got %v, want %v", got, want)
		}
	})
	t.Run("ExponentialJitterRetryStrategy", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			var (
				got   = ExponentialJitterRetryStrategy(base, jitter)(n)
				left  = base*time.Duration(1<<int64(n)) - jitter
				right = base*time.Duration(1<<int64(n)) + jitter
			)
			if got < left || got > right {
				t.Errorf("ExponentialJitterRetryStrategy got %v, want between %v and %v", got, left, right)
			}
		}
	})
}

var errTest = errors.New("test error")

// testFunc implements Function interface,
// it returns error nil after call Do() ok times.
type testFunc struct {
	mu sync.Mutex
	n  int
	ok int
}

func (f *testFunc) Do(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.n += 1
	if f.n >= f.ok {
		return nil
	}
	return errTest
}
