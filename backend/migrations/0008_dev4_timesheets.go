package migrations

import (
	"gorm.io/gorm"

	"github.com/habeshan-rems/backend/internal/timesheets" // match your actual module path
)

func Migrate0008Timesheets(db *gorm.DB) error {
	return db.AutoMigrate(&timesheets.Timesheet{})
}
