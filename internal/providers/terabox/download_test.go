package terabox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadStreamsFromOfficialLink(t *testing.T){
	var server *httptest.Server
	server=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		switch r.URL.Path{
		case "/openapi/api/download":
			if r.URL.Query().Get("fidlist")!="[42]"||r.URL.Query().Get("type")!="dlink"{t.Errorf("unexpected query: %v",r.URL.Query())}
			fmt.Fprintf(w,`{"errno":0,"dlink":[{"fs_id":"42","dlink":%q}]}`,server.URL+"/file?fid=42")
		case "/file":
			if r.URL.Query().Get("access_tokens")!="token"{t.Error("access token missing from download URL")}
			w.Write([]byte("streamed file bytes"))
		default:http.NotFound(w,r)
		}
	}))
	defer server.Close()
	api,err:=newTestAPI(server)
	if err!=nil{t.Fatal(err)}
	body,size,err:=api.DownloadWithSize(context.Background(),"42");if err!=nil{t.Fatal(err)};defer body.Close()
	if size!=int64(len("streamed file bytes")){t.Fatalf("content length = %d, want %d",size,len("streamed file bytes"))}
	got,err:=io.ReadAll(body);if err!=nil{t.Fatal(err)}
	if string(got)!="streamed file bytes"{t.Fatalf("got %q",got)}
	legacyBody,err:=api.Download(context.Background(),"42");if err!=nil{t.Fatal(err)};defer legacyBody.Close()
	legacyData,err:=io.ReadAll(legacyBody);if err!=nil{t.Fatal(err)}
	if string(legacyData)!="streamed file bytes"{t.Fatalf("legacy Download returned %q",legacyData)}
}
func TestDownloadRejectsInvalidID(t *testing.T){
	api,err:=NewAPIClient(APIConfig{AccessToken:"token",TokenInfo:TokenInfo{APIDomain:"https://www.terabox.com",UploadDomain:"https://c-jp.terabox.com"},ApplicationRoot:"/app"});if err!=nil{t.Fatal(err)}
	if _,err:=api.DownloadURL(context.Background(),"../bad");err==nil{t.Fatal("expected invalid file ID")}
}
func TestDownloadRejectsUntrustedHost(t *testing.T){
	var server *httptest.Server
	server=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		fmt.Fprint(w,`{"errno":0,"dlink":[{"fs_id":"7","dlink":"https://evil.example/file"}]}`)
	}))
	defer server.Close()
	api,err:=newTestAPI(server);if err!=nil{t.Fatal(err)}
	if _,err:=api.DownloadURL(context.Background(),"7");err==nil{t.Fatal("expected untrusted host rejection")}
}



func TestDownloadRejectsLinkWithoutMatchingFileID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		fmt.Fprint(w,`{"errno":0,"dlink":[{"dlink":"https://d.terabox.com/file/test"}]}`)
	}))
	defer server.Close()
	api,err := newTestAPI(server)
	if err != nil { t.Fatal(err) }
	if _,err := api.DownloadURL(context.Background(),"42"); err == nil {
		t.Fatal("expected a download link without the requested fs_id to be rejected")
	}
}
