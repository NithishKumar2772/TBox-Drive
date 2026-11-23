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

func TestUploadFileResumesAfterInterruptedPart(t *testing.T) {
	var mu sync.Mutex
	precreateCalls := 0
	partCalls := map[int]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/openapi/api/precreate":
			mu.Lock()
			precreateCalls++
			call := precreateCalls
			mu.Unlock()
			if call == 1 {
				fmt.Fprint(w, `{"errno":0,"uploadid":"resume-session","return_type":1,"block_list":[0,1]}`)
			} else {
				fmt.Fprint(w, `{"errno":0,"uploadid":"resume-session","return_type":1,"block_list":[1]}`)
			}
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
			mu.Unlock()
			if part == 1 && call == 1 {
				http.Error(w, "temporary interruption", http.StatusBadGateway)
				return
			}
			fmt.Fprintf(w, `{"errno":0,"md5":%q,"uploadid":"resume-session","partseq":%d}`, digest, part)
		case "/openapi/api/create":
			fmt.Fprint(w, `{"errno":0,"fs_id":987,"path":"/From: Other Applications/TBox-Drive-123/resume.bin","server_filename":"resume.bin","size":10485760,"isdir":0}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	local := filepath.Join(t.TempDir(), "resume.bin")
	if err := os.WriteFile(local, make([]byte, 10<<20), 0600); err != nil { t.Fatal(err) }
	api, err := NewAPIClient(APIConfig{AccessToken:"test", TokenInfo:TokenInfo{APIDomain:server.URL, UploadDomain:server.URL}, ApplicationRoot:"/From: Other Applications/TBox-Drive-123"})
	if err != nil { t.Fatal(err) }
	store, err := sqlite.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer store.Close()

	options := UploadOptions{ChunkSize:8<<20, Repository:store, TransferID:"resume-task"}
	remote := "/From: Other Applications/TBox-Drive-123/resume.bin"
	if _, err := api.UploadFile(context.Background(), local, remote, options); err == nil {
		t.Fatal("expected first attempt to stop on interrupted part")
	}
	first, err := store.GetTransfer(context.Background(), "resume-task")
	if err != nil { t.Fatal(err) }
	if first.Status != transfer.StatusFailed || first.Transferred != 5<<20 {
		t.Fatalf("first attempt state: %#v", first)
	}
	firstParts, err := store.ListParts(context.Background(), "resume-task")
	if err != nil { t.Fatal(err) }
	if len(firstParts) != 1 || firstParts[0].Index != 0 {
		t.Fatalf("first attempt parts: %#v", firstParts)
	}

	result, err := api.UploadFile(context.Background(), local, remote, options)
	if err != nil { t.Fatal(err) }
	if result.TaskID != "resume-task" || result.File.ID != "987" { t.Fatalf("result: %#v", result) }
	final, err := store.GetTransfer(context.Background(), "resume-task")
	if err != nil { t.Fatal(err) }
	if final.Status != transfer.StatusCompleted || final.Transferred != 10<<20 {
		t.Fatalf("final state: %#v", final)
	}
	parts, err := store.ListParts(context.Background(), "resume-task")
	if err != nil { t.Fatal(err) }
	if len(parts) != 2 || parts[0].Status != "completed" || parts[1].Status != "completed" {
		t.Fatalf("final parts: %#v", parts)
	}
	mu.Lock()
	defer mu.Unlock()
	if precreateCalls != 2 || partCalls[0] != 1 || partCalls[1] != 2 {
		t.Fatalf("precreate calls=%d part calls=%v; expected to skip completed part 0 and retry missing part 1", precreateCalls, partCalls)
	}
}
