package terabox

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
)

const (
	DefaultUploadChunkSize int64 = 8 << 20
	minMultiPartSize int64 = 4 << 20
	teraboxAppID = "250528"
)
type UploadOptions struct { ChunkSize int64; RenamePolicy *int; OnProgress func(transferred,total int64) }
type precreateResponse struct { Errno int `json:"errno"`;ShowMsg string `json:"show_msg"`;UploadID string `json:"uploadid"`;ReturnType int `json:"return_type"`;BlockList []int `json:"block_list"`;Info apiFile `json:"info"` }
type partResponse struct { Errno int `json:"errno"`;ShowMsg string `json:"show_msg"`;MD5 string `json:"md5"`;UploadID string `json:"uploadid"`;PartSeq int `json:"partseq"` }
type createResponse struct { Errno int `json:"errno"`;ShowMsg string `json:"show_msg"`;FSID uint64 `json:"fs_id"`;Path string `json:"path"`;Name string `json:"server_filename"`;Size uint64 `json:"size"`;IsDir int `json:"isdir"`;Modified uint64 `json:"mtime"` }

func PlanTeraBoxChunks(fileSize,preferredSize int64)([]Chunk,error){
	if fileSize<0{return nil,errors.New("file size cannot be negative")}
	if preferredSize<=minMultiPartSize{return nil,errors.New("preferred chunk size must exceed 4 MiB")}
	if fileSize==0{return nil,errors.New("empty-file upload is not yet supported")}
	count:=1+(fileSize-1)/preferredSize
	for count>1&&fileSize/count<=minMultiPartSize{count--}
	if count>MaxPlannedChunks{return nil,fmt.Errorf("upload requires too many parts (maximum %d)",MaxPlannedChunks)}
	base,extra:=fileSize/count,fileSize%count
	chunks:=make([]Chunk,0,int(count));var offset int64
	for i:=int64(0);i<count;i++{size:=base;if i<extra{size++};if count>1&&size<=minMultiPartSize{return nil,errors.New("unable to satisfy TeraBox multi-part minimum size")};chunks=append(chunks,Chunk{Index:int(i),Offset:offset,Size:size});offset+=size}
	return chunks,nil
}

