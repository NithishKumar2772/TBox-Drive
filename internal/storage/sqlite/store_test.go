package sqlite

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
)

func TestStorePersistsTransferAndParts(t *testing.T){
	db,err:=Open(filepath.Join(t.TempDir(),"tbox.db"));if err!=nil{t.Fatal(err)};defer db.Close()
	ctx:=context.Background()
	task:=transfer.Task{ID:"task-1",ProviderID:"terabox",Direction:"upload",Source:"/tmp/source.bin",Destination:"/cloud/source.bin",Size:100,Status:transfer.StatusQueued}
	if err:=db.CreateTransfer(ctx,task);err!=nil{t.Fatal(err)}
	if err:=db.UpdateProgress(ctx,"task-1",40,transfer.StatusRunning,"");err!=nil{t.Fatal(err)}
	got,err:=db.GetTransfer(ctx,"task-1");if err!=nil{t.Fatal(err)}
	if got.Transferred!=40||got.Status!=transfer.StatusRunning{t.Fatalf("got %#v",got)}
	part:=Part{TransferID:"task-1",Index:0,Offset:0,Size:40,Checksum:"abc",Status:"uploaded"}
	if err:=db.SavePart(ctx,part);err!=nil{t.Fatal(err)}
	parts,err:=db.ListParts(ctx,"task-1");if err!=nil{t.Fatal(err)}
	if len(parts)!=1||parts[0].Checksum!="abc"{t.Fatalf("got %#v",parts)}
}
func TestStoreRejectsProgressBeyondSize(t *testing.T){
	db,err:=Open(filepath.Join(t.TempDir(),"tbox.db"));if err!=nil{t.Fatal(err)};defer db.Close()
	ctx:=context.Background()
	if err:=db.CreateTransfer(ctx,transfer.Task{ID:"task-2",ProviderID:"terabox",Direction:"download",Size:10});err!=nil{t.Fatal(err)}
	if err:=db.UpdateProgress(ctx,"task-2",11,transfer.StatusRunning,"");err==nil{t.Fatal("expected progress validation error")}
}
func TestStoreListTransfers(t *testing.T){
	db,err:=Open(":memory:");if err!=nil{t.Fatal(err)};defer db.Close()
	ctx:=context.Background()
	for _,id:=range []string{"a","b"}{if err:=db.CreateTransfer(ctx,transfer.Task{ID:id,ProviderID:"terabox",Direction:"upload"});err!=nil{t.Fatal(err)}}
	items,err:=db.ListTransfers(ctx,10);if err!=nil{t.Fatal(err)}
	if len(items)!=2{t.Fatalf("want 2 transfers, got %d",len(items))}
}
