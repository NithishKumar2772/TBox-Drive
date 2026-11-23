package terabox

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"strings"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
	"github.com/NithishKumar2772/TBox-Drive/internal/core/transfer"
)

const (
	DefaultUploadChunkSize int64 = 8 << 20
	minMultiPartSize int64 = 4 << 20
	teraboxAppID = "250528"
	maxTeraBoxParts = 1_000_000
)
type uploadPart struct { Index int; Offset,Size int64 }
type UploadOptions struct {
	ChunkSize int64
	RenamePolicy *int
	OnProgress func(transferred,total int64)
	Repository transfer.Repository
	TransferID string
	MaxPartAttempts int
	RetryDelay time.Duration
	MaxRetryDelay time.Duration
}
type precreateResponse struct { Errno int `json:"errno"`;ShowMsg string `json:"show_msg"`;UploadID string `json:"uploadid"`;ReturnType int `json:"return_type"`;BlockList []int `json:"block_list"`;Info apiFile `json:"info"` }
type partResponse struct { Errno int `json:"errno"`;ShowMsg string `json:"show_msg"`;MD5 string `json:"md5"`;UploadID string `json:"uploadid"`;PartSeq int `json:"partseq"` }
type createResponse struct { Errno int `json:"errno"`;ShowMsg string `json:"show_msg"`;FSID uint64 `json:"fs_id"`;Path string `json:"path"`;Name string `json:"server_filename"`;Size uint64 `json:"size"`;IsDir int `json:"isdir"`;Modified uint64 `json:"mtime"` }

func PlanTeraBoxChunks(fileSize,preferredSize int64)([]uploadPart,error){
	if fileSize<0{return nil,errors.New("file size cannot be negative")}
	if preferredSize<=minMultiPartSize{return nil,errors.New("preferred chunk size must exceed 4 MiB")}
	if fileSize==0{return nil,errors.New("empty-file upload is not yet supported")}
	count:=1+(fileSize-1)/preferredSize
	for count>1&&fileSize/count<=minMultiPartSize{count--}
	if count>maxTeraBoxParts{return nil,fmt.Errorf("upload requires too many parts (maximum %d)",maxTeraBoxParts)}
	base,extra:=fileSize/count,fileSize%count
	chunks:=make([]uploadPart,0,int(count));var offset int64
	for i:=int64(0);i<count;i++{size:=base;if i<extra{size++};if count>1&&size<=minMultiPartSize{return nil,errors.New("unable to satisfy TeraBox multi-part minimum size")};chunks=append(chunks,uploadPart{Index:int(i),Offset:offset,Size:size});offset+=size}
	return chunks,nil
}

