package terabox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
)

type APIConfig struct { AccessToken string; TokenInfo TokenInfo; ApplicationRoot string; HTTPClient *http.Client }
type APIClient struct { accessToken,apiBase,uploadBase,applicationRoot string; http,downloadHTTP *http.Client }
type Quota struct { Total uint64 `json:"total"`; Used uint64 `json:"used"` }
type apiFile struct { FSID uint64 `json:"fs_id"`; Path string `json:"path"`; Name string `json:"server_filename"`; Size uint64 `json:"size"`; IsDir int `json:"isdir"`; Modified uint64 `json:"server_mtime"` }
type listResponse struct { Errno int `json:"errno"`; ShowMsg string `json:"show_msg"`; Info []apiFile `json:"info"`; List []apiFile `json:"list"`; HasMore int `json:"has_more"` }
type quotaResponse struct { Errno int `json:"errno"`; ShowMsg string `json:"show_msg"`; Total uint64 `json:"total"`; Used uint64 `json:"used"` }

func NewAPIClient(cfg APIConfig)(*APIClient,error){
	if strings.TrimSpace(cfg.AccessToken)==""{return nil,errors.New("TeraBox access token is required")}
	apiBase,err:=serviceBase(cfg.TokenInfo.APIDomain);if err!=nil{return nil,fmt.Errorf("invalid TeraBox API domain: %w",err)}
	uploadBase,err:=serviceBase(cfg.TokenInfo.UploadDomain);if err!=nil{return nil,fmt.Errorf("invalid TeraBox upload domain: %w",err)}
	if cfg.ApplicationRoot!=""&&!strings.HasPrefix(cfg.ApplicationRoot,"/"){return nil,errors.New("application root must be an absolute TeraBox path")}
	client:=cfg.HTTPClient;if client==nil{client=&http.Client{Timeout:2*time.Minute}}
	downloadClient:=*client;downloadClient.Timeout=0
	return &APIClient{accessToken:cfg.AccessToken,apiBase:apiBase,uploadBase:uploadBase,applicationRoot:strings.TrimRight(cfg.ApplicationRoot,"/"),http:client,downloadHTTP:&downloadClient},nil
}
func serviceBase(raw string)(string,error){
	raw=strings.TrimSpace(raw);if raw==""{return "",errors.New("domain is empty")};if !strings.Contains(raw,"://"){raw="https://"+raw}
	u,err:=url.Parse(raw);if err!=nil{return "",err}
	localHTTP:=u.Scheme=="http"&&(u.Hostname()=="127.0.0.1"||u.Hostname()=="localhost"||u.Hostname()=="::1")
	if u.Host==""||(u.Scheme!="https"&&!localHTTP)||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||(u.Path!=""&&u.Path!="/"){return "",errors.New("service domain must be an HTTPS origin")}
	return strings.TrimRight(u.String(),"/"),nil
}
func(c *APIClient)ListDirectory(ctx context.Context,dir string,page,pageSize int)([]provider.File,bool,error){
	if page<1{return nil,false,errors.New("page must be at least 1")};if pageSize<1||pageSize>10000{return nil,false,errors.New("page size must be between 1 and 10000")}
	if err:=c.validateRemotePath(dir);err!=nil{return nil,false,err}
	q:=url.Values{"access_tokens":{c.accessToken},"dir":{dir},"page":{strconv.Itoa(page)},"num":{strconv.Itoa(pageSize)}}
	req,err:=http.NewRequestWithContext(ctx,http.MethodGet,c.apiBase+"/openapi/api/list?"+q.Encode(),nil);if err!=nil{return nil,false,err}
	var result listResponse;if err:=c.getJSON(req,&result);err!=nil{return nil,false,err}
	items:=result.Info;if items==nil{items=result.List};files:=make([]provider.File,0,len(items))
	for _,item:=range items{name:=item.Name;if name==""{name=pathBase(item.Path)};files=append(files,provider.File{ID:strconv.FormatUint(item.FSID,10),Name:name,Path:item.Path,Size:int64(item.Size),IsDirectory:item.IsDir==1,ModifiedAt:time.Unix(int64(item.Modified),0)})}
	return files,result.HasMore==1,nil
}
func(c *APIClient)StorageQuota(ctx context.Context)(Quota,error){
	q:=url.Values{"access_tokens":{c.accessToken}};req,err:=http.NewRequestWithContext(ctx,http.MethodGet,c.apiBase+"/openapi/api/quota?"+q.Encode(),nil);if err!=nil{return Quota{},err}
	var result quotaResponse;if err:=c.getJSON(req,&result);err!=nil{return Quota{},err};return Quota{Total:result.Total,Used:result.Used},nil
}
func(c *APIClient)validateRemotePath(remotePath string)error{
	if strings.TrimSpace(remotePath)==""||!strings.HasPrefix(remotePath,"/")||strings.Contains(remotePath,"\\"){return errors.New("TeraBox path must be an absolute cloud path")}
	for _,part:=range strings.Split(remotePath,"/"){if part=="."||part==".."{return errors.New("TeraBox path cannot contain traversal segments")}}
	if c.applicationRoot!=""&&remotePath!=c.applicationRoot&&!strings.HasPrefix(remotePath,c.applicationRoot+"/"){return errors.New("path is outside the TeraBox application folder")}
	return nil
}
func(c *APIClient)getJSON(req *http.Request,dst any)error{
	resp,err:=c.http.Do(req);if err!=nil{return fmt.Errorf("TeraBox API request failed: %w",err)};defer resp.Body.Close()
	const max=4<<20;body,err:=io.ReadAll(io.LimitReader(resp.Body,max+1));if err!=nil{return fmt.Errorf("read TeraBox API response: %w",err)}
	if len(body)>max{return errors.New("TeraBox API response exceeded 4 MiB")};if resp.StatusCode<200||resp.StatusCode>=300{return fmt.Errorf("TeraBox API returned HTTP %d",resp.StatusCode)}
	if err:=json.Unmarshal(body,dst);err!=nil{return fmt.Errorf("decode TeraBox API response: %w",err)}
	var status struct{Errno int `json:"errno"`;ShowMsg string `json:"show_msg"`}
	if err:=json.Unmarshal(body,&status);err==nil&&status.Errno!=0{msg:=strings.TrimSpace(status.ShowMsg);if msg==""{msg="request rejected"};return fmt.Errorf("TeraBox API error %d: %s",status.Errno,msg)}
	return nil
}
func pathBase(p string)string{p=strings.TrimRight(p,"/");if p==""{return "/"};i:=strings.LastIndex(p,"/");return p[i+1:]}
