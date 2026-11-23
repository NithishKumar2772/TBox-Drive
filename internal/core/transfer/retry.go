package transfer

import("context";"errors";"math/rand";"time")
type RetryPolicy struct{MaxAttempts int;InitialDelay time.Duration;MaxDelay time.Duration;Retryable func(error)bool}
func(p RetryPolicy)normalized()RetryPolicy{if p.MaxAttempts<1{p.MaxAttempts=1};if p.InitialDelay<=0{p.InitialDelay=250*time.Millisecond};if p.MaxDelay<=0{p.MaxDelay=10*time.Second};if p.MaxDelay<p.InitialDelay{p.MaxDelay=p.InitialDelay};return p}
// Do retries only operations whose repetition is safe under verified TeraBox semantics.
func Do(ctx context.Context,policy RetryPolicy,operation func(context.Context)error)error{p:=policy.normalized();var last error;for attempt:=1;attempt<=p.MaxAttempts;attempt++{if err:=ctx.Err();err!=nil{return err};last=operation(ctx);if last==nil{return nil};if errors.Is(last,context.Canceled)||errors.Is(last,context.DeadlineExceeded){return last};if p.Retryable!=nil&&!p.Retryable(last){return last};if attempt==p.MaxAttempts{break};delay:=p.InitialDelay;for n:=1;n<attempt&&delay<p.MaxDelay/2;n++{delay*=2};if delay>p.MaxDelay{delay=p.MaxDelay};if delay>1{delay=time.Duration(rand.Int63n(int64(delay)+1))};timer:=time.NewTimer(delay);select{case<-ctx.Done():if !timer.Stop(){<-timer.C};return ctx.Err();case<-timer.C:}};return last}