func(c *APIClient)UploadFile(ctx context.Context,localPath,remotePath string,options UploadOptions)(result provider.UploadResult,err error){
	if err=c.validateRemotePath(remotePath);err!=nil{return provider.UploadResult{},err}
	chunkSize:=options.ChunkSize;if chunkSize==0{chunkSize=DefaultUploadChunkSize}
	file,openErr:=os.Open(localPath);if openErr!=nil{return provider.UploadResult{},fmt.Errorf("open upload source: %w",openErr)};defer file.Close()
	stat,statErr:=file.Stat();if statErr!=nil{return provider.UploadResult{},fmt.Errorf("stat upload source: %w",statErr)}
	if !stat.Mode().IsRegular(){return provider.UploadResult{},errors.New("upload source must be a regular file")}
	chunks,planErr:=PlanTeraBoxChunks(stat.Size(),chunkSize);if planErr!=nil{return provider.UploadResult{},planErr}

	taskID:=options.TransferID
	if taskID==""{var idErr error;taskID,idErr=newTransferID();if idErr!=nil{return provider.UploadResult{},idErr}}
	result.TaskID=taskID
	repo:=options.Repository
	transferred:=int64(0)
	if repo!=nil{
		existing,getErr:=repo.GetTransfer(ctx,taskID)
		if errors.Is(getErr,sql.ErrNoRows){
			createErr:=repo.CreateTransfer(ctx,transfer.Task{ID:taskID,ProviderID:"terabox",Direction:"upload",Source:localPath,Destination:remotePath,Size:stat.Size(),Status:transfer.StatusQueued})
			if createErr!=nil{return provider.UploadResult{},fmt.Errorf("create TeraBox transfer record: %w",createErr)}
		}else if getErr!=nil{return provider.UploadResult{},fmt.Errorf("load TeraBox transfer record: %w",getErr)
		}else if existing.ProviderID!="terabox"||existing.Direction!="upload"||existing.Source!=localPath||existing.Destination!=remotePath||existing.Size!=stat.Size(){
			return provider.UploadResult{},errors.New("transfer ID belongs to a different TeraBox upload")
		}
		defer func(){
			status:=transfer.StatusFailed;message:="upload failed"
			if err==nil{status=transfer.StatusCompleted;transferred=stat.Size();message=""}else if errors.Is(err,context.Canceled){status=transfer.StatusCanceled;message="upload canceled"}
			_ = repo.UpdateProgress(context.Background(),taskID,transferred,status,message)
		}()
		if err=repo.UpdateProgress(ctx,taskID,0,transfer.StatusRunning,"");err!=nil{return provider.UploadResult{},fmt.Errorf("start TeraBox transfer record: %w",err)}
	}

	blockMD5:=make([]string,len(chunks))
	for i,ch:=range chunks{if err=ctx.Err();err!=nil{return provider.UploadResult{},err};h:=md5.New();if _,err=io.Copy(h,io.NewSectionReader(file,ch.Offset,ch.Size));err!=nil{return provider.UploadResult{},fmt.Errorf("hash upload part %d: %w",i,err)};blockMD5[i]=hex.EncodeToString(h.Sum(nil))}
	if err=ensureUploadSourceUnchanged(localPath,file,stat);err!=nil{return provider.UploadResult{},err}
	blockJSON,_:=json.Marshal(blockMD5)
	pre,preErr:=c.precreate(ctx,remotePath,string(blockJSON));if preErr!=nil{return provider.UploadResult{},preErr}
	if pre.ReturnType==2{
		if pre.Info.FSID==0{return provider.UploadResult{},errors.New("TeraBox reported an existing completed file without a valid file ID")}
		result.File=provider.File{ID:strconv.FormatUint(pre.Info.FSID,10),Name:pathBase(pre.Info.Path),Path:pre.Info.Path,Size:int64(pre.Info.Size),IsDirectory:pre.Info.IsDir==1}
		transferred=stat.Size()
		return result,nil
	}
	if pre.UploadID==""{return provider.UploadResult{},errors.New("TeraBox precreate response omitted upload ID")}
	if repo!=nil{if err=repo.SaveProviderSession(ctx,taskID,pre.UploadID,nil);err!=nil{return provider.UploadResult{},fmt.Errorf("save TeraBox upload session: %w",err)}}
	missing:=pre.BlockList;if len(missing)==0{missing=[]int{0}}
	missingSet:=make(map[int]struct{},len(missing));var remaining int64
	for _,index:=range missing{if index<0||index>=len(chunks){return provider.UploadResult{},fmt.Errorf("TeraBox requested invalid part index %d",index)};if _,ok:=missingSet[index];ok{return provider.UploadResult{},fmt.Errorf("TeraBox returned duplicate part index %d",index)};missingSet[index]=struct{}{};remaining+=chunks[index].Size}
	transferred=stat.Size()-remaining
	for _,part:=range chunks{if _,isMissing:=missingSet[part.Index];!isMissing&&repo!=nil{if err=repo.SavePart(ctx,transfer.Part{TransferID:taskID,Index:part.Index,Offset:part.Offset,Size:part.Size,Checksum:blockMD5[part.Index],Status:"completed"});err!=nil{return provider.UploadResult{},fmt.Errorf("persist completed TeraBox part %d: %w",part.Index,err)}}}
	if repo!=nil{if err=repo.UpdateProgress(ctx,taskID,transferred,transfer.StatusRunning,"");err!=nil{return provider.UploadResult{},err}}
	if options.OnProgress!=nil{options.OnProgress(transferred,stat.Size())}
	for _,index:=range missing{if err=ctx.Err();err!=nil{return provider.UploadResult{},err}
		part:=chunks[index];got,completedFile,partErr:=c.uploadPartWithRecovery(ctx,file,remotePath,pre.UploadID,part,blockMD5[index],string(blockJSON),len(chunks),options);if partErr!=nil{return provider.UploadResult{},fmt.Errorf("upload part %d: %w",index,partErr)}
		if completedFile!=nil{result.File=*completedFile;transferred=stat.Size();if repo!=nil{for _,done:=range chunks{if err=repo.SavePart(ctx,transfer.Part{TransferID:taskID,Index:done.Index,Offset:done.Offset,Size:done.Size,Checksum:blockMD5[done.Index],Status:"completed"});err!=nil{return provider.UploadResult{},err}}};return result,nil}
		if !strings.EqualFold(got,blockMD5[index]){return provider.UploadResult{},fmt.Errorf("TeraBox part %d checksum mismatch",index)}
		transferred+=part.Size
		if repo!=nil{
			if err=repo.SavePart(ctx,transfer.Part{TransferID:taskID,Index:index,Offset:part.Offset,Size:part.Size,Checksum:got,RemotePartID:strconv.Itoa(index),Status:"completed"});err!=nil{return provider.UploadResult{},fmt.Errorf("persist TeraBox part %d: %w",index,err)}
			if err=repo.UpdateProgress(ctx,taskID,transferred,transfer.StatusRunning,"");err!=nil{return provider.UploadResult{},fmt.Errorf("persist TeraBox progress: %w",err)}
		}
		if options.OnProgress!=nil{options.OnProgress(transferred,stat.Size())}
	}
	if err=ensureUploadSourceUnchanged(localPath,file,stat);err!=nil{return provider.UploadResult{},err}
	if repo!=nil{if err=repo.UpdateProgress(ctx,taskID,transferred,transfer.StatusFinalizing,"");err!=nil{return provider.UploadResult{},err}}
	rtype:=1;if options.RenamePolicy!=nil{rtype=*options.RenamePolicy}
	created,createErr:=c.createFile(ctx,remotePath,stat.Size(),pre.UploadID,string(blockJSON),rtype);if createErr!=nil{return provider.UploadResult{},createErr}
	if created.FSID==0{return provider.UploadResult{},errors.New("TeraBox create response omitted file ID")}
	name:=created.Name;if name==""{name=pathBase(created.Path)}
	result.File=provider.File{ID:strconv.FormatUint(created.FSID,10),Name:name,Path:created.Path,Size:int64(created.Size),IsDirectory:created.IsDir==1}
	return result,nil
}
func newTransferID()(string,error){var b [16]byte;if _,err:=rand.Read(b[:]);err!=nil{return "",fmt.Errorf("generate transfer ID: %w",err)};return hex.EncodeToString(b[:]),nil}
func ensureUploadSourceUnchanged(path string,file *os.File,initial os.FileInfo)error{
	current,err:=file.Stat();if err!=nil{return fmt.Errorf("recheck upload source: %w",err)}
	pathInfo,err:=os.Stat(path);if err!=nil{return fmt.Errorf("recheck upload source path: %w",err)}
	if !os.SameFile(initial,pathInfo)||current.Size()!=initial.Size()||!current.ModTime().Equal(initial.ModTime())||pathInfo.Size()!=initial.Size()||!pathInfo.ModTime().Equal(initial.ModTime()){return errors.New("upload source changed during transfer; refusing to finalize")}
	return nil
}

