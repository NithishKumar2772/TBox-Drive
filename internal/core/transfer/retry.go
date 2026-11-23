package transfer

import (
 "context"
 "errors"
 "math"
 "math/rand"
 "time"
)

type RetryPolicy struct {
 MaxAttempts int
 InitialDelay time.Duration
 MaxDelay time.Duration
 Multiplier float64
 Jitter float64
}
func(p RetryPolicy)Validate()error{
 if p.MaxAttempts<1{return errors.New("max attempts must be at least 1")}
 if p.InitialDelay<0||p.MaxDelay<0{return errors.New("retry delays cannot be negative")}
 if p.MaxDelay>0&&p.InitialDelay>p.MaxDelay{return errors.New("initial delay cannot exceed max delay")}
 if p.Multiplier<1{return errors.New("retry multiplier must be at least 1")}
 if p.Jitter<0||p.Jitter>1{return errors.New("jitter must be between 0 and 1")}
 return nil
}
// Delay computes the delay before retry n (first retry is n=1).
func(p RetryPolicy)Delay(n int,random *rand.Rand)time.Duration{
 if n<1||p.InitialDelay<=0{return 0}
 d:=float64(p.InitialDelay)*math.Pow(p.Multiplier,float64(n-1))
 if d>float64(math.MaxInt64){d=float64(math.MaxInt64)}
 if p.Jitter>0&&random!=nil{d*=1-p.Jitter+random.Float64()*2*p.Jitter}
 if p.MaxDelay>0&&d>float64(p.MaxDelay){d=float64(p.MaxDelay)}
 if d>float64(math.MaxInt64){return time.Duration(math.MaxInt64)}
 return time.Duration(d)
}
// Do retries only errors classified as transient. MaxAttempts includes the
// initial attempt; callers must ensure the operation is safe/idempotent.
func(p RetryPolicy)Do(ctx context.Context,operation func(context.Context)error,shouldRetry func(error)bool)error{
 if err:=p.Validate();err!=nil{return err}
 if operation==nil{return errors.New("retry operation is nil")}
 random:=rand.New(rand.NewSource(time.Now().UnixNano()))
 for attempt:=1;attempt<=p.MaxAttempts;attempt++{
  if err:=ctx.Err();err!=nil{return err}
  err:=operation(ctx);if err==nil{return nil}
  if attempt==p.MaxAttempts||shouldRetry==nil||!shouldRetry(err){return err}
  timer:=time.NewTimer(p.Delay(attempt,random))
  select{
  case <-ctx.Done():
   if !timer.Stop(){<-timer.C}
   return ctx.Err()
  case <-timer.C:
  }
 }
 return nil
}
