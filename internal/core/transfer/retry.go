package transfer

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryPolicy bounds attempts and delay growth. Jitter should be enabled for
// network clients to avoid synchronized retry bursts.
type RetryPolicy struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay time.Duration
	Multiplier float64
	Jitter float64 // 0..1 fraction of the computed delay
}

func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 1 { return errors.New("max attempts must be at least 1") }
	if p.InitialDelay < 0 || p.MaxDelay < 0 { return errors.New("retry delays cannot be negative") }
	if p.MaxDelay > 0 && p.InitialDelay > p.MaxDelay { return errors.New("initial delay cannot exceed max delay") }
	if p.Multiplier < 1 { return errors.New("retry multiplier must be at least 1") }
	if p.Jitter < 0 || p.Jitter > 1 { return errors.New("jitter must be between 0 and 1") }
	return nil
}

// Delay returns the delay before retry attempt n (where n=1 is the first retry).
// The optional random source makes jitter deterministic in tests.
func (p RetryPolicy) Delay(n int, random *rand.Rand) time.Duration {
	if n < 1 || p.InitialDelay <= 0 { return 0 }
	multiplier := math.Pow(p.Multiplier, float64(n-1))
	delay := float64(p.InitialDelay) * multiplier
	if p.MaxDelay > 0 && delay > float64(p.MaxDelay) { delay = float64(p.MaxDelay) }
	if p.Jitter > 0 && random != nil {
		factor := 1 - p.Jitter + random.Float64()*2*p.Jitter
		delay *= factor
	}
	if delay > float64(math.MaxInt64) { return time.Duration(math.MaxInt64) }
	return time.Duration(delay)
}

// Do retries only when shouldRetry says the error is transient. MaxAttempts
// includes the initial call. The operation must be safe to retry.
func (p RetryPolicy) Do(ctx context.Context, operation func(context.Context) error, shouldRetry func(error) bool) error {
	if err := p.Validate(); err != nil { return err }
	if operation == nil { return errors.New("retry operation is nil") }
	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil { return err }
		err := operation(ctx)
		if err == nil { return nil }
		if attempt == p.MaxAttempts || shouldRetry == nil || !shouldRetry(err) { return err }
		delay := p.Delay(attempt, nil)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() { <-timer.C }
			return ctx.Err()
		case <-timer.C:
		}
	}
	return nil
}
