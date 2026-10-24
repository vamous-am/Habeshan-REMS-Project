package migrations

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Migrate0012TimesheetsUniquePeriod adds a unique index on
// timesheets (org_id, user_id, period_start, period_end)
// after safely resolving any existing duplicate groups.
func Migrate0012TimesheetsUniquePeriod(db *gorm.DB) error {
	// 1. Detect duplicate groups
	type duplicateGroup struct {
		OrgID       string `gorm:"column:org_id"`
		UserID      string `gorm:"column:user_id"`
		PeriodStart string `gorm:"column:period_start"`
		PeriodEnd   string `gorm:"column:period_end"`
		Count       int64  `gorm:"column:count"`
	}

	var duplicates []duplicateGroup
	err := db.Raw(`
		SELECT org_id, user_id, period_start, period_end, COUNT(*) as count
		FROM timesheets
		GROUP BY org_id, user_id, period_start, period_end
		HAVING COUNT(*) > 1
	`).Scan(&duplicates).Error
	if err != nil {
		return fmt.Errorf("failed to check for duplicate timesheets: %w", err)
	}

	// 2. Resolve duplicates by keeping the row with the most advanced status or latest update
	if len(duplicates) > 0 {
		log.Printf("⚠️ migration 0012: found %d duplicate timesheet group(s); resolving...", len(duplicates))
		for _, dup := range duplicates {
			// Delete duplicate rows keeping the one with the latest updated_at
			deleteErr := db.Exec(`
				DELETE FROM timesheets
				WHERE id IN (
					SELECT id FROM timesheets
					WHERE org_id = ? AND user_id = ? AND period_start = ? AND period_end = ?
					ORDER BY 
						CASE status 
							WHEN 'approved' THEN 4 
							WHEN 'submitted' THEN 3 
							WHEN 'rejected' THEN 2 
							ELSE 1 
						END DESC,
						updated_at DESC
					OFFSET 1
				)
			`, dup.OrgID, dup.UserID, dup.PeriodStart, dup.PeriodEnd).Error
			if deleteErr != nil {
				return fmt.Errorf("failed to deduplicate timesheets for user %s: %w", dup.UserID, deleteErr)
			}
		}
		log.Printf("✅ migration 0012: all duplicate timesheets resolved")
	}

	// 3. Create unique index
	indexSQL := `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_timesheets_org_user_period
		ON timesheets (org_id, user_id, period_start, period_end)
	`
	if err := db.Exec(indexSQL).Error; err != nil {
		return fmt.Errorf("failed to create idx_timesheets_org_user_period: %w", err)
	}

	return nil
}
