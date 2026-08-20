package attendance

import (
	"time"

	"github.com/google/uuid"
)

// ClockInRequest payload for real-time online clock-in
type ClockInRequest struct {
	RecordUUID uuid.UUID `json:"record_uuid"`
	DeviceHash string    `json:"device_hash"`
}

// ClockOutRequest payload for real-time online clock-out
type ClockOutRequest struct {
	RecordUUID uuid.UUID `json:"record_uuid"`
	DeviceHash string    `json:"device_hash"`
}

// SyncRecordRequest represents an individual attendance log sent from offline queue
type SyncRecordRequest struct {
	RecordUUID string `json:"record_uuid"`
	OrgID      string `json:"org_id"`
	UserID     string `json:"user_id"`
	ActionType string `json:"action_type"` // "CLOCK_IN" or "CLOCK_OUT"
	Timestamp  string `json:"timestamp"`
	DeviceHash string `json:"device_hash"`
}

// BatchSyncRequest accepts batch synced array or single record fallback
type BatchSyncRequest struct {
	Records []SyncRecordRequest `json:"records"`
}

// SyncResult defines individual record sync result returned to the client
type SyncResult struct {
	RecordUUID string `json:"record_uuid"`
	Status     string `json:"status"` // "SYNCED_VERIFIED", "REJECTED_TAMPERED", "ALREADY_SYNCED"
	Message    string `json:"message,omitempty"`
}

// BatchSyncResponse standardizes batch processing output
type BatchSyncResponse struct {
	Processed int          `json:"processed"`
	Results   []SyncResult `json:"results"`
}

// AttendanceResponse standardizes online endpoint JSON responses
type AttendanceResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	ClockIn    time.Time  `json:"clock_in"`
	ClockOut   *time.Time `json:"clock_out"`
	TotalHours *float64   `json:"total_hours"`
	SyncStatus SyncStatus `json:"sync_status"`
	DeviceHash string     `json:"device_hash"`
	RecordUUID uuid.UUID  `json:"record_uuid"`
}