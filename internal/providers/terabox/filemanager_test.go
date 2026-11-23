package terabox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFileManagerRenameAndDeletePayloads(t *testing.T) {
	requests := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/openapi/api/filemanager" { t.Errorf("unexpected path %s", r.URL.Path) }
		if err := r.ParseForm(); err != nil { t.Error(err) }
		op := r.URL.Query().Get("opera")
		switch op {
		case "rename":
			var items []fileManagerItem
			if err := json.Unmarshal([]byte(r.Form.Get("filelist")), &items); err != nil { t.Error(err) }
			if len(items) != 1 || items[0].Path != "/From: Other Applications/TBox-Drive-123/old.txt" || items[0].NewName != "new.txt" { t.Errorf("bad rename payload: %#v", items) }
		case "delete":
			var paths []string
			if err := json.Unmarshal([]byte(r.Form.Get("filelist")), &paths); err != nil { t.Error(err) }
			if len(paths) != 1 || paths[0] != "/From: Other Applications/TBox-Drive-123/new.txt" { t.Errorf("bad delete payload: %#v", paths) }
		default:
			t.Errorf("unexpected operation %q", op)
		}
		fmt.Fprint(w, `{"errno":0,"info":[{"errno":0,"path":"/From: Other Applications/TBox-Drive-123/new.txt"}]}`)
	}))
	defer s.Close()
	api, err := newTestAPI(s)
	if err != nil { t.Fatal(err) }
	ctx := context.Background()
	if err := api.Rename(ctx, "/From: Other Applications/TBox-Drive-123/old.txt", "new.txt"); err != nil { t.Fatal(err) }
	if err := api.DeletePath(ctx, "/From: Other Applications/TBox-Drive-123/new.txt"); err != nil { t.Fatal(err) }
	if requests != 2 { t.Fatalf("want 2 requests, got %d", requests) }
}

func TestFileManagerRejectsUnsafePathsNamesAndRootOperations(t *testing.T) {
	api, err := NewAPIClient(APIConfig{AccessToken:"token", TokenInfo:TokenInfo{APIDomain:"https://api.terabox.com", UploadDomain:"https://upload.terabox.com"}, ApplicationRoot:"/From: Other Applications/TBox-Drive-123"})
	if err != nil { t.Fatal(err) }
	ctx := context.Background()
	for _, p := range []string{"/outside/secret.txt", "/From: Other Applications/TBox-Drive-123/../escape", "/From: Other Applications/TBox-Drive-123\\child"} {
		if err := api.Rename(ctx, p, "renamed.txt"); err == nil { t.Errorf("expected path rejection for %q", p) }
	}
	for _, p := range []string{"/From: Other Applications/TBox-Drive-123", "/From: Other Applications/TBox-Drive-123/", "//From: Other Applications//TBox-Drive-123//"} {
		if err := api.DeletePath(ctx, p); err == nil { t.Errorf("expected root delete rejection for %q", p) }
		if err := api.Rename(ctx, p, "renamed"); err == nil { t.Errorf("expected root rename rejection for %q", p) }
		if err := api.Move(ctx, p, "/From: Other Applications/TBox-Drive-123/sub", "moved"); err == nil { t.Errorf("expected root move rejection for %q", p) }
	}
	if err := api.Copy(ctx, "/From: Other Applications/TBox-Drive-123/file", "/From: Other Applications/TBox-Drive-123", "copy"); err != nil && err.Error() == "" { t.Fatal(err) }
	if err := api.Rename(ctx, "/From: Other Applications/TBox-Drive-123/file", "../bad"); err == nil { t.Fatal("expected unsafe name rejection") }
}

func TestFileManagerRejectsUnconfirmedAndPartialResults(t *testing.T) {
	for _, body := range []string{`{"errno":0}`, `{"errno":0,"info":[]}`, `{"errno":0,"info":[{"errno":0}]}`, `{"errno":0,"info":[{"path":"/app/file"}]}`, `{"errno":0,"info":[{"errno":0},{"errno":-2,"show_msg":"failed"}]}`} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
		api, err := newTestAPI(s)
		if err != nil { t.Fatal(err) }
		err = api.DeletePath(context.Background(), "/From: Other Applications/TBox-Drive-123/file")
		s.Close()
		if err == nil { t.Errorf("expected result validation for %s", body) }
	}
}

func newTestAPI(s *httptest.Server) (*APIClient,error) {
	cfg := APIConfig{AccessToken:"token", TokenInfo:TokenInfo{APIDomain:s.URL,UploadDomain:s.URL}, ApplicationRoot:"/From: Other Applications/TBox-Drive-123", HTTPClient:s.Client()}
	return newTestAPIWithLoopback(cfg)
}
