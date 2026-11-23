//go:build !windows

package virtualdrive

import "context"

type Manager struct{}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) Mount(context.Context, APIClientFactory, string, string, string) (Status, error) {
	return Status{}, ErrUnsupported
}

func (m *Manager) Unmount() Status {
	return Status{Mounted: false, Message: "Virtual-drive mounting is supported on Windows only."}
}

func (m *Manager) GetStatus() Status {
	return Status{Mounted: false, Message: "Virtual-drive mounting is supported on Windows only."}
}
