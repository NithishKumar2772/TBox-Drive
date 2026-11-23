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
	for _, segment := range strings.Split(name, "/") {
		if segment == ".." { return "", errors.New("path traversal is not allowed") }
	}
	clean := path.Clean("/" + strings.TrimLeft(name, "/"))
	return strings.TrimPrefix(clean, "/"), nil
}


// ValidateAccountDisconnect prevents removing credentials for an account whose
// virtual drive is still mounted. If the mounted account is unknown, fail closed.
func ValidateAccountDisconnect(status Status, mountedAccountID, accountID string) error {
	if strings.TrimSpace(accountID) == "" {
		return errors.New("account ID is required")
	}
	if status.Mounted && (mountedAccountID == "" || mountedAccountID == accountID) {
		return errors.New("unmount the TeraBox virtual drive before disconnecting this account")
	}
	return nil
}
