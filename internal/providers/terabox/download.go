package terabox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type downloadLinkResponse struct {
	Errno int `json:"errno"`
	ShowMsg string `json:"show_msg"`
	Links []struct { FSID string `json:"fs_id"`; URL string `json:"dlink"` } `json:"dlink"`
}

// DownloadURL asks the official API for a temporary link and attaches the
// access token as required by the Open Platform documentation.
func(c *APIClient)DownloadURL(ctx context.Context,fileID string)(string,error){
	id,err:=strconv.ParseUint(fileID,10,64);if err!=nil||id==0{return "",errors.New("TeraBox file ID must be a positive integer")}
	q:=url.Values{"access_tokens":{c.accessToken},"fidlist":{"["+strconv.FormatUint(id,10)+"]"},"type":{"dlink"}}
	req,err:=http.NewRequestWithContext(ctx,http.MethodGet,c.apiBase+"/openapi/api/download?"+q.Encode(),nil);if err!=nil{return "",err}
	var result downloadLinkResponse
	if err:=c.getJSON(req,&result);err!=nil{return "",err}
	for _,item:=range result.Links{
		if item.FSID!=fileID{continue}
		u,err:=url.Parse(item.URL);if err!=nil||!c.isTrustedDownloadURL(u){continue}
		values:=u.Query();values.Set("access_tokens",c.accessToken);u.RawQuery=values.Encode()
		return u.String(),nil
	}
	return "",errors.New("TeraBox did not return a valid download link for this file")
}
func(c *APIClient)isTrustedDownloadURL(u *url.URL)bool{
	if u==nil||u.User!=nil||u.Host==""{return false}
	if u.Scheme=="https"&&(u.Hostname()=="terabox.com"||strings.HasSuffix(u.Hostname(),".terabox.com")||u.Hostname()=="terabox.app"||strings.HasSuffix(u.Hostname(),".terabox.app")){return true}
	return u.Scheme=="http"&&(u.Hostname()=="127.0.0.1"||u.Hostname()=="localhost"||u.Hostname()=="::1")
}
// Download returns a streaming response body. The caller must close it.
// The request context controls the lifetime so large files are not cut off
// by a fixed whole-response timeout.
func(c *APIClient)Download(ctx context.Context,fileID string)(io.ReadCloser,error){
	link,err:=c.DownloadURL(ctx,fileID);if err!=nil{return nil,err}
	req,err:=http.NewRequestWithContext(ctx,http.MethodGet,link,nil);if err!=nil{return nil,err}
	resp,err:=c.downloadHTTP.Do(req);if err!=nil{return nil,fmt.Errorf("TeraBox download request failed: %w",err)}
	if resp.StatusCode<200||resp.StatusCode>=300{resp.Body.Close();return nil,fmt.Errorf("TeraBox download returned HTTP %d",resp.StatusCode)}
	return resp.Body,nil
}
