package terabox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFileManagerRenameAndDeletePayloads(t *testing.T){
	requests:=0
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		requests++
		if r.URL.Path!="/openapi/api/filemanager"{t.Errorf("unexpected path %s",r.URL.Path)}
		if err:=r.ParseForm();err!=nil{t.Error(err)}
		op:=r.URL.Query().Get("opera")
		switch op{
		case "rename":
			var items []fileManagerItem
			if err:=json.Unmarshal([]byte(r.Form.Get("filelist")),&items);err!=nil{t.Error(err)}
			if len(items)!=1||items[0].Path!="/From: Other Applications/TBox-Drive-123/old.txt"||items[0].NewName!="new.txt"{t.Errorf("bad rename payload: %#v",items)}
		case "delete":
			var paths []string
			if err:=json.Unmarshal([]byte(r.Form.Get("filelist")),&paths);err!=nil{t.Error(err)}
			if len(paths)!=1||paths[0]!="/From: Other Applications/TBox-Drive-123/new.txt"{t.Errorf("bad delete payload: %#v",paths)}
		default:t.Errorf("unexpected operation %q",op)
		}
		fmt.Fprint(w,`{"errno":0,"info":[{"errno":0,"path":"/From: Other Applications/TBox-Drive-123/new.txt"}]}`)
	}));defer s.Close()
	api,err:=NewAPIClient(APIConfig{AccessToken:"token",TokenInfo:TokenInfo{APIDomain:s.URL,UploadDomain:s.URL},ApplicationRoot:"/From: Other Applications/TBox-Drive-123"});if err!=nil{t.Fatal(err)}
	ctx:=context.Background()
	if err:=api.Rename(ctx,"/From: Other Applications/TBox-Drive-123/old.txt","new.txt");err!=nil{t.Fatal(err)}
	if err:=api.DeletePath(ctx,"/From: Other Applications/TBox-Drive-123/new.txt");err!=nil{t.Fatal(err)}
	if requests!=2{t.Fatalf("want 2 requests, got %d",requests)}
}
func TestFileManagerRejectsOutsideRootAndUnsafeName(t *testing.T){
	api,err:=NewAPIClient(APIConfig{AccessToken:"token",TokenInfo:TokenInfo{APIDomain:"https://www.terabox.com",UploadDomain:"https://c-jp.terabox.com"},ApplicationRoot:"/From: Other Applications/TBox-Drive-123"});if err!=nil{t.Fatal(err)}
	if err:=api.Rename(context.Background(),"/outside/secret.txt","renamed.txt");err==nil{t.Fatal("expected outside-root path rejection")}
	if err:=api.Rename(context.Background(),"/From: Other Applications/TBox-Drive-123/a.txt","../b.txt");err==nil{t.Fatal("expected unsafe name rejection")}
	if err:=api.DeletePath(context.Background(),"/From: Other Applications/TBox-Drive-123");err==nil{t.Fatal("expected application-root delete rejection")}
}
func TestFileManagerSurfacesPerItemErrors(t *testing.T){
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){fmt.Fprint(w,`{"errno":0,"info":[{"errno":-9,"path":"/missing.txt","show_msg":"not found"}]}`)}));defer s.Close()
	api,err:=NewAPIClient(APIConfig{AccessToken:"token",TokenInfo:TokenInfo{APIDomain:s.URL,UploadDomain:s.URL},ApplicationRoot:"/From: Other Applications/TBox-Drive-123"});if err!=nil{t.Fatal(err)}
	if err:=api.DeletePath(context.Background(),"/From: Other Applications/TBox-Drive-123/missing.txt");err==nil{t.Fatal("expected per-item error")}
}
