//go:build windows

package virtualdrive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
	"github.com/winfsp/go-winfsp"
	"github.com/winfsp/go-winfsp/gofs"
)

type Manager struct {
	mu sync.Mutex
	mount *winfsp.FileSystem
	filesystem *cloudFileSystem
	status Status
}

func NewManager() *Manager {
	return &Manager{status: Status{Message: "Virtual drive is not mounted."}}
}

func (m *Manager) Mount(ctx context.Context, factory APIClientFactory, applicationRoot, cacheDir, mountpoint string) (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mount != nil {
		return m.status, errors.New("TBox Drive virtual drive is already mounted")
	}
	mountpoint = strings.ToUpper(strings.TrimSpace(mountpoint))
	if len(mountpoint) != 2 || mountpoint[1] != ':' || mountpoint[0] < 'A' || mountpoint[0] > 'Z' {
		return m.status, errors.New("mount point must be a drive letter such as T:")
	}
	if factory == nil || strings.TrimSpace(applicationRoot) == "" {
		return m.status, errors.New("TeraBox account and application folder are required")
	}
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return m.status, fmt.Errorf("create virtual-drive cache: %w", err)
	}
	fs := &cloudFileSystem{
		ctx: ctx, api: factory, root: strings.TrimRight(applicationRoot, "/"),
		cacheDir: cacheDir,
	}
	mounted, err := winfsp.Mount(gofs.New(fs), mountpoint)
	if err != nil {
		return m.status, fmt.Errorf("mount %s failed; install the WinFsp runtime and try again: %w", mountpoint, err)
	}
	m.mount, m.filesystem = mounted, fs
	m.status = Status{Mounted: true, Mountpoint: mountpoint, Provider: "TeraBox", Mode: "Hybrid prototype", Message: "TeraBox virtual drive mounted. Files are downloaded to the local cache when opened; edited files upload when closed."}
	return m.status, nil
}

func (m *Manager) Unmount() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mount != nil {
		m.mount.Unmount()
		m.mount = nil
		m.filesystem = nil
	}
	m.status = Status{Mounted: false, Message: "Virtual drive is unmounted."}
	return m.status
}

func (m *Manager) GetStatus() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

type cloudFileSystem struct {
	ctx context.Context
	api APIClientFactory
	root string
	cacheDir string
}

func (fs *cloudFileSystem) apiClient() (*terabox.APIClient, error) {
	return fs.api(fs.ctx)
}

func (fs *cloudFileSystem) remotePath(name string) (string, error) {
	rel, err := NormalizeVirtualPath(name)
	if err != nil { return "", err }
	if rel == "" { return fs.root + "/", nil }
	return path.Join(fs.root, rel), nil
}

func (fs *cloudFileSystem) list(remoteDir string) ([]provider.File, error) {
	api, err := fs.apiClient()
	if err != nil { return nil, err }
	var all []provider.File
	for page := 1; page <= 1000; page++ {
		items, more, err := api.ListDirectory(fs.ctx, remoteDir, page, 1000)
		if err != nil { return nil, err }
		all = append(all, items...)
		if !more { return all, nil }
	}
	return nil, errors.New("TeraBox directory exceeded the 1000-page safety limit")
}

func (fs *cloudFileSystem) lookup(name string) (provider.File, error) {
	remote, err := fs.remotePath(name)
	if err != nil { return provider.File{}, err }
	if remote == fs.root + "/" {
		return provider.File{ID: "root", Name: "", Path: remote, IsDirectory: true, ModifiedAt: time.Now()}, nil
	}
	parent := path.Dir(remote)
	items, err := fs.list(parent)
	if err != nil { return provider.File{}, err }
	for _, item := range items {
		if strings.EqualFold(strings.TrimRight(item.Path, "/"), strings.TrimRight(remote, "/")) {
			return item, nil
		}
	}
	return provider.File{}, os.ErrNotExist
}

func (fs *cloudFileSystem) Stat(name string) (os.FileInfo, error) {
	item, err := fs.lookup(name)
	if err != nil { return nil, err }
	return cloudInfo{file: item}, nil
}

func (fs *cloudFileSystem) Mkdir(name string, perm os.FileMode) error {
	return ErrFolderCreateUnsupported
}

