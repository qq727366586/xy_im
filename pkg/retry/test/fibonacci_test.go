package test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
	"xy_im/pkg/retry"
	"xy_im/pkg/retry/fibonacci"
	"xy_im/pkg/retry/retry_error"
)

func TestFibonacci(t *testing.T) {
	ts := time.Now()
	c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := retry.Do(c, retry.MaxRetry(4, fibonacci.NewFibonacci(1*time.Second)), func(ctx context.Context) error {
		fmt.Println(time.Since(ts))
		return retry_error.RetryableError(errors.New("123"))
	})
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
}
