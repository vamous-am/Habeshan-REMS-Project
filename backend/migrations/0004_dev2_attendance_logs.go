package migrations

// TODO Dev 2: AutoMigrate attendance_logs table

import (
	"github.com/habeshan-rems/backend/internal/attendance"
	"gorm.io/gorm"
)

// MigrateAttendanceLogs creates the attendance_logs table
func MigrateAttendanceLogs(db *gorm.DB) error {
	return db.AutoMigrate(&attendance.AttendanceLog{})
}
