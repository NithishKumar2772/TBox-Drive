package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

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

func TestCredentialLockCoordinatesSeparateProcesses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shared.db")
	first, err := Open(path)
	if err != nil { t.Fatal(err) }
	defer first.Close()
	second, err := Open(path)
	if err != nil { t.Fatal(err) }
	defer second.Close()

	releaseFirst, err := first.AcquireCredentialLock(context.Background(), "terabox:42", time.Minute)
	if err != nil { t.Fatal(err) }
	waitCtx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	if _, err := second.AcquireCredentialLock(waitCtx, "terabox:42", time.Minute); err == nil {
		t.Fatal("second process acquired a lock that is still held")
	} else if err != context.DeadlineExceeded {
		t.Fatalf("expected deadline while waiting for lock, got %v", err)
	}
	if err := releaseFirst(); err != nil { t.Fatal(err) }

	releaseSecond, err := second.AcquireCredentialLock(context.Background(), "terabox:42", time.Minute)
	if err != nil { t.Fatalf("second process could not acquire released lock: %v", err) }
	if err := releaseSecond(); err != nil { t.Fatal(err) }
}

func TestCredentialLockReclaimsExpiredLease(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer db.Close()
	release, err := db.AcquireCredentialLock(context.Background(), "terabox:99", time.Minute)
	if err != nil { t.Fatal(err) }
	if _, err := db.db.Exec(`UPDATE credential_locks SET expires_at_ns=0 WHERE lock_key=?`, "terabox:99"); err != nil { t.Fatal(err) }
	nextRelease, err := db.AcquireCredentialLock(context.Background(), "terabox:99", time.Minute)
	if err != nil { t.Fatalf("expired credential lock was not reclaimed: %v", err) }
	_ = release()
	if err := nextRelease(); err != nil { t.Fatal(err) }
}

func TestCredentialLockRequiresKey(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer db.Close()
	if _, err := db.AcquireCredentialLock(context.Background(), " ", time.Minute); err == nil {
		t.Fatal("expected empty credential lock key to be rejected")
	}
}
