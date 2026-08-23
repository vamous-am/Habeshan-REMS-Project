package timesheets

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StartScheduler kicks off a goroutine that generates draft timesheets
// automatically at the end of each weekly period.
// Call this once from main.go after the DB is initialised.
// FR-TS-01
func StartScheduler(db *gorm.DB, orgID uuid.UUID) {
	service := NewService(db)

	go func() {
		for {
			now := time.Now()

			// Calculate the start of the current ISO week (Monday)
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday → 7 so Monday is always day 1
			}
			monday := now.AddDate(0, 0, -(weekday - 1))
			periodStart := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
			periodEnd := periodStart.AddDate(0, 0, 7)

			// Sleep until the end of the current period
			sleepDuration := periodEnd.Sub(now)
			log.Printf("⏰ timesheet scheduler: next run in %s (period %s → %s)",
				sleepDuration.Round(time.Minute),
				periodStart.Format("2006-01-02"),
				periodEnd.Format("2006-01-02"),
			)
			time.Sleep(sleepDuration)

			// Generate drafts for the completed period
			log.Printf("⚙️  generating draft timesheets for period %s → %s",
				periodStart.Format("2006-01-02"),
				periodEnd.Format("2006-01-02"),
			)
			if err := service.GenerateDraftTimesheets(orgID, periodStart, periodEnd); err != nil {
				log.Printf("❌ timesheet scheduler error: %v", err)
			} else {
				log.Printf("✅ draft timesheets generated")
			}
		}
	}()
}