// UploadFile streams shards from disk, verifies each returned MD5 and finalizes
// only after all missing shard indexes from precreate have been uploaded.
func(c *APIClient)UploadFile(ctx context.Context,localPath,remotePath string,options UploadOptions)(provider.UploadResult,error){
	if err:=c.validateRemotePath(remotePath);err!=nil{return provider.UploadResult{},err}
	chunkSize:=options.ChunkSize;if chunkSize==0{chunkSize=DefaultUploadChunkSize}
	file,err:=os.Open(localPath);if err!=nil{return provider.UploadResult{},fmt.Errorf("open upload source: %w",err)};defer file.Close()
	stat,err:=file.Stat();if err!=nil{return provider.UploadResult{},fmt.Errorf("stat upload source: %w",err)}
	if !stat.Mode().IsRegular(){return provider.UploadResult{},errors.New("upload source must be a regular file")}
	chunks,err:=PlanTeraBoxChunks(stat.Size(),chunkSize);if err!=nil{return provider.UploadResult{},err}
	blockMD5:=make([]string,len(chunks))
	for i,ch:=range chunks{if err:=ctx.Err();err!=nil{return provider.UploadResult{},err};h:=md5.New();if _,err:=io.Copy(h,io.NewSectionReader(file,ch.Offset,ch.Size));err!=nil{return provider.UploadResult{},fmt.Errorf("hash upload part %d: %w",i,err)};blockMD5[i]=hex.EncodeToString(h.Sum(nil))}
	blockJSON,_:=json.Marshal(blockMD5)
	pre,err:=c.precreate(ctx,remotePath,string(blockJSON));if err!=nil{return provider.UploadResult{},err}
	if pre.ReturnType==2{return provider.UploadResult{File:provider.File{ID:strconv.FormatUint(pre.Info.FSID,10),Name:pathBase(pre.Info.Path),Path:pre.Info.Path,Size:int64(pre.Info.Size),IsDirectory:pre.Info.IsDir==1}},nil}
	if pre.UploadID==""{return provider.UploadResult{},errors.New("TeraBox precreate response omitted upload ID")}
	missing:=pre.BlockList;if len(missing)==0{missing=[]int{0}}
	missingSet:=make(map[int]struct{},len(missing));var remaining int64
	for _,index:=range missing{if index<0||index>=len(chunks){return provider.UploadResult{},fmt.Errorf("TeraBox requested invalid part index %d",index)};if _,ok:=missingSet[index];ok{return provider.UploadResult{},fmt.Errorf("TeraBox returned duplicate part index %d",index)};missingSet[index]=struct{}{};remaining+=chunks[index].Size}
	transferred:=stat.Size()-remaining
	if options.OnProgress!=nil{options.OnProgress(transferred,stat.Size())}
	for _,index:=range missing{if err:=ctx.Err();err!=nil{return provider.UploadResult{},err}
		part:=chunks[index];got,err:=c.uploadPart(ctx,file,remotePath,pre.UploadID,part);if err!=nil{return provider.UploadResult{},fmt.Errorf("upload part %d: %w",index,err)}
		if !strings.EqualFold(got,blockMD5[index]){return provider.UploadResult{},fmt.Errorf("TeraBox part %d checksum mismatch",index)}
		transferred+=part.Size;if options.OnProgress!=nil{options.OnProgress(transferred,stat.Size())}
	}
	rtype:=1 // documented default: rename when a path conflict exists
	if options.RenamePolicy!=nil{rtype=*options.RenamePolicy}
	created,err:=c.createFile(ctx,remotePath,stat.Size(),pre.UploadID,string(blockJSON),rtype);if err!=nil{return provider.UploadResult{},err}
	if created.FSID==0{return provider.UploadResult{},errors.New("TeraBox create response omitted file ID")}
	name:=created.Name;if name==""{name=pathBase(created.Path)}
	return provider.UploadResult{File:provider.File{ID:strconv.FormatUint(created.FSID,10),Name:name,Path:created.Path,Size:int64(created.Size),IsDirectory:created.IsDir==1}},nil
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
func(c *APIClient)uploadPart(ctx context.Context,file *os.File,remotePath,uploadID string,part Chunk)(string,error){
	var header bytes.Buffer;writer:=multipart.NewWriter(&header);if _,err:=writer.CreateFormFile("file",filepath.Base(file.Name()));err!=nil{return "",err};prefixLen:=header.Len()
	if err:=writer.Close();err!=nil{return "",err};all:=header.Bytes();prefix:=append([]byte(nil),all[:prefixLen]...);suffix:=append([]byte(nil),all[prefixLen:]...)
	body:=io.MultiReader(bytes.NewReader(prefix),io.NewSectionReader(file,part.Offset,part.Size),bytes.NewReader(suffix))
	q:=url.Values{"method":{"upload"},"app_id":{teraboxAppID},"path":{remotePath},"uploadid":{uploadID},"partseq":{strconv.Itoa(part.Index)},"access_tokens":{c.accessToken}}
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,c.uploadBase+"/rest/2.0/pcs/superfile2?"+q.Encode(),body);if err!=nil{return "",err}
	req.Header.Set("Content-Type",writer.FormDataContentType());req.ContentLength=int64(len(prefix)+len(suffix))+part.Size
	resp,err:=c.http.Do(req);if err!=nil{return "",fmt.Errorf("TeraBox part request failed: %w",err)};defer resp.Body.Close()
	const max=1<<20;data,err:=io.ReadAll(io.LimitReader(resp.Body,max+1));if err!=nil{return "",err};if len(data)>max{return "",errors.New("TeraBox part response exceeded 1 MiB")}
	if resp.StatusCode<200||resp.StatusCode>=300{return "",fmt.Errorf("TeraBox part upload returned HTTP %d",resp.StatusCode)}
	var result partResponse;if err:=json.Unmarshal(data,&result);err!=nil{return "",fmt.Errorf("decode TeraBox part response: %w",err)}
	if result.Errno!=0{return "",fmt.Errorf("TeraBox part upload error %d: %s",result.Errno,result.ShowMsg)}
	if result.MD5==""{return "",errors.New("TeraBox part response omitted MD5")};return result.MD5,nil
}
func(c *APIClient)postFormJSON(ctx context.Context,endpoint string,form url.Values,dst any)error{
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,endpoint,strings.NewReader(form.Encode()));if err!=nil{return err};req.Header.Set("Content-Type","application/x-www-form-urlencoded");return c.getJSON(req,dst)
}
