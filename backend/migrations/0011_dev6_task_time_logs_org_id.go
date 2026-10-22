package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// Migrate0011TaskTimeLogsOrgID safely adds the org_id column to task_time_logs,
// backfills it from the parent task, verifies tenant integrity, and enforces NOT NULL.
func Migrate0011TaskTimeLogsOrgID(db *gorm.DB) error {
	// 1. Add column org_id as nullable if it does not already exist
	if !db.Migrator().HasColumn("task_time_logs", "org_id") {
		if err := db.Exec("ALTER TABLE task_time_logs ADD COLUMN org_id uuid;").Error; err != nil {
			return fmt.Errorf("failed to add org_id column to task_time_logs: %w", err)
		}
	}

	// 2. Backfill org_id from the associated task
	backfillSQL := `
		UPDATE task_time_logs ttl
		SET org_id = t.org_id
		FROM tasks t
		WHERE ttl.task_id = t.id
		  AND ttl.org_id IS NULL;
	`
	if err := db.Exec(backfillSQL).Error; err != nil {
		return fmt.Errorf("failed to backfill org_id in task_time_logs: %w", err)
	}

	// 3. Fail loudly if any rows remain without an organization
	var orphanCount int64
	if err := db.Table("task_time_logs").Where("org_id IS NULL").Count(&orphanCount).Error; err != nil {
		return fmt.Errorf("failed to check for orphaned task_time_logs: %w", err)
	}
	if orphanCount > 0 {
		return fmt.Errorf("migration failed: %d task_time_logs rows could not be associated with a valid task org_id", orphanCount)
	}

	// 4. Enforce NOT NULL constraint
	if err := db.Exec("ALTER TABLE task_time_logs ALTER COLUMN org_id SET NOT NULL;").Error; err != nil {
		return fmt.Errorf("failed to set org_id NOT NULL on task_time_logs: %w", err)
	}

	// 5. Create index if it does not exist
	if !db.Migrator().HasIndex("task_time_logs", "idx_task_time_logs_org_id") {
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_task_time_logs_org_id ON task_time_logs(org_id);").Error; err != nil {
			return fmt.Errorf("failed to create idx_task_time_logs_org_id: %w", err)
		}
	}

	return nil
}
