package terabox

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"sync/atomic"
)

func newTestAPIWithLoopback(cfg APIConfig) (*APIClient,error) {
	// Explicit loopback-only transport is permitted for hermetic tests.
	api,err:=newAPIClientWithTestDomains(cfg)
	return api,err
}

func TestAdapterListUsesConnectedClient(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w,`{"errno":0,"has_more":0,"info":[{"fs_id":77,"path":"/From: Other Applications/TBox-Drive-123/file.bin","server_filename":"file.bin","size":8,"isdir":0}]}`)
	}))
	defer s.Close()
	api,err:=newTestAPI(s)
	if err!=nil { t.Fatal(err) }
	a:=NewConnected(api)
	files,next,err:=a.List(context.Background(),"","")
	if err!=nil { t.Fatal(err) }
	if next!=""||len(files)!=1||files[0].ID!="77" { t.Fatalf("files=%#v next=%q",files,next) }
}

func TestAdapterListRejectsInvalidPageToken(t *testing.T) {
	a:=NewConnected(&APIClient{applicationRoot:"/root"})
	if _,_,err:=a.List(context.Background(),"/root","not-a-page");err==nil { t.Fatal("expected invalid page token") }
}

func TestUnconfiguredAdapterReturnsNotImplemented(t *testing.T) {
	if _,_,err:=New().List(context.Background(),"/","");err!=ErrNotImplemented { t.Fatalf("got %v",err) }
}

func TestNewAPIClientRejectsUntrustedServiceDomains(t *testing.T) {
	for _, domain := range []string{
		"https://evil.example",
		"https://terabox.com.evil.example",
		"https://terabox.com@evil.example",
		"http://api.terabox.com",
		"https://api.terabox.com:8443",
	} {
		_, err := NewAPIClient(APIConfig{
			AccessToken: "token",
			TokenInfo: TokenInfo{APIDomain: domain, UploadDomain: "https://upload.terabox.com"},
			ApplicationRoot: "/app",
		})
		if err == nil { t.Errorf("expected untrusted API domain rejection for %q", domain) }
	}
}

func TestNewAPIClientRejectsUntrustedUploadDomain(t *testing.T) {
	_, err := NewAPIClient(APIConfig{
		AccessToken: "token",
		TokenInfo: TokenInfo{APIDomain: "https://api.terabox.com", UploadDomain: "https://attacker.example"},
		ApplicationRoot: "/app",
	})
	if err == nil { t.Fatal("expected untrusted upload domain rejection") }
}

func TestAPIClientDoesNotFollowCredentialRedirect(t *testing.T) {
	var redirected atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirected.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer source.Close()
	api, err := newTestAPI(source)
	if err != nil { t.Fatal(err) }
	_, _, err = api.ListDirectory(context.Background(), "/From: Other Applications/TBox-Drive-123", 1, 10)
	if err == nil { t.Fatal("expected redirect response to be rejected") }
	if redirected.Load() { t.Fatal("credential-bearing request followed redirect to a different host") }
}
