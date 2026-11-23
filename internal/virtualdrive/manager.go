package virtualdrive

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/NithishKumar2772/TBox-Drive/internal/providers/terabox"
)

var ErrUnsupported = errors.New("Windows virtual drives require Windows and WinFsp")
var ErrFolderCreateUnsupported = errors.New("TeraBox Open Platform does not currently document folder creation for this application")

type APIClientFactory func(context.Context) (*terabox.APIClient, error)

type Status struct {
	Mounted   bool   `json:"mounted"`
	Mountpoint string `json:"mountpoint,omitempty"`
	Provider  string `json:"provider,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Message   string `json:"message"`
}

// NormalizeVirtualPath keeps all virtual-drive paths rooted and prevents
// path traversal before a path is mapped into the TeraBox application folder.
func NormalizeVirtualPath(name string) (string, error) {
	if name == "" || name == "/" || name == "\\" {
		return "", nil
	}
	name = strings.ReplaceAll(name, "\\", "/")
	if strings.ContainsRune(name, 0) {
		return "", errors.New("path contains a NUL character")
	}
	clean := path.Clean("/" + strings.TrimLeft(name, "/"))
	if clean == "/.." || strings.HasPrefix(clean, "/../") {
		return "", errors.New("path traversal is not allowed")
	}
	return strings.TrimPrefix(clean, "/"), nil
}
