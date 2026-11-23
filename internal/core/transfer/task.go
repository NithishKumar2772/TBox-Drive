package transfer

import "time"

type Status string
const (
	StatusQueued Status = "queued"
	StatusRunning Status = "running"
	StatusPaused Status = "paused"
	StatusRetrying Status = "retrying"
	StatusFinalizing Status = "finalizing"
	StatusCompleted Status = "completed"
	StatusFailed Status = "failed"
	StatusCanceled Status = "canceled"
)
type Task struct {
	ID string `json:"id"`
	ProviderID string `json:"provider_id"`
	Direction string `json:"direction"`
	Source string `json:"source"`
	Destination string `json:"destination"`
	Size int64 `json:"size"`
	Transferred int64 `json:"transferred"`
	Status Status `json:"status"`
	Error string `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