type partHTTPError struct { StatusCode int }
func(e *partHTTPError)Error()string{return fmt.Sprintf("TeraBox part upload returned HTTP %d",e.StatusCode)}

func retryablePartError(err error)bool{
	var status *partHTTPError
	if errors.As(err,&status){return status.StatusCode==http.StatusRequestTimeout||status.StatusCode==http.StatusTooManyRequests||status.StatusCode>=500}
	var network net.Error
	return errors.As(err,&network)||errors.Is(err,io.ErrUnexpectedEOF)
}

// uploadPartWithRecovery checks the authoritative precreate missing-part list
// after an ambiguous transient failure. It retries only when TeraBox still
// reports that exact shard as missing.
func(c *APIClient)uploadPartWithRecovery(ctx context.Context,file *os.File,remotePath,uploadID string,part uploadPart,expectedMD5,blockJSON string,totalParts int,options UploadOptions)(string,*provider.File,error){
	attempts:=options.MaxPartAttempts;if attempts==0{attempts=3};if attempts<1{return "",nil,errors.New("max part attempts must be positive")}
	delay:=options.RetryDelay;if delay==0{delay=400*time.Millisecond};maxDelay:=options.MaxRetryDelay;if maxDelay==0{maxDelay=5*time.Second}
	if delay<0||maxDelay<0||maxDelay<delay{return "",nil,errors.New("invalid part retry delay configuration")}
	for attempt:=1;attempt<=attempts;attempt++{
		got,err:=c.uploadPart(ctx,file,remotePath,uploadID,part)
		if err==nil{return got,nil,nil}
		if !retryablePartError(err)||attempt==attempts{return "",nil,err}
		state,stateErr:=c.precreate(ctx,remotePath,blockJSON)
		if stateErr!=nil{return "",nil,fmt.Errorf("part upload failed (%v) and TeraBox resume check failed: %w",err,stateErr)}
		if state.ReturnType==2{
			if state.Info.FSID==0{return "",nil,errors.New("TeraBox reported completed file without file ID")}
			name:=state.Info.Name;if name==""{name=pathBase(state.Info.Path)}
			done:=&provider.File{ID:strconv.FormatUint(state.Info.FSID,10),Name:name,Path:state.Info.Path,Size:int64(state.Info.Size),IsDirectory:state.Info.IsDir==1}
			return "",done,nil
		}
		if state.UploadID!=uploadID{return "",nil,errors.New("TeraBox upload session changed during resume check")}
		missing:=state.BlockList;if len(missing)==0{missing=[]int{0}}
		missingSet:=make(map[int]struct{},len(missing))
		for _,index:=range missing{if index<0||index>=totalParts{return "",nil,fmt.Errorf("TeraBox resume check returned invalid part index %d",index)};if _,exists:=missingSet[index];exists{return "",nil,fmt.Errorf("TeraBox resume check returned duplicate part index %d",index)};missingSet[index]=struct{}{}}
		if _,stillMissing:=missingSet[part.Index];!stillMissing{return expectedMD5,nil,nil}
		if attempt<attempts{
			timer:=time.NewTimer(delay)
			select{case<-ctx.Done():if !timer.Stop(){<-timer.C};return "",nil,ctx.Err();case<-timer.C:}
			delay*=2;if delay>maxDelay{delay=maxDelay}
		}
	}
	return "",nil,errors.New("TeraBox part upload attempts exhausted")
}

