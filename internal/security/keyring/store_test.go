package keyring

import (
	"context"
	"testing"
)

func TestNewRequiresServiceName(t *testing.T){
	if _,err:=New(" ");err==nil{t.Fatal("expected service name validation")}
	if _,err:=New("TBox Drive");err!=nil{t.Fatal(err)}
}
func TestStoreHonorsCanceledContext(t *testing.T){
	s,_:=New("TBox Drive");ctx,cancel:=context.WithCancel(context.Background());cancel()
	if _,err:=s.Get(ctx,"account");err==nil{t.Fatal("expected canceled context")}
	if err:=s.Set(ctx,"account",[]byte("secret"));err==nil{t.Fatal("expected canceled context")}
	if err:=s.Delete(ctx,"account");err==nil{t.Fatal("expected canceled context")}
}
