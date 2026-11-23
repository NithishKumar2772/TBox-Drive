package provider

import (
	"context"
	"io"
	"time"
)

type File struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64 `json:"size"`
	IsDirectory bool `json:"isDirectory"`
	ModifiedAt time.Time `json:"modifiedAt"`
	ETag string `json:"etag,omitempty"`
}
type UploadRequest struct { LocalPath string; ParentID string; Name string; Size int64; TaskID string }
type UploadResult struct { File File `json:"file"`; TaskID string `json:"taskID,omitempty"` }
type Provider interface {
	ID() string
	Name() string
	Capabilities() Capabilities
	List(ctx context.Context,parentID,pageToken string)([]File,string,error)
	Download(ctx context.Context,fileID string)(io.ReadCloser,error)
	Upload(ctx context.Context,req UploadRequest)(UploadResult,error)
	Delete(ctx context.Context,fileID string)error
}
