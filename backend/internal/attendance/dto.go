package attendance

import (
	"time"

	"github.com/google/uuid"
)

type ClockInRequest struct {
	RecordUUID uuid.UUID `json:"record_uuid"`
	DeviceHash string    `json:"device_hash"`
}

type ClockOutRequest struct {
	RecordUUID uuid.UUID `json:"record_uuid"`
	DeviceHash string    `json:"device_hash"`
}

type SyncRecordRequest struct {
	RecordUUID string `json:"record_uuid"`
	OrgID      string `json:"org_id"`
	UserID     string `json:"user_id"`
	ActionType string `json:"action_type"` // "CLOCK_IN" or "CLOCK_OUT"
	Timestamp  string `json:"timestamp"`
	DeviceHash string `json:"device_hash"`
}

type BatchSyncRequest struct {
	Records []SyncRecordRequest `json:"records"`
}

type SyncResult struct {
	RecordUUID string `json:"record_uuid"`
	Status     string `json:"status"` // "SYNCED_VERIFIED", "REJECTED_TAMPERED", or "ALREADY_SYNCED"
	Message    string `json:"message,omitempty"`
}

type BatchSyncResponse struct {
	Processed int          `json:"processed"`
	Results   []SyncResult `json:"results"`
}

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