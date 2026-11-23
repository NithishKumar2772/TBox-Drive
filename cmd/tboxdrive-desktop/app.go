//go:build desktop

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/app"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/account"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
	"github.com/NithishKumar2772/TBox-Drive/internal/storage/sqlite"
	"github.com/NithishKumar2772/TBox-Drive/internal/virtualdrive"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppStatus struct { Ready bool `json:"ready"`; Message string `json:"message"`; ApplicationRoot string `json:"applicationRoot,omitempty"` }
type FilePage struct { Files []provider.File `json:"files"`; NextPageToken string `json:"nextPageToken,omitempty"` }
type UpdateInfo struct {
	LatestVersion string `json:"latestVersion"`
	ReleaseURL string `json:"releaseURL"`
	PublishedAt string `json:"publishedAt,omitempty"`
	CheckedAt time.Time `json:"checkedAt"`
}
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}
type DesktopApp struct { ctx context.Context; store *sqlite.Store; service *app.TeraBoxService; drive *virtualdrive.Manager; applicationRoot string; message string }
func NewDesktopApp()*DesktopApp{return &DesktopApp{drive:virtualdrive.NewManager(),message:"Starting TBox Drive…"}}
func(d *DesktopApp)startup(ctx context.Context){
	d.ctx=ctx
	configDir,err:=os.UserConfigDir();if err!=nil{d.message="Cannot locate the user configuration directory: "+err.Error();return}
	store,err:=sqlite.Open(filepath.Join(configDir,"TBox Drive","tbox.db"));if err!=nil{d.message="Cannot open local database: "+err.Error();return};d.store=store
	clientID:=os.Getenv("TBOX_TERABOX_CLIENT_ID");clientSecret:=os.Getenv("TBOX_TERABOX_CLIENT_SECRET");privateSecret:=os.Getenv("TBOX_TERABOX_PRIVATE_SECRET");d.applicationRoot=os.Getenv("TBOX_TERABOX_APP_FOLDER")
	if clientID==""||clientSecret==""||privateSecret==""||d.applicationRoot==""{d.message="TeraBox Open Platform setup is required. Configure TBOX_TERABOX_CLIENT_ID, TBOX_TERABOX_CLIENT_SECRET, TBOX_TERABOX_PRIVATE_SECRET and TBOX_TERABOX_APP_FOLDER.";return}
	oauth,err:=terabox.NewOAuthClient(terabox.OAuthConfig{ClientID:clientID,ClientSecret:clientSecret,PrivateSecret:privateSecret});if err!=nil{d.message="TeraBox OAuth configuration error: "+err.Error();return}
	secrets,err:=app.NewOSSecretStore();if err!=nil{d.message="Cannot initialize the operating-system credential vault: "+err.Error();return}
	service,err:=app.NewTeraBoxService(oauth,secrets,store,store,d.applicationRoot);if err!=nil{d.message="Cannot initialize TeraBox service: "+err.Error();return}
	d.service=service;d.message="TBox Drive is ready."
}
func(d *DesktopApp)shutdown(context.Context){if d.drive!=nil{d.drive.Unmount()};if d.store!=nil{_ = d.store.Close()}}
func(d *DesktopApp)GetAppStatus()AppStatus{return AppStatus{Ready:d.service!=nil,Message:d.message,ApplicationRoot:d.applicationRoot}}