func (fs *cloudFileSystem) Remove(name string) error {
	remote, err := fs.remotePath(name)
	if err != nil { return err }
	if remote == fs.root + "/" { return errors.New("the TeraBox application root cannot be deleted") }
	api, err := fs.apiClient()
	if err != nil { return err }
	return api.DeletePath(fs.ctx, remote)
}

func (fs *cloudFileSystem) Rename(source, target string) error {
	src, err := fs.remotePath(source)
	if err != nil { return err }
	dst, err := fs.remotePath(target)
	if err != nil { return err }
	if src == fs.root + "/" || dst == fs.root + "/" { return errors.New("the TeraBox application root cannot be renamed") }
	api, err := fs.apiClient()
	if err != nil { return err }
	if path.Dir(src) == path.Dir(dst) {
		return api.Rename(fs.ctx, src, path.Base(dst))
	}
	return api.Move(fs.ctx, src, path.Dir(dst), path.Base(dst))
}

func (fs *cloudFileSystem) OpenFile(name string, flag int, perm os.FileMode) (gofs.File, error) {
	remote, err := fs.remotePath(name)
	if err != nil { return nil, err }
	item, lookupErr := fs.lookup(name)
	exists := lookupErr == nil
	if lookupErr != nil && !errors.Is(lookupErr, os.ErrNotExist) { return nil, lookupErr }
	writing := flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0
	if exists && item.IsDirectory {
		if writing { return nil, errors.New("cannot open a directory for writing") }
		children, err := fs.list(remote)
		if err != nil { return nil, err }
		infos := make([]os.FileInfo, 0, len(children))
		for _, child := range children { infos = append(infos, cloudInfo{file: child}) }
		return &cloudHandle{info: cloudInfo{file: item}, entries: infos}, nil
	}
	if !exists && flag&os.O_CREATE == 0 { return nil, os.ErrNotExist }
	if !exists && flag&os.O_EXCL == 0 && flag&os.O_CREATE == 0 { return nil, os.ErrNotExist }
	if exists && flag&os.O_CREATE != 0 && flag&os.O_EXCL != 0 { return nil, os.ErrExist }

	var localPath string
	if exists && flag&os.O_TRUNC == 0 {
		localPath, err = fs.ensureCached(item)
		if err != nil { return nil, err }
	} else {
		tmp, createErr := os.CreateTemp(fs.cacheDir, "pending-*")
		if createErr != nil { return nil, createErr }
		localPath = tmp.Name()
		if closeErr := tmp.Close(); closeErr != nil { return nil, closeErr }
	}
	openFlags := flag
	if writing { openFlags |= os.O_CREATE }
	if !exists { if err := os.Remove(localPath); err != nil && !errors.Is(err, os.ErrNotExist) { return nil, err } }
	file, err := os.OpenFile(localPath, openFlags, perm)
	if err != nil { return nil, err }
	info := cloudInfo{file: item}
	if !exists {
		info = cloudInfo{file: provider.File{Name: path.Base(remote), Path: remote, ModifiedAt: time.Now()}}
	}
	return &cloudHandle{file: file, info: info, filesystem: fs, remotePath: remote, dirty: writing}, nil
}

func (fs *cloudFileSystem) ensureCached(item provider.File) (string, error) {
	key := fmt.Sprintf("%s:%d:%d", item.ID, item.Size, item.ModifiedAt.Unix())
	sum := sha256.Sum256([]byte(key))
	local := filepath.Join(fs.cacheDir, hex.EncodeToString(sum[:]))
	if stat, err := os.Stat(local); err == nil && stat.Size() == item.Size { return local, nil }
	api, err := fs.apiClient()
	if err != nil { return "", err }
	body, err := api.Download(fs.ctx, item.ID)
	if err != nil { return "", err }
	defer body.Close()
	tmp, err := os.CreateTemp(fs.cacheDir, "download-*")
	if err != nil { return "", err }
	tmpName := tmp.Name()
	ok := false
	defer func(){ _ = tmp.Close(); if !ok { _ = os.Remove(tmpName) } }()
	written, err := io.Copy(tmp, body)
	if err != nil { return "", fmt.Errorf("cache TeraBox file: %w", err) }
	if item.Size > 0 && written != item.Size { return "", fmt.Errorf("download size mismatch: expected %d bytes, received %d", item.Size, written) }
	if err := tmp.Sync(); err != nil { return "", err }
	if err := tmp.Close(); err != nil { return "", err }
	if err := os.Rename(tmpName, local); err != nil {
		if _, statErr := os.Stat(local); statErr != nil { return "", err }
	}
	ok = true
	return local, nil
}

