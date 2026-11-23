package transfer

import "testing"

func TestPlanChunks(t *testing.T){
 got,err:=PlanChunks(10,4);if err!=nil{t.Fatal(err)}
 if len(got)!=3{t.Fatalf("want 3 chunks, got %d",len(got))}
 if got[0].Offset!=0||got[0].Size!=4||got[2].Offset!=8||got[2].Size!=2{t.Fatalf("unexpected chunks: %#v",got)}
}
func TestPlanChunksEmpty(t *testing.T){
 got,err:=PlanChunks(0,4);if err!=nil{t.Fatal(err)}
 if len(got)!=0{t.Fatalf("want zero chunks, got %d",len(got))}
}
func TestPlanChunksInvalid(t *testing.T){
 if _,err:=PlanChunks(-1,4);err==nil{t.Fatal("expected negative-size error")}
 if _,err:=PlanChunks(10,0);err==nil{t.Fatal("expected invalid chunk-size error")}
}
func TestPlanChunksRejectsHugePlan(t *testing.T){
 if _,err:=PlanChunks(1<<63-1,1);err==nil{t.Fatal("expected huge plan to be rejected")}
}