func(d *DesktopApp)GetVirtualDriveStatus()virtualdrive.Status{
	if d.drive==nil{return virtualdrive.Status{Mounted:false,Message:"Virtual drive manager is unavailable."}}
	return d.drive.GetStatus()
}
func(d *DesktopApp)MountVirtualDrive(accountID,mountpoint string)(virtualdrive.Status,error){
	if err:=d.requireService();err!=nil{return virtualdrive.Status{},err}
	if d.drive==nil{return virtualdrive.Status{},errors.New("virtual drive manager is unavailable")}
	if _,err:=d.store.Get(d.ctx,accountID);err!=nil{return virtualdrive.Status{},fmt.Errorf("select a connected TeraBox account: %w",err)}
	cacheRoot,err:=os.UserCacheDir();if err!=nil{return virtualdrive.Status{},err}
	cacheDir:=filepath.Join(cacheRoot,"TBox Drive","drive-cache")
	factory:=func(ctx context.Context)(*terabox.APIClient,error){return d.service.NewAPIClient(ctx,accountID)}
	return d.drive.Mount(d.ctx,factory,d.applicationRoot,cacheDir,mountpoint)
}
func(d *DesktopApp)UnmountVirtualDrive()virtualdrive.Status{
	if d.drive==nil{return virtualdrive.Status{Mounted:false,Message:"Virtual drive manager is unavailable."}}
	return d.drive.Unmount()
}
func(d *DesktopApp)requireService()error{if d.service==nil{return errors.New(d.message)};return nil}
func(d *DesktopApp)StartTeraBoxConnect()(terabox.DeviceAuthorization,error){if err:=d.requireService();err!=nil{return terabox.DeviceAuthorization{},err};return d.service.BeginConnect(d.ctx)}
func(d *DesktopApp)CompleteTeraBoxConnect(device terabox.DeviceAuthorization)(account.Account,error){if err:=d.requireService();err!=nil{return account.Account{},err};return d.service.CompleteConnect(d.ctx,device)}
func(d *DesktopApp)ListAccounts()([]account.Account,error){if err:=d.requireService();err!=nil{return nil,err};return d.service.ListAccounts(d.ctx)}
func(d *DesktopApp)DisconnectAccount(accountID string)error{if err:=d.requireService();err!=nil{return err};return d.service.Disconnect(d.ctx,accountID)}
func(d *DesktopApp)ListFiles(accountID,parentPath,pageToken string)(FilePage,error){
	if err:=d.requireService();err!=nil{return FilePage{},err};api,err:=d.service.NewAPIClient(d.ctx,accountID);if err!=nil{return FilePage{},err}
	if strings.TrimSpace(parentPath)==""{parentPath=d.applicationRoot+"/"}
	files,more,err:=api.ListDirectory(d.ctx,parentPath,parsePage(pageToken),100);if err!=nil{return FilePage{},err}
	next:="";if more{next=fmt.Sprint(parsePage(pageToken)+1)};return FilePage{Files:files,NextPageToken:next},nil
}
func parsePage(token string)int{if token==""{return 1};var n int;if _,err:=fmt.Sscanf(token,"%d",&n);err!=nil||n<1{return 1};return n}
func(d *DesktopApp)GetStorageQuota(accountID string)(terabox.Quota,error){if err:=d.requireService();err!=nil{return terabox.Quota{},err};api,err:=d.service.NewAPIClient(d.ctx,accountID);if err!=nil{return terabox.Quota{},err};return api.StorageQuota(d.ctx)}
func(d *DesktopApp)SelectUploadFile()(string,error){return runtime.OpenFileDialog(d.ctx,runtime.OpenDialogOptions{Title:"Select a file to upload"})}
func(d *DesktopApp)SelectUploadFiles()([]string,error){return runtime.OpenMultipleFilesDialog(d.ctx,runtime.OpenDialogOptions{Title:"Select files to upload"})}
func(d *DesktopApp)UploadFile(accountID,localPath,parentPath,name,taskID string)(provider.UploadResult,error){
	if err:=d.requireService();err!=nil{return provider.UploadResult{},err};if localPath==""{return provider.UploadResult{},errors.New("select a local file first")}
	if name==""{name=path.Base(localPath)};if name=="."||name=="/"||name==".."||strings.ContainsAny(name,"/\\"){return provider.UploadResult{},errors.New("file name must not contain path separators")}
	if taskID==""{var id [16]byte;if _,err:=rand.Read(id[:]);err!=nil{return provider.UploadResult{},err};taskID="tbox-"+hex.EncodeToString(id[:])}
	if parentPath==""{parentPath=d.applicationRoot};api,err:=d.service.NewAPIClient(d.ctx,accountID);if err!=nil{return provider.UploadResult{},err}
	return api.UploadFile(d.ctx,localPath,path.Join(parentPath,name),terabox.UploadOptions{Repository:d.store,TransferID:taskID,OnProgress:func(done,total int64){runtime.EventsEmit(d.ctx,"transfer-progress",map[string]any{"transferId":taskID,"transferred":done,"total":total})}})
}
func(d *DesktopApp)RenameFile(accountID,remotePath,newName string)error{if err:=d.requireService();err!=nil{return err};api,err:=d.service.NewAPIClient(d.ctx,accountID);if err!=nil{return err};return api.Rename(d.ctx,remotePath,newName)}
func(d *DesktopApp)DeleteFile(accountID,remotePath string)error{if err:=d.requireService();err!=nil{return err};api,err:=d.service.NewAPIClient(d.ctx,accountID);if err!=nil{return err};return api.DeletePath(d.ctx,remotePath)}
func(d *DesktopApp)MoveFile(accountID,remotePath,destinationDir,newName string)error{if err:=d.requireService();err!=nil{return err};api,err:=d.service.NewAPIClient(d.ctx,accountID);if err!=nil{return err};return api.Move(d.ctx,remotePath,destinationDir,newName)}
func(d *DesktopApp)DownloadFile(accountID,fileID,fileName string)(string,error){
	if err:=d.requireService();err!=nil{return "",err};api,err:=d.service.NewAPIClient(d.ctx,accountID);if err!=nil{return "",err}
	destination,err:=runtime.SaveFileDialog(d.ctx,runtime.SaveDialogOptions{Title:"Save downloaded file",DefaultFilename:fileName});if err!=nil{return "",err};if destination==""{return "",nil}
	body,err:=api.Download(d.ctx,fileID);if err!=nil{return "",err};defer body.Close()
	temp,err:=os.CreateTemp(filepath.Dir(destination),".tbox-download-*");if err!=nil{return "",err}
	tempName:=temp.Name();committed:=false;defer func(){_ = temp.Close();if !committed{_ = os.Remove(tempName)}}()
	if _,err:=io.Copy(temp,body);err!=nil{return "",fmt.Errorf("download file: %w",err)}
	if err:=temp.Sync();err!=nil{return "",err};if err:=temp.Close();err!=nil{return "",err}
	if err:=os.Rename(tempName,destination);err!=nil{return "",err};committed=true;return destination,nil
}

