package transfer
import("io";"os";"path/filepath";"testing")
func TestPlanChunks(t *testing.T){got,err:=PlanChunks(10,4);if err!=nil{t.Fatal(err)};want:=[]Chunk{{0,0,4},{1,4,4},{2,8,2}};if len(got)!=len(want){t.Fatalf("got %#v",got)};for i:=range want{if got[i]!=want[i]{t.Fatalf("got %#v want %#v",got[i],want[i])}}}
func TestPlanChunksEmpty(t *testing.T){got,err:=PlanChunks(0,4);if err!=nil{t.Fatal(err)};if len(got)!=0{t.Fatalf("got %#v",got)}}
func TestPlanChunksRejectsInvalid(t *testing.T){for _,v:=range []struct{s,c int64;m int}{{-1,4,10},{10,0,10},{10,4,-1},{100,1,10},{int64(^uint64(0)>>1),1,10}}{if _,err:=PlanChunksLimited(v.s,v.c,v.m);err==nil{t.Fatalf("expected error %+v",v)}}}
func TestOpenChunk(t *testing.T){p:=filepath.Join(t.TempDir(),"f");if err:=os.WriteFile(p,[]byte("abcdefghij"),0600);err!=nil{t.Fatal(err)};r,err:=OpenChunk(p,Chunk{1,3,4});if err!=nil{t.Fatal(err)};defer r.Close();b,err:=io.ReadAll(r);if err!=nil{t.Fatal(err)};if string(b)!="defg"{t.Fatalf("got %q",b)};if _,err:=OpenChunk(p,Chunk{0,9,2});err==nil{t.Fatal("expected out of range")}}
