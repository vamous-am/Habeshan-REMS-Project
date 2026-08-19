package migrations

import (
	"github.com/habeshan-rems/backend/internal/tasks"
	"gorm.io/gorm"
)

func MigrateTaskAssignments(db *gorm.DB) error {
	return db.AutoMigrate(&tasks.TaskAssignment{})
}