package terabox

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdapterListUsesConnectedClient(t *testing.T){
	s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		fmt.Fprint(w,`{"errno":0,"has_more":0,"info":[{"fs_id":77,"path":"/From: Other Applications/TBox-Drive-123/file.bin","server_filename":"file.bin","size":8,"isdir":0}]}`)
	}));defer s.Close()
	api,err:=NewAPIClient(APIConfig{AccessToken:"token",TokenInfo:TokenInfo{APIDomain:s.URL,UploadDomain:s.URL},ApplicationRoot:"/From: Other Applications/TBox-Drive-123"})
	if err!=nil{t.Fatal(err)}
	a:=NewConnected(api)
	files,next,err:=a.List(context.Background(),"","")
	if err!=nil{t.Fatal(err)}
	if next!=""||len(files)!=1||files[0].ID!="77"{t.Fatalf("files=%#v next=%q",files,next)}
}
func TestAdapterListRejectsInvalidPageToken(t *testing.T){
	a:=NewConnected(&APIClient{applicationRoot:"/root"})
	if _,_,err:=a.List(context.Background(),"/root","not-a-page");err==nil{t.Fatal("expected invalid page token")}
}
func TestUnconfiguredAdapterReturnsNotImplemented(t *testing.T){
	if _,_,err:=New().List(context.Background(),"/","");err!=ErrNotImplemented{t.Fatalf("got %v",err)}
}
