package terabox

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWaitForDeviceAuthorizationPollsUntilApproved(t *testing.T){
	attempts:=0
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		attempts++
		if r.URL.Path!="/oauth/gettoken"{t.Errorf("unexpected path %s",r.URL.Path)}
		if attempts==1{fmt.Fprint(w,`{"errno":400001,"show_msg":"not approved"}`);return}
		fmt.Fprint(w,`{"errno":0,"data":{"access_token":"access","refresh_token":"refresh","expires_in":172800}}`)
	}));defer s.Close()
	c,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:s.URL});if err!=nil{t.Fatal(err)}
	start:=time.Now()
	token,err:=c.WaitForDeviceAuthorization(context.Background(),DeviceAuthorization{DeviceCode:"device",ExpiresIn:10,Interval:1})
	if err!=nil{t.Fatal(err)}
	if token.AccessToken!="access"||attempts!=2{t.Fatalf("token=%#v attempts=%d",token,attempts)}
	if time.Since(start)<time.Second{t.Fatal("polling interval was not respected")}
}
func TestWaitForDeviceAuthorizationHonorsCancellation(t *testing.T){
	c,err:=NewOAuthClient(OAuthConfig{ClientID:"id",ClientSecret:"secret",PrivateSecret:"private",BaseURL:"https://example.test"});if err!=nil{t.Fatal(err)}
	ctx,cancel:=context.WithCancel(context.Background());cancel()
	_,err=c.WaitForDeviceAuthorization(ctx,DeviceAuthorization{DeviceCode:"device",ExpiresIn:10,Interval:1})
	if !errors.Is(err,context.Canceled){t.Fatalf("got %v",err)}
}