func(d *DesktopApp)CheckForUpdates()(UpdateInfo,error){
	ctx:=d.ctx
	if ctx==nil{ctx=context.Background()}
	req,err:=http.NewRequestWithContext(ctx,http.MethodGet,"https://api.github.com/repos/NithishKumar2772/TBox-Drive/releases/latest",nil)
	if err!=nil{return UpdateInfo{},err}
	req.Header.Set("Accept","application/vnd.github+json")
	req.Header.Set("User-Agent","TBox-Drive-Update-Checker")
	client:=&http.Client{Timeout:8*time.Second}
	response,err:=client.Do(req)
	if err!=nil{return UpdateInfo{},fmt.Errorf("check GitHub releases: %w",err)}
	defer response.Body.Close()
	if response.StatusCode!=http.StatusOK{return UpdateInfo{},fmt.Errorf("GitHub release check returned HTTP %d",response.StatusCode)}
	var release githubRelease
	if err:=json.NewDecoder(io.LimitReader(response.Body,1<<20)).Decode(&release);err!=nil{return UpdateInfo{},fmt.Errorf("decode GitHub release information: %w",err)}
	if strings.TrimSpace(release.TagName)==""||strings.TrimSpace(release.HTMLURL)==""{return UpdateInfo{},errors.New("latest GitHub release is missing its tag or URL")}
	return UpdateInfo{LatestVersion:release.TagName,ReleaseURL:release.HTMLURL,PublishedAt:release.PublishedAt,CheckedAt:time.Now().UTC()},nil
}

func(d *DesktopApp)ListTransfers(limit int)([]transfer.Task,error){if d.store==nil{return nil,errors.New("local database is not available")};if limit<1{limit=50};return d.store.ListTransfers(d.ctx,limit)}
