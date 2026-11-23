package web

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NithishKumar2772/TBox-Drive/internal/app"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/account"
	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
)

type testSecrets struct{}
func (testSecrets) Get(context.Context,string)([]byte,error){ return nil, nil }
func (testSecrets) Set(context.Context,string,[]byte)error{return nil}
func (testSecrets) Delete(context.Context,string)error{return nil}

func testServer(t *testing.T) *Server {
	t.Helper()
	store,err:=sqlite.Open(":memory:")
	if err!=nil{t.Fatal(err)}
	t.Cleanup(func(){_ = store.Close()})
	oauth,err:=terabox.NewOAuthClient(terabox.OAuthConfig{ClientID:"test",ClientSecret:"test",PrivateSecret:"test"})
	if err!=nil{t.Fatal(err)}
	service,err:=app.NewTeraBoxService(oauth,testSecrets{},store,store,"/Apps/TBox-Drive")
	if err!=nil{t.Fatal(err)}
	server,err:=NewServer(service,store,"/Apps/TBox-Drive")
	if err!=nil{t.Fatal(err)}
	return server
}

func TestLoopbackStatusIsAvailable(t *testing.T){
	server:=testServer(t)
	req:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:8765/api/status",nil)
	req.Host="127.0.0.1:8765"
	req.AddCookie(&http.Cookie{Name:"tbox_local_session",Value:server.sessionToken})
	response:=httptest.NewRecorder()
	server.Handler().ServeHTTP(response,req)
	if response.Code!=http.StatusOK{t.Fatalf("status = %d, body=%s",response.Code,response.Body.String())}
	if response.Header().Get("X-Content-Type-Options")!="nosniff"{t.Fatal("security headers were not applied")}
	if response.Header().Get("Content-Security-Policy")==""{t.Fatal("content security policy was not applied")}
	if response.Header().Get("Cache-Control")!="no-store"{t.Fatal("API response should not be cached")}
}

func TestRejectsNonLoopbackHost(t *testing.T){
	server:=testServer(t)
	req:=httptest.NewRequest(http.MethodGet,"http://attacker.example/api/status",nil)
	req.Host="attacker.example"
	response:=httptest.NewRecorder()
	server.Handler().ServeHTTP(response,req)
	if response.Code!=http.StatusForbidden{t.Fatalf("status = %d, want 403",response.Code)}
}

func TestRejectsMutationWithoutSameOrigin(t *testing.T){
	server:=testServer(t)
	tests:=[]struct{name,host,origin,fetchSite string}{
		{name:"missing origin",host:"127.0.0.1:8765"},
		{name:"foreign origin",host:"127.0.0.1:8765",origin:"https://attacker.example"},
		{name:"cross-site fetch",host:"127.0.0.1:8765",origin:"http://127.0.0.1:8765",fetchSite:"cross-site"},
	}
	for _,test:=range tests{t.Run(test.name,func(t *testing.T){
		req:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8765/api/connect/start",nil)
		req.Host=test.host
		if test.origin!=""{req.Header.Set("Origin",test.origin)}
		if test.fetchSite!=""{req.Header.Set("Sec-Fetch-Site",test.fetchSite)}
		response:=httptest.NewRecorder()
		server.Handler().ServeHTTP(response,req)
		if response.Code!=http.StatusForbidden{t.Fatalf("status = %d, want 403",response.Code)}
	})}
}

func TestCloudPathMustStayInsideApplicationFolder(t *testing.T){
	server:=testServer(t)
	for _,value:=range []string{"/Apps/TBox-Drive","/Apps/TBox-Drive/docs/file.pdf"}{
		if err:=server.validateCloudPath(value);err!=nil{t.Errorf("valid path %q rejected: %v",value,err)}
	}
	for _,value:=range []string{"/Apps/Other/file.pdf","/Apps/TBox-Drive/../../Other/file.pdf","/Apps/TBox-Drive\\secret"}{
		if err:=server.validateCloudPath(value);err==nil{t.Errorf("unsafe path %q accepted",value)}
	}
}

func TestListenAndServeRejectsPublicBindAddress(t *testing.T){
	server:=testServer(t)
	if err:=server.ListenAndServe(context.Background(),"0.0.0.0:8765");err==nil{t.Fatal("expected public bind address to be rejected")}
}

func TestAccountsEndpointDoesNotExposeCredentialReference(t *testing.T) {
	server := testServer(t)
	item := account.Account{ID:"terabox:123",ProviderID:"terabox",DisplayName:"Personal account",CredentialRef:"vault-secret-reference"}
	if err := server.store.Save(context.Background(),item); err != nil { t.Fatal(err) }
	req := httptest.NewRequest(http.MethodGet,"http://127.0.0.1:8765/api/accounts",nil)
	req.Host = "127.0.0.1:8765"
	req.AddCookie(&http.Cookie{Name:"tbox_local_session",Value:server.sessionToken})
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response,req)
	if response.Code != http.StatusOK { t.Fatalf("status = %d, body=%s",response.Code,response.Body.String()) }
	body := response.Body.String()
	if strings.Contains(body,"credentialRef") || strings.Contains(body,"vault-secret-reference") {
		t.Fatalf("account response exposed credential metadata: %s",body)
	}
	if !strings.Contains(body,"Personal account") { t.Fatalf("account metadata missing: %s",body) }
}