func(c *APIClient)precreate(ctx context.Context,remotePath,blockJSON string)(precreateResponse,error){
	q:=url.Values{"access_tokens":{c.accessToken}};form:=url.Values{"path":{remotePath},"autoinit":{"1"},"block_list":{blockJSON}}
	var result precreateResponse;if err:=c.postFormJSON(ctx,c.apiBase+"/openapi/api/precreate?"+q.Encode(),form,&result);err!=nil{return result,err};return result,nil
}
func(c *APIClient)createFile(ctx context.Context,remotePath string,size int64,uploadID,blockJSON string,rtype int)(createResponse,error){
	if rtype<0||rtype>3{return createResponse{},errors.New("rename policy must be between 0 and 3")}
	q:=url.Values{"access_tokens":{c.accessToken}};form:=url.Values{"path":{remotePath},"size":{strconv.FormatInt(size,10)},"uploadid":{uploadID},"block_list":{blockJSON},"rtype":{strconv.Itoa(rtype)}}
	var result createResponse;if err:=c.postFormJSON(ctx,c.apiBase+"/openapi/api/create?"+q.Encode(),form,&result);err!=nil{return result,err};return result,nil
}
func(c *APIClient)uploadPart(ctx context.Context,file *os.File,remotePath,uploadID string,part uploadPart)(string,error){
	var header bytes.Buffer;writer:=multipart.NewWriter(&header);if _,err:=writer.CreateFormFile("file",filepath.Base(file.Name()));err!=nil{return "",err};prefixLen:=header.Len()
	if err:=writer.Close();err!=nil{return "",err};all:=header.Bytes();prefix:=append([]byte(nil),all[:prefixLen]...);suffix:=append([]byte(nil),all[prefixLen:]...)
	body:=io.MultiReader(bytes.NewReader(prefix),io.NewSectionReader(file,part.Offset,part.Size),bytes.NewReader(suffix))
	q:=url.Values{"method":{"upload"},"app_id":{teraboxAppID},"path":{remotePath},"uploadid":{uploadID},"partseq":{strconv.Itoa(part.Index)},"access_tokens":{c.accessToken}}
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,c.uploadBase+"/rest/2.0/pcs/superfile2?"+q.Encode(),body);if err!=nil{return "",err}
	req.Header.Set("Content-Type",writer.FormDataContentType());req.ContentLength=int64(len(prefix)+len(suffix))+part.Size
	resp,err:=c.http.Do(req);if err!=nil{return "",fmt.Errorf("TeraBox part request failed: %w",err)};defer resp.Body.Close()
	const max=1<<20;data,err:=io.ReadAll(io.LimitReader(resp.Body,max+1));if err!=nil{return "",err};if len(data)>max{return "",errors.New("TeraBox part response exceeded 1 MiB")}
	if resp.StatusCode<200||resp.StatusCode>=300{return "",&partHTTPError{StatusCode:resp.StatusCode}}
	var result partResponse;if err:=json.Unmarshal(data,&result);err!=nil{return "",fmt.Errorf("decode TeraBox part response: %w",err)}
	if result.Errno!=0{return "",fmt.Errorf("TeraBox part upload error %d: %s",result.Errno,result.ShowMsg)}
	if result.UploadID!=uploadID{return "",errors.New("TeraBox part response upload ID mismatch")};if result.PartSeq!=part.Index{return "",errors.New("TeraBox part response sequence mismatch")};if result.MD5==""{return "",errors.New("TeraBox part response omitted MD5")};return result.MD5,nil
}
func(c *APIClient)postFormJSON(ctx context.Context,endpoint string,form url.Values,dst any)error{
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,endpoint,strings.NewReader(form.Encode()));if err!=nil{return err};req.Header.Set("Content-Type","application/x-www-form-urlencoded");return c.getJSON(req,dst)
}
