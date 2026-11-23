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


func TestValidateAccountDisconnect(t *testing.T) {
	tests := []struct {
		name string
		status Status
		mountedAccountID string
		accountID string
		wantErr bool
	}{
		{name: "mounted account cannot disconnect", status: Status{Mounted: true}, mountedAccountID: "terabox:1", accountID: "terabox:1", wantErr: true},
		{name: "unknown mounted account fails closed", status: Status{Mounted: true}, accountID: "terabox:1", wantErr: true},
		{name: "different account may disconnect", status: Status{Mounted: true}, mountedAccountID: "terabox:1", accountID: "terabox:2"},
		{name: "unmounted account may disconnect", accountID: "terabox:1"},
		{name: "empty account ID rejected", status: Status{Mounted: true}, mountedAccountID: "terabox:1", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAccountDisconnect(test.status, test.mountedAccountID, test.accountID)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateAccountDisconnect() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
