package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/account"
	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
)

type memorySecrets struct{mu sync.Mutex;items map[string][]byte}
func(m *memorySecrets)Get(_ context.Context,key string)([]byte,error){m.mu.Lock();defer m.mu.Unlock();v,ok:=m.items[key];if !ok{return nil,fmt.Errorf("secret not found")};return append([]byte(nil),v...),nil}
func(m *memorySecrets)Set(_ context.Context,key string,value []byte)error{m.mu.Lock();defer m.mu.Unlock();if m.items==nil{m.items=map[string][]byte{}};m.items[key]=append([]byte(nil),value...);return nil}
func(m *memorySecrets)Delete(_ context.Context,key string)error{m.mu.Lock();defer m.mu.Unlock();delete(m.items,key);return nil}

func TestTeraBoxConnectStoresSecretsOutsideDatabase(t *testing.T){
	var server *httptest.Server
	server=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		switch r.URL.Path{
		case "/oauth/devicecode":fmt.Fprint(w,`{"errno":0,"data":{"device_code":"device-1","qrcode_url":"data:image/png;base64,AA==","expires_in":20,"interval":1}}`)
		case "/oauth/gettoken":fmt.Fprint(w,`{"errno":0,"data":{"access_token":"access-secret","refresh_token":"refresh-secret","expires_in":172800}}`)
		case "/oauth/tokeninfo":fmt.Fprintf(w,`{"errno":0,"data":{"api_domain":%q,"upload_domain":%q,"user_id":42}}`,server.URL,server.URL)
		default:http.NotFound(w,r)
		}
	}))
	defer server.Close()
	oauth,err:=terabox.NewOAuthClient(terabox.OAuthConfig{ClientID:"client",ClientSecret:"secret",PrivateSecret:"private",BaseURL:server.URL});if err!=nil{t.Fatal(err)}
	store,err:=sqlite.Open(":memory:");if err!=nil{t.Fatal(err)};defer store.Close()
	secrets:=&memorySecrets{}
	service,err:=NewTeraBoxService(oauth,secrets,store,nil,"/From: Other Applications/TBox-Drive-42");if err!=nil{t.Fatal(err)}
	device,err:=service.BeginConnect(context.Background());if err!=nil{t.Fatal(err)}
	account,err:=service.CompleteConnect(context.Background(),device);if err!=nil{t.Fatal(err)}
	if account.ID!="terabox:42"||account.CredentialRef!="terabox:42"{t.Fatalf("account %#v",account)}
	if _,err:=service.NewAPIClient(context.Background(),account.ID);err==nil{t.Fatal("expected production API client to reject loopback token-info domains")}
	stored,err:=secrets.Get(context.Background(),"terabox:42");if err!=nil{t.Fatal(err)}
	var credential storedTeraBoxCredential;if err:=json.Unmarshal(stored,&credential);err!=nil{t.Fatal(err)}
	if credential.Token.AccessToken!="access-secret"{t.Fatal("access token not stored in secret store")}
	persisted,err:=store.Get(context.Background(),account.ID);if err!=nil{t.Fatal(err)}
	if persisted.CredentialRef!="terabox:42"{t.Fatalf("database should contain only credential reference: %#v",persisted)}
}


type failingDeleteAccountRepository struct {
	account.Repository
}

func (r failingDeleteAccountRepository) Delete(context.Context, string) error {
	return errors.New("simulated account database failure")
}

func TestDisconnectDoesNotDeleteSecretWhenAccountDeleteFails(t *testing.T) {
	ctx := context.Background()
	store, err := sqlite.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer store.Close()
	item := account.Account{ID:"terabox:77",ProviderID:"terabox",CredentialRef:"terabox:77",DisplayName:"Test account"}
	if err := store.Save(ctx,item); err != nil { t.Fatal(err) }
	secrets := &memorySecrets{}
	if err := secrets.Set(ctx,item.CredentialRef,[]byte("credential-payload")); err != nil { t.Fatal(err) }
	oauth, err := terabox.NewOAuthClient(terabox.OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private"})
	if err != nil { t.Fatal(err) }
	service, err := NewTeraBoxService(oauth,secrets,failingDeleteAccountRepository{Repository:store},nil,"/app")
	if err != nil { t.Fatal(err) }
	if err := service.Disconnect(ctx,item.ID); err == nil { t.Fatal("expected account deletion failure") }
	if _, err := secrets.Get(ctx,item.CredentialRef); err != nil {
		t.Fatalf("credential was deleted despite account database failure: %v",err)
	}
}

func TestRefreshTokenIsPersistedBeforeTokenInfoRequest(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		switch r.URL.Path {
		case "/oauth/refreshtoken":
			if err := r.ParseMultipartForm(1<<20); err != nil { t.Error(err) }
			if r.Form.Get("refresh_token") != "old-refresh" { t.Errorf("unexpected refresh token: %q",r.Form.Get("refresh_token")) }
			fmt.Fprint(w,`{"errno":0,"data":{"access_token":"new-access","refresh_token":"new-refresh","expires_in":172800}}`)
		case "/oauth/tokeninfo":
			http.Error(w,"temporary provider failure",http.StatusBadGateway)
		default:
			http.NotFound(w,r)
		}
	}))
	defer server.Close()
	oauth, err := terabox.NewOAuthClient(terabox.OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:server.URL})
	if err != nil { t.Fatal(err) }
	store, err := sqlite.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer store.Close()
	item := account.Account{ID:"terabox:88",ProviderID:"terabox",CredentialRef:"terabox:88",DisplayName:"Test account"}
	if err := store.Save(context.Background(),item); err != nil { t.Fatal(err) }
	now := time.Date(2026,11,23,12,0,0,0,time.UTC)
	secrets := &memorySecrets{}
	initial, err := json.Marshal(storedTeraBoxCredential{
		Token:terabox.OAuthToken{AccessToken:"old-access",RefreshToken:"old-refresh",ExpiresIn:3600},
		AccessExpiresAt:now.Add(-time.Minute),
		TokenInfo:terabox.TokenInfo{APIDomain:"api.terabox.com",UploadDomain:"upload.terabox.com",UserID:88},
		TokenInfoFetchedAt:now.Add(-2*time.Hour),
	})
	if err != nil { t.Fatal(err) }
	if err := secrets.Set(context.Background(),item.CredentialRef,initial); err != nil { t.Fatal(err) }
	service, err := NewTeraBoxService(oauth,secrets,store,nil,"/app")
	if err != nil { t.Fatal(err) }
	service.now = func() time.Time { return now }
	if _, err := service.NewAPIClient(context.Background(),item.ID); err == nil { t.Fatal("expected token-info request to fail") }
	stored, err := secrets.Get(context.Background(),item.CredentialRef)
	if err != nil { t.Fatal(err) }
	var got storedTeraBoxCredential
	if err := json.Unmarshal(stored,&got); err != nil { t.Fatal(err) }
	if got.Token.RefreshToken != "new-refresh" || got.Token.AccessToken != "new-access" {
		t.Fatalf("rotated token was not persisted before token-info failure: %#v",got.Token)
	}
}
