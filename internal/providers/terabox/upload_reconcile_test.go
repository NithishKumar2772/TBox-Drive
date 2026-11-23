package terabox

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
)

func TestUploadFileReconcilesAmbiguousShardFailure(t *testing.T) {
	var mu sync.Mutex
	precreateCalls := 0
	partCalls := map[int]int{}
	stored := map[int]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/openapi/api/precreate":
			mu.Lock()
			precreateCalls++
			missing := []int{}
			for i := 0; i < 2; i++ { if !stored[i] { missing = append(missing, i) } }
			mu.Unlock()
			fmt.Fprintf(w, `{"errno":0,"uploadid":"reconcile-session","return_type":1,"block_list":%v}`, func() string {
				if len(missing) == 1 { return fmt.Sprintf("[%d]", missing[0]) }
				if len(missing) == 2 { return "[0,1]" }
				return "[]"
			}())
		case "/rest/2.0/pcs/superfile2":
			part, _ := strconv.Atoi(r.URL.Query().Get("partseq"))
			f, _, err := r.FormFile("file")
			if err != nil { t.Error(err); http.Error(w, "bad multipart", 400); return }
			data, err := io.ReadAll(f); f.Close()
			if err != nil { t.Error(err); http.Error(w, "read failed", 500); return }
			sum := md5.Sum(data)
			digest := hex.EncodeToString(sum[:])
			mu.Lock()
			partCalls[part]++
			call := partCalls[part]
			stored[part] = true
			mu.Unlock()
			if part == 0 && call == 1 {
				http.Error(w, "response lost after remote acceptance", http.StatusBadGateway)
				return
			}
			fmt.Fprintf(w, `{"errno":0,"md5":%q,"uploadid":"reconcile-session","partseq":%d}`, digest, part)
		case "/openapi/api/create":
			fmt.Fprint(w, `{"errno":0,"fs_id":988,"path":"/From: Other Applications/TBox-Drive-123/reconciled.bin","server_filename":"reconciled.bin","size":10485760,"isdir":0}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	local := filepath.Join(t.TempDir(), "reconciled.bin")
	data := make([]byte, 10<<20)
	for i := range data { data[i] = byte(i % 251) }
	if err := os.WriteFile(local, data, 0600); err != nil { t.Fatal(err) }
	api, err := NewAPIClient(APIConfig{AccessToken:"test", TokenInfo:TokenInfo{APIDomain:server.URL, UploadDomain:server.URL}, ApplicationRoot:"/From: Other Applications/TBox-Drive-123"})
	if err != nil { t.Fatal(err) }
	store, err := sqlite.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer store.Close()

	result, err := api.UploadFile(context.Background(), local, "/From: Other Applications/TBox-Drive-123/reconciled.bin", UploadOptions{ChunkSize:8<<20, Repository:store, TransferID:"reconcile-task"})
	if err != nil { t.Fatal(err) }
	if result.File.ID != "988" { t.Fatalf("unexpected result: %#v", result) }
	task, err := store.GetTransfer(context.Background(), "reconcile-task")
	if err != nil { t.Fatal(err) }
	if task.Status != transfer.StatusCompleted || task.Transferred != 10<<20 { t.Fatalf("task: %#v", task) }
	mu.Lock()
	defer mu.Unlock()
	if precreateCalls != 2 || partCalls[0] != 1 || partCalls[1] != 1 {
		t.Fatalf("precreate=%d part calls=%v; expected part 0 to be reconciled without re-upload", precreateCalls, partCalls)
	}
}
