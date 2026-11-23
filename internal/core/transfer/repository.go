package transfer

import (
	"context"
	"time"
)

// Part is the durable local record for one TeraBox upload/download range.
type Part struct {
	TransferID string
	Index int
	Offset int64
	Size int64
	Checksum string
	RemotePartID string
	Status string
	UpdatedAt time.Time
}

// Repository is the persistence contract used by the TeraBox transfer workflow.
type Repository interface {
	CreateTransfer(context.Context, Task) error
	GetTransfer(context.Context, string) (Task, error)
	UpdateProgress(context.Context, string, int64, Status, string) error
	SavePart(context.Context, Part) error
	SaveProviderSession(context.Context, string, string, *time.Time) error
}
