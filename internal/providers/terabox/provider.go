package terabox

import (
	"context"
	"errors"
	"io"
	"path"
	"strconv"
	"strings"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
)

var ErrNotImplemented = errors.New("TeraBox operation is not implemented")
type Adapter struct { api *APIClient; repository transfer.Repository }
func New() *Adapter { return &Adapter{} }
func NewConnected(api *APIClient) *Adapter { return &Adapter{api:api} }
func NewConnectedWithRepository(api *APIClient,repository transfer.Repository)*Adapter{return &Adapter{api:api,repository:repository}}
func (*Adapter) ID() string { return "terabox" }
func (*Adapter) Name() string { return "TeraBox" }
func (*Adapter) Capabilities() provider.Capabilities { return provider.Capabilities{MultipartUpload:true,ResumableUpload:true,StreamingDownload:true} }
func (a *Adapter) List(ctx context.Context,parentPath,pageToken string)([]provider.File,string,error){
	if a.api==nil{return nil,"",ErrNotImplemented}
	if strings.TrimSpace(parentPath)==""{parentPath=a.api.applicationRoot+"/"}
	page:=1;if pageToken!=""{n,err:=strconv.Atoi(pageToken);if err!=nil||n<1{return nil,"",errors.New("invalid TeraBox page token")};page=n}
	files,more,err:=a.api.ListDirectory(ctx,parentPath,page,100);if err!=nil{return nil,"",err};if !more{return files,"",nil};return files,strconv.Itoa(page+1),nil
}
func (a *Adapter) Upload(ctx context.Context,req provider.UploadRequest)(provider.UploadResult,error){
	if a.api==nil{return provider.UploadResult{},ErrNotImplemented}
	name:=req.Name;if name==""{name=path.Base(strings.ReplaceAll(req.LocalPath,"\\","/"))}
	if name=="."||name=="/"||name==".."||strings.ContainsAny(name,"/\\"){return provider.UploadResult{},errors.New("upload name must be a single safe filename")}
	parent:=req.ParentID;if parent==""{parent=a.api.applicationRoot}
	return a.api.UploadFile(ctx,req.LocalPath,path.Join(parent,name),UploadOptions{Repository:a.repository})
}
func (a *Adapter) Download(ctx context.Context,fileID string)(io.ReadCloser,error){if a.api==nil{return nil,ErrNotImplemented};return a.api.Download(ctx,fileID)}
func (*Adapter) Delete(context.Context,string)error{return ErrNotImplemented}
