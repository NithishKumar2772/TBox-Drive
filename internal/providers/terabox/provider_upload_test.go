package terabox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
)

func TestAdapterUploadUsesCallerTransferID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openapi/api/precreate" { http.NotFound(w,r); return }
		fmt.Fprint(w, `{"errno":0,"return_type":2,"info":{"fs_id":55,"path":"/root/file.txt","server_filename":"file.txt","size":1,"isdir":0}}`)
	}))
	defer server.Close()
	local:=filepath.Join(t.TempDir(),"file.txt")
	if err:=os.WriteFile(local,[]byte("x"),0600);err!=nil{t.Fatal(err)}
	api, err := newTestAPI(&httptest.Server{URL: server.URL, Client: server.Client()})
	if err!=nil{t.Fatal(err)}
	store,err:=sqlite.Open(":memory:");if err!=nil{t.Fatal(err)};defer store.Close()
	adapter:=NewConnectedWithRepository(api,store)
	result,err:=adapter.Upload(context.Background(),provider.UploadRequest{LocalPath:local,ParentID:"/root",Name:"file.txt",TaskID:"caller-task"})
	if err!=nil{t.Fatal(err)}
	if result.TaskID!="caller-task"{t.Fatalf("got task ID %q",result.TaskID)}
	task,err:=store.GetTransfer(context.Background(),"caller-task");if err!=nil{t.Fatal(err)}
	if task.Status!=transfer.StatusCompleted||task.Destination!="/root/file.txt"{t.Fatalf("unexpected task: %#v",task)}
}