type cloudInfo struct { file provider.File }
func (i cloudInfo) Name() string { return i.file.Name }
func (i cloudInfo) Size() int64 { return i.file.Size }
func (i cloudInfo) Mode() os.FileMode { if i.file.IsDirectory { return os.ModeDir|0555 }; return 0444|0200 }
func (i cloudInfo) ModTime() time.Time { if i.file.ModifiedAt.IsZero() { return time.Unix(0,0) }; return i.file.ModifiedAt }
func (i cloudInfo) IsDir() bool { return i.file.IsDirectory }
func (i cloudInfo) Sys() any { return nil }

type cloudHandle struct {
	file *os.File
	info cloudInfo
	filesystem *cloudFileSystem
	remotePath string
	dirty bool
	entries []os.FileInfo
	index int
	closed bool
	mu sync.Mutex
}

func (h *cloudHandle) Read(p []byte) (int,error) { if h.file==nil{return 0,errors.New("cannot read a directory")}; return h.file.Read(p) }
func (h *cloudHandle) Write(p []byte) (int,error) { if h.file==nil{return 0,errors.New("cannot write a directory")}; h.dirty=true; return h.file.Write(p) }
func (h *cloudHandle) ReadAt(p []byte, off int64) (int,error) { if h.file==nil{return 0,errors.New("cannot read a directory")}; return h.file.ReadAt(p,off) }
func (h *cloudHandle) WriteAt(p []byte, off int64) (int,error) { if h.file==nil{return 0,errors.New("cannot write a directory")}; h.dirty=true; return h.file.WriteAt(p,off) }
func (h *cloudHandle) Seek(off int64, whence int) (int64,error) { if h.file==nil{return 0,errors.New("cannot seek a directory")}; return h.file.Seek(off,whence) }
func (h *cloudHandle) Stat() (os.FileInfo,error) { if h.file!=nil { return h.file.Stat() }; return h.info,nil }
func (h *cloudHandle) Sync() error { if h.file==nil{return nil}; if err:=h.file.Sync();err!=nil{return err}; return h.uploadIfDirty() }
func (h *cloudHandle) Truncate(size int64) error { if h.file==nil{return errors.New("cannot truncate a directory")}; h.dirty=true; return h.file.Truncate(size) }
func (h *cloudHandle) Readdir(count int) ([]os.FileInfo,error) {
	if h.file!=nil{return nil,errors.New("not a directory")}
	if h.index>=len(h.entries){return nil,io.EOF}
	end:=len(h.entries);if count>0&&end-h.index>count{end=h.index+count}
	result:=append([]os.FileInfo(nil),h.entries[h.index:end]...);h.index=end
	if count>0&&len(result)<count{return result,io.EOF}
	return result,nil
}
func (h *cloudHandle) uploadIfDirty() error {
	if !h.dirty || h.file==nil || h.filesystem==nil{return nil}
	if err:=h.file.Sync();err!=nil{return err}
	api,err:=h.filesystem.apiClient();if err!=nil{return err}
	if _,err:=api.UploadFile(h.filesystem.ctx,h.file.Name(),h.remotePath,terabox.UploadOptions{});err!=nil{return fmt.Errorf("upload edited virtual-drive file: %w",err)}
	h.dirty=false
	return nil
}
func (h *cloudHandle) Close() error {
	h.mu.Lock();defer h.mu.Unlock()
	if h.closed{return nil};h.closed=true
	uploadErr:=h.uploadIfDirty()
	var closeErr error;if h.file!=nil{closeErr=h.file.Close()}
	if uploadErr!=nil{return uploadErr};return closeErr
}

var _ gofs.FileSystem = (*cloudFileSystem)(nil)
var _ gofs.File = (*cloudHandle)(nil)
var _ os.FileInfo = cloudInfo{}
