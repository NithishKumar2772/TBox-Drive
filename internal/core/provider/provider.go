package provider

import (
	"context"
	"io"
	"time"
)

type File struct {
	ID string
	Name string
	Path string
	Size int64
	IsDirectory bool
	ModifiedAt time.Time
	ETag string
}

type UploadRequest struct { LocalPath, ParentID, Name string; Size int64 }
type UploadResult struct {
	File File
	TaskID string
}

// Provider defines the common provider-neutral operations.
type Provider interface {
	ID() string
	Name() string
	Capabilities() Capabilities
	List(ctx context.Context, parentID, pageToken string) ([]File, string, error)
	Download(ctx context.Context, fileID string) (io.ReadCloser, error)
	Upload(ctx context.Context, req UploadRequest) (UploadResult, error)
	Delete(ctx context.Context, fileID string) error
}
