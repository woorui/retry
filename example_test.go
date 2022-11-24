package retry_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/woorui/retry"
)

func Example() {
	errfn := func(ctx context.Context) error {
		return errors.New("errString")
	}

	err := retry.Call(
		context.Background(),
		retry.FunctionFunc(errfn),
		2,
		retry.FixedRetryStrategy(time.Second),
	)

	fmt.Println(err)

	// Output:
	// errString
}
