package terabox

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
)

var ErrNotImplemented = errors.New("TeraBox operation is not implemented")

type Adapter struct { api *APIClient }
func New() *Adapter { return &Adapter{} }
func NewConnected(api *APIClient) *Adapter { return &Adapter{api:api} }
func (*Adapter) ID() string { return "terabox" }
func (*Adapter) Name() string { return "TeraBox" }
func (*Adapter) Capabilities() provider.Capabilities { return provider.Capabilities{} }

func (a *Adapter) List(ctx context.Context,parentPath,pageToken string)([]provider.File,string,error){
	if a.api==nil{return nil,"",ErrNotImplemented}
	if strings.TrimSpace(parentPath)==""{parentPath=a.api.applicationRoot+"/"}
	page:=1
	if pageToken!=""{n,err:=strconv.Atoi(pageToken);if err!=nil||n<1{return nil,"",errors.New("invalid TeraBox page token")};page=n}
	files,more,err:=a.api.ListDirectory(ctx,parentPath,page,100)
	if err!=nil{return nil,"",err}
	if !more{return files,"",nil}
	return files,strconv.Itoa(page+1),nil
}
func (*Adapter) Download(context.Context,string)(io.ReadCloser,error){return nil,ErrNotImplemented}
func (*Adapter) Upload(context.Context,provider.UploadRequest)(provider.UploadResult,error){return provider.UploadResult{},ErrNotImplemented}
func (*Adapter) Delete(context.Context,string)error{return ErrNotImplemented}
