package terabox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type fileManagerItem struct {
	Path string `json:"path"`
	Dest string `json:"dest,omitempty"`
	NewName string `json:"newname,omitempty"`
}
type fileManagerInfo struct {
	Errno *int `json:"errno"`
	Path string `json:"path"`
	ShowMsg string `json:"show_msg"`
}
type fileManagerResponse struct {
	Errno int `json:"errno"`
	ShowMsg string `json:"show_msg"`
	Info []fileManagerInfo `json:"info"`
	TaskID uint64 `json:"taskid"`
}

// Rename changes a cloud item's name without moving it to another directory.
func (c *APIClient) Rename(ctx context.Context, remotePath, newName string) error {
	if err := c.validateRemotePath(remotePath); err != nil { return err }
	if c.isApplicationRoot(remotePath) { return errors.New("refusing to rename the TeraBox application root") }
	if err := validateCloudName(newName); err != nil { return err }
	return c.fileManager(ctx, "rename", []fileManagerItem{{Path: remotePath, NewName: newName}})
}

// DeletePath deletes a cloud item by its absolute application-folder path.
func (c *APIClient) DeletePath(ctx context.Context, remotePath string) error {
	if err := c.validateRemotePath(remotePath); err != nil { return err }
	if c.isApplicationRoot(remotePath) { return errors.New("refusing to delete the TeraBox application root") }
	return c.fileManager(ctx, "delete", []string{remotePath})
}

func (c *APIClient) Move(ctx context.Context, remotePath, destinationDir, newName string) error {
	if err := c.validateRemotePath(remotePath); err != nil { return err }
	if c.isApplicationRoot(remotePath) { return errors.New("refusing to move the TeraBox application root") }
	if err := c.validateRemotePath(destinationDir); err != nil { return err }
	if newName != "" {
		if err := validateCloudName(newName); err != nil { return err }
	}
	return c.fileManager(ctx, "move", []fileManagerItem{{Path: remotePath, Dest: destinationDir, NewName: newName}})
}

func (c *APIClient) Copy(ctx context.Context, remotePath, destinationDir, newName string) error {
	if err := c.validateRemotePath(remotePath); err != nil { return err }
	if c.isApplicationRoot(remotePath) { return errors.New("refusing to copy the TeraBox application root") }
	if err := c.validateRemotePath(destinationDir); err != nil { return err }
	if newName != "" {
		if err := validateCloudName(newName); err != nil { return err }
	}
	return c.fileManager(ctx, "copy", []fileManagerItem{{Path: remotePath, Dest: destinationDir, NewName: newName}})
}

func (c *APIClient) isApplicationRoot(remotePath string) bool {
	clean, err := normalizeAbsoluteCloudPath(remotePath)
	return err == nil && c.applicationRoot != "" && clean == c.applicationRoot
}

func validateCloudName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00\r\n") {
		return errors.New("name must be a non-empty filename without path separators")
	}
	return nil
}

func (c *APIClient) fileManager(ctx context.Context, operation string, fileList any) error {
	switch operation {
	case "copy", "move", "rename", "delete":
	default:
		return fmt.Errorf("unsupported TeraBox file operation %q", operation)
	}
	encoded, err := json.Marshal(fileList)
	if err != nil { return fmt.Errorf("encode TeraBox file operation: %w", err) }
	q := url.Values{"access_tokens": {c.accessToken}, "opera": {operation}, "async": {"0"}}
	form := url.Values{"filelist": {string(encoded)}}
	var result fileManagerResponse
	if err := c.postFormJSON(ctx, c.apiBase+"/openapi/api/filemanager?"+q.Encode(), form, &result); err != nil { return err }
	if result.Errno != 0 {
		message := strings.TrimSpace(result.ShowMsg)
		if message == "" { message = "operation rejected" }
		return fmt.Errorf("TeraBox %s error %d: %s", operation, result.Errno, message)
	}
	if len(result.Info) == 0 {
		return errors.New("TeraBox file operation returned no per-item result; completion is unverified")
	}
	for _, item := range result.Info {
		if item.Errno == nil {
			return errors.New("TeraBox file operation omitted per-item status; completion is unverified")
		}
		if *item.Errno != 0 {
			message := strings.TrimSpace(item.ShowMsg)
			if message == "" { message = "operation failed" }
			return fmt.Errorf("TeraBox %s error %d for %s: %s", operation, *item.Errno, item.Path, message)
		}
		if strings.TrimSpace(item.Path) == "" {
			return errors.New("TeraBox file operation omitted the affected path; completion is unverified")
		}
	}
	return nil
}
