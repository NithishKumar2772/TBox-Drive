package transfer
import("context";"errors";"sync/atomic";"testing")
func TestRunWorkers(t *testing.T){var n atomic.Int32;jobs:=[]int{1,2,3,4,5};if err:=RunWorkers(context.Background(),jobs,2,func(context.Context,int)error{n.Add(1);return nil});err!=nil{t.Fatal(err)};if n.Load()!=5{t.Fatalf("processed %d",n.Load())}}
func TestRunWorkersError(t *testing.T){want:=errors.New("part failed");err:=RunWorkers(context.Background(),[]int{1,2},1,func(_ context.Context,n int)error{if n==1{return want};return nil});if !errors.Is(err,want){t.Fatalf("got %v",err)}}
func TestRunWorkersZero(t *testing.T){if err:=RunWorkers(context.Background(),[]int{1},0,func(context.Context,int)error{return nil});err==nil{t.Fatal("expected error")}}
