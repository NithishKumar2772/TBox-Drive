package transfer
import("context";"errors";"testing";"time")
func TestRetryEventuallySucceeds(t *testing.T){n:=0;err:=Do(context.Background(),RetryPolicy{MaxAttempts:3,InitialDelay:time.Nanosecond,MaxDelay:time.Nanosecond},func(context.Context)error{n++;if n<3{return errors.New("temporary")};return nil});if err!=nil{t.Fatal(err)};if n!=3{t.Fatalf("attempts=%d",n)}}
func TestRetryStopsOnPermanent(t *testing.T){n:=0;want:=errors.New("permanent");err:=Do(context.Background(),RetryPolicy{MaxAttempts:5,Retryable:func(error)bool{return false}},func(context.Context)error{n++;return want});if !errors.Is(err,want)||n!=1{t.Fatalf("err=%v n=%d",err,n)}}
func TestRetryCanceled(t *testing.T){ctx,cancel:=context.WithCancel(context.Background());cancel();err:=Do(ctx,RetryPolicy{MaxAttempts:3},func(context.Context)error{return errors.New("should not run")});if !errors.Is(err,context.Canceled){t.Fatalf("got %v",err)}}
