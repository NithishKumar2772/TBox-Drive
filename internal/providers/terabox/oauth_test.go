package terabox

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSignature(t *testing.T) {
	c,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:"https://example.test"})
	if err!=nil{t.Fatal(err)}
	ts:=int64(1700000000);sum:=md5.Sum([]byte(fmt.Sprintf("id_%d_secret_private",ts)))
	if got,want:=c.signature(ts),hex.EncodeToString(sum[:]);got!=want{t.Fatalf("got %s want %s",got,want)}
}
func TestBeginDeviceAuthorization(t *testing.T) {
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.URL.Path!="/oauth/devicecode"||r.URL.Query().Get("client_id")!="id"{t.Errorf("unexpected request: %s",r.URL.String())}
		fmt.Fprint(w,`{"errno":0,"data":{"device_code":"device-1","qrcode_url":"data:image/png;base64,AA==","expires_in":600,"interval":5}}`)
	}));defer s.Close()
	c,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:s.URL});if err!=nil{t.Fatal(err)}
	got,err:=c.BeginDeviceAuthorization(context.Background());if err!=nil{t.Fatal(err)}
	if got.DeviceCode!="device-1"||got.Interval!=5{t.Fatalf("unexpected response: %#v",got)}
}
func TestExchangeCodeUsesMultipart(t *testing.T) {
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if err:=r.ParseMultipartForm(1024);err!=nil{t.Error(err)}
		if r.Form.Get("grant_type")!="device_code"||r.Form.Get("code")!="device-1"{t.Errorf("unexpected form: %#v",r.Form)}
		if r.Form.Get("sign")==""||r.Form.Get("timestamp")==""{t.Error("signature fields missing")}
		fmt.Fprint(w,`{"errno":0,"data":{"access_token":"access","refresh_token":"refresh","expires_in":172800}}`)
	}));defer s.Close()
	c,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:s.URL});if err!=nil{t.Fatal(err)}
	got,err:=c.ExchangeCode(context.Background(),"device_code","device-1");if err!=nil{t.Fatal(err)}
	if got.AccessToken!="access"||got.RefreshToken!="refresh"{t.Fatalf("unexpected token: %#v",got)}
}
func TestRefreshRejectsEmptyToken(t *testing.T) {
	c,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private"});if err!=nil{t.Fatal(err)}
	if _,err:=c.Refresh(context.Background()," ");err==nil{t.Fatal("expected empty refresh token error")}
}
func TestTokenInfo(t *testing.T) {
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		body,_:=io.ReadAll(r.Body);v,_:=url.ParseQuery(string(body))
		if v.Get("access_token")!="access"{t.Error("unexpected access token")}
		fmt.Fprint(w,`{"errno":0,"data":{"api_domain":"api.example.test","upload_domain":"upload.example.test","user_id":42}}`)
	}));defer s.Close()
	c,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:s.URL});if err!=nil{t.Fatal(err)}
	info,err:=c.TokenInformation(context.Background(),"access");if err!=nil{t.Fatal(err)}
	if info.APIDomain!="api.example.test"||info.UserID!=42{t.Fatalf("unexpected info: %#v",info)}
}
func TestOAuthRejectsInsecureRemoteBaseURL(t *testing.T) {
	_,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:"http://example.com"})
	if err==nil{t.Fatal("expected insecure URL rejection")}
}


func TestOAuthDoesNotFollowCredentialBearingRedirects(t *testing.T) {
	var redirectedRequests atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectedRequests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	client, err := NewOAuthClient(OAuthConfig{
		ClientID: "id", ClientSecret: "secret", PrivateSecret: "private", BaseURL: origin.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.BeginDeviceAuthorization(context.Background())
	if err == nil || !strings.Contains(err.Error(), "HTTP 307") {
		t.Fatalf("expected redirect response to be rejected, got %v", err)
	}
	if got := redirectedRequests.Load(); got != 0 {
		t.Fatalf("credentials were forwarded through redirect; target requests = %d", got)
	}
}


func TestTokenInfoRejectsMissingUploadDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		fmt.Fprint(w,`{"errno":0,"data":{"api_domain":"api.terabox.com","user_id":42}}`)
	}))
	defer server.Close()
	client,err := NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:server.URL})
	if err != nil { t.Fatal(err) }
	if _,err := client.TokenInformation(context.Background(),"access"); err == nil {
		t.Fatal("expected token information without upload domain to be rejected")
	}
}
