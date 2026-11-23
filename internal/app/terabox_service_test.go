package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
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
