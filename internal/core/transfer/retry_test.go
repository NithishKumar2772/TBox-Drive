package transfer

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"
)

func TestRetryDelayCaps(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 5, InitialDelay: time.Second, MaxDelay: 3*time.Second, Multiplier: 2}
	if got := p.Delay(1,nil); got != time.Second { t.Fatalf("got %v",got) }
	if got := p.Delay(3,nil); got != 3*time.Second { t.Fatalf("got %v",got) }
}

func TestRetryDelayJitterIsBounded(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 4, InitialDelay: time.Second, MaxDelay: 10*time.Second, Multiplier: 2, Jitter: .25}
	got := p.Delay(1,rand.New(rand.NewSource(1)))
	if got < 750*time.Millisecond || got > 1250*time.Millisecond { t.Fatalf("jitter out of bounds: %v",got) }
}

func TestRetryDoRetriesTransientError(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 3, InitialDelay: time.Nanosecond, MaxDelay: time.Millisecond, Multiplier: 2}
	calls := 0
	err := p.Do(context.Background(),func(context.Context)error{calls++;if calls<3{return errors.New("temporary")};return nil},func(error)bool{return true})
	if err != nil { t.Fatal(err) }
	if calls != 3 { t.Fatalf("want 3 calls, got %d",calls) }
}

func TestRetryDoDoesNotRetryPermanentError(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 5, InitialDelay: time.Nanosecond, MaxDelay: time.Millisecond, Multiplier: 2}
	calls := 0
	err := p.Do(context.Background(),func(context.Context)error{calls++;return errors.New("unauthorized")},func(error)bool{return false})
	if err == nil || calls != 1 { t.Fatalf("expected one permanent failure; calls=%d err=%v",calls,err) }
}

func TestRetryDoHonorsCanceledContext(t *testing.T) {
	ctx,cancel := context.WithCancel(context.Background()); cancel()
	p := RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaxDelay: time.Second, Multiplier: 2}
	err := p.Do(ctx,func(context.Context)error{t.Fatal("must not run");return nil},func(error)bool{return true})
	if !errors.Is(err,context.Canceled) { t.Fatalf("got %v",err) }
}
