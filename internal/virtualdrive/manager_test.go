package virtualdrive

import "testing"

func TestNormalizeVirtualPath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "", want: ""},
		{name: "/", want: ""},
		{name: "\\", want: ""},
		{name: "Pictures/Family/photo.jpg", want: "Pictures/Family/photo.jpg"},
		{name: "/Documents/report.pdf", want: "Documents/report.pdf"},
		{name: "\\Pictures\\photo.jpg", want: "Pictures/photo.jpg"},
	}
	for _, test := range tests {
		got, err := NormalizeVirtualPath(test.name)
		if err != nil { t.Fatalf("NormalizeVirtualPath(%q): %v", test.name, err) }
		if got != test.want { t.Errorf("NormalizeVirtualPath(%q) = %q, want %q", test.name, got, test.want) }
	}
}

func TestNormalizeVirtualPathRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../secret.txt", "/folder/../../secret.txt", "folder/../secret.txt", "folder\\..\\secret.txt", "bad\x00name"} {
		if _, err := NormalizeVirtualPath(name); err == nil {
			t.Errorf("NormalizeVirtualPath(%q) accepted an unsafe path", name)
		}
	}
}