func TestAPIRequiresLocalBrowserSession(t *testing.T) {
	server := testServer(t)
	req := httptest.NewRequest(http.MethodGet,"http://127.0.0.1:8765/api/status",nil)
	req.Host = "127.0.0.1:8765"
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response,req)
	if response.Code != http.StatusUnauthorized { t.Fatalf("status = %d, want 401",response.Code) }
}

func TestLaunchTokenExchangesForHttpOnlySessionCookie(t *testing.T) {
	server := testServer(t)
	req := httptest.NewRequest(http.MethodPost,"http://127.0.0.1:8765/api/session",nil)
	req.Host = "127.0.0.1:8765"
	req.Header.Set("Origin","http://127.0.0.1:8765")
	req.Header.Set("X-TBox-Session",server.sessionToken)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response,req)
	if response.Code != http.StatusNoContent { t.Fatalf("status = %d, body=%s",response.Code,response.Body.String()) }
	cookie := response.Result().Cookies()
	if len(cookie)!=1 || cookie[0].Name!="tbox_local_session" || !cookie[0].HttpOnly || cookie[0].SameSite!=http.SameSiteStrictMode {
		t.Fatalf("unexpected session cookie: %#v",cookie)
	}
}

func TestParsePageDefaultsAndValidates(t *testing.T) {
	tests := []struct {
		name string
		value string
		want int
	}{
		{name: "empty", value: "", want: 1},
		{name: "zero", value: "0", want: 1},
		{name: "negative", value: "-4", want: 1},
		{name: "invalid", value: "next", want: 1},
		{name: "page two", value: "2", want: 2},
		{name: "large page", value: "250", want: 250},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parsePage(test.value); got != test.want {
				t.Fatalf("parsePage(%q) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}

func TestMoveEndpointRejectsPathsOutsideApplicationFolder(t *testing.T) {
	server := testServer(t)
	body := `{"accountID":"missing-account","path":"/Apps/TBox-Drive/file.txt","destination":"/Apps/Other"}`
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8765/api/files/move", strings.NewReader(body))
	req.Host = "127.0.0.1:8765"
	req.Header.Set("Origin", "http://127.0.0.1:8765")
	req.AddCookie(&http.Cookie{Name: "tbox_local_session", Value: server.sessionToken})
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "inside the TBox Drive application folder") {
		t.Fatalf("unexpected validation response: %s", response.Body.String())
	}
}

func TestValidateFileNameRejectsPathComponents(t *testing.T) {
	for _, name := range []string{"", " ", ".", "..", "../secret", `folder\\file.txt`, "folder/file.txt"} {
		if err := validateFileName(name); err == nil {
			t.Errorf("validateFileName(%q) accepted an unsafe name", name)
		}
	}
	for _, name := range []string{"report.pdf", "folder name", "資料.txt"} {
		if err := validateFileName(name); err != nil {
			t.Errorf("validateFileName(%q) rejected a valid filename: %v", name, err)
		}
	}
}

func TestCopyEndpointRejectsPathsOutsideApplicationFolder(t *testing.T) {
	server := testServer(t)
	body := `{"accountID":"missing-account","path":"/Apps/TBox-Drive/file.txt","destination":"/Apps/Other"}`
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8765/api/files/copy", strings.NewReader(body))
	req.Host = "127.0.0.1:8765"
	req.Header.Set("Origin", "http://127.0.0.1:8765")
	req.AddCookie(&http.Cookie{Name: "tbox_local_session", Value: server.sessionToken})
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "inside the TBox Drive application folder") {
		t.Fatalf("unexpected validation response: %s", response.Body.String())
	}
}

func TestMoveAndCopyEndpointsRejectDescendantDestinations(t *testing.T) {
	for _, endpoint := range []string{"/api/files/move", "/api/files/copy"} {
		t.Run(endpoint, func(t *testing.T) {
			server := testServer(t)
			body := `{"accountID":"missing-account","path":"/Apps/TBox-Drive/folder","destination":"/Apps/TBox-Drive/folder/child"}`
			req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8765"+endpoint, strings.NewReader(body))
			req.Host = "127.0.0.1:8765"
			req.Header.Set("Origin", "http://127.0.0.1:8765")
			req.AddCookie(&http.Cookie{Name: "tbox_local_session", Value: server.sessionToken})
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, req)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), "descendants") {
				t.Fatalf("unexpected validation response: %s", response.Body.String())
			}
		})
	}
}

func TestDownloadResponseWriterPersistsStreamingProgress(t *testing.T) {
	server := testServer(t)
	ctx := context.Background()
	task := transfer.Task{ID: "web-download-progress", ProviderID: "terabox", Direction: "download", Source: "42", Destination: "file.bin", Size: 5, Status: transfer.StatusQueued}
	if err := server.store.CreateTransfer(ctx, task); err != nil { t.Fatal(err) }
	if err := server.store.UpdateProgress(ctx, task.ID, 0, transfer.StatusRunning, ""); err != nil { t.Fatal(err) }

	var output bytes.Buffer
	writer := &downloadResponseWriter{writer: &output, store: server.store, transferID: task.ID, total: 5}
	if _, err := writer.Write([]byte("hello")); err != nil { t.Fatal(err) }
	writer.report()
	if output.String() != "hello" { t.Fatalf("response body = %q, want hello", output.String()) }
	got, err := server.store.GetTransfer(ctx, task.ID)
	if err != nil { t.Fatal(err) }
	if got.Transferred != 5 || got.Status != transfer.StatusRunning {
		t.Fatalf("transfer progress = %d/%d status %s, want 5/5 running", got.Transferred, got.Size, got.Status)
	}
}
