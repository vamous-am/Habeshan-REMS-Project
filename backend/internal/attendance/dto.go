package attendance

import (
	"time"

	"github.com/google/uuid"
)

// ClockInRequest holds incoming JSON payload for POST /attendance/clock-in
type ClockInRequest struct {
	DeviceHash string    `json:"device_hash"`
	RecordUUID uuid.UUID `json:"record_uuid"`
	Timestamp  time.Time `json:"timestamp"`
}

// ClockOutRequest holds incoming JSON payload for POST /attendance/clock-out
type ClockOutRequest struct {
	DeviceHash string    `json:"device_hash,omitempty"`
	RecordUUID uuid.UUID `json:"record_uuid,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// AttendanceResponse represents the standardized response payload returned to clients
type AttendanceResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	ClockIn    time.Time  `json:"clock_in"`
	ClockOut   *time.Time `json:"clock_out,omitempty"`
	TotalHours *float64   `json:"total_hours,omitempty"`
	SyncStatus SyncStatus `json:"sync_status"`
	DeviceHash string     `json:"device_hash"`
	RecordUUID uuid.UUID  `json:"record_uuid"`
}