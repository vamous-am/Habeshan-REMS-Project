package attendance

import (
	"time"

	"github.com/google/uuid"
)

// ClockInRequest defines the expected JSON payload from the frontend client
type ClockInRequest struct {
	RecordUUID uuid.UUID `json:"record_uuid"`
	DeviceHash string    `json:"device_hash"`
	ClockIn    time.Time `json:"clock_in"`
}

// ClockInResponse defines the response payload sent back to the client
type ClockInResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	ClockIn    time.Time  `json:"clock_in"`
	SyncStatus SyncStatus `json:"sync_status"`
	RecordUUID uuid.UUID  `json:"record_uuid"`
}
