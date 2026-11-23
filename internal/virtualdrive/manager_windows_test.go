//go:build windows

package virtualdrive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NithishKumar2772/TBox-Drive/internal/core/provider"
)

func cachePathForTest(dir string, item provider.File) string {
	key := fmt.Sprintf("%s:%d:%d", item.ID, item.Size, item.ModifiedAt.Unix())
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(dir, hex.EncodeToString(sum[:]))
}

func TestEnsureCachedPreservesPossibleUnsyncedEdits(t *testing.T) {
	cacheDir := t.TempDir()
	item := provider.File{ID: "123", Name: "edited.txt", Size: 10, ModifiedAt: time.Unix(100, 0)}
	local := cachePathForTest(cacheDir, item)
	original := []byte("local edit")
	if err := os.WriteFile(local, original, 0600); err != nil {
		t.Fatal(err)
	}

	fs := &cloudFileSystem{cacheDir: cacheDir}
	if _, err := fs.ensureCached(item); err == nil {
		t.Fatal("expected size mismatch to fail closed")
	}
	got, err := os.ReadFile(local)
	if err != nil {
		t.Fatalf("cached edit was deleted: %v", err)
	}
	if string(got) != string(original) {
		t.Fatalf("cached edit changed: got %q, want %q", got, original)
	}
}

func TestEnsureCachedReusesCompleteCacheFile(t *testing.T) {
	cacheDir := t.TempDir()
	item := provider.File{ID: "456", Name: "cached.txt", Size: 6, ModifiedAt: time.Unix(200, 0)}
	local := cachePathForTest(cacheDir, item)
	if err := os.WriteFile(local, []byte("cached"), 0600); err != nil {
		t.Fatal(err)
	}

	fs := &cloudFileSystem{cacheDir: cacheDir}
	got, err := fs.ensureCached(item)
	if err != nil {
		t.Fatalf("complete cache should be reused without provider access: %v", err)
	}
	if got != local {
		t.Fatalf("cache path = %q, want %q", got, local)
	}
}
