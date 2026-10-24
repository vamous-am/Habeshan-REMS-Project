package timesheets

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AddisLocation defines the UTC+3 timezone for Ethiopia (no DST).
var AddisLocation = time.FixedZone("Africa/Addis_Ababa", 3*60*60)

// CalculateWeeklyPeriod returns the start (Monday 00:00:00) and end (next Monday 00:00:00)
// for the ISO week containing 'now' in Addis Ababa time, along with sleep duration until periodEnd.
func CalculateWeeklyPeriod(now time.Time) (periodStart, periodEnd time.Time, sleepDuration time.Duration) {
	localNow := now.In(AddisLocation)

	weekday := int(localNow.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday -> 7 so Monday is always day 1
	}
	monday := localNow.AddDate(0, 0, -(weekday - 1))
	periodStart = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, AddisLocation)
	periodEnd = periodStart.AddDate(0, 0, 7)

	sleepDuration = periodEnd.Sub(localNow)
	if sleepDuration <= 0 {
		sleepDuration = 1 * time.Minute
	}
	return periodStart, periodEnd, sleepDuration
}

// ListActiveOrganizationIDs returns all active, non-suspended organization IDs.
func ListActiveOrganizationIDs(db *gorm.DB) ([]uuid.UUID, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	var orgIDs []uuid.UUID
	err := db.Table("organizations").
		Where("deleted_at IS NULL AND (subscription_status IS NULL OR subscription_status != 'suspended')").
		Pluck("id", &orgIDs).Error
	if err != nil {
		return nil, err
	}
	return orgIDs, nil
}

// StartScheduler kicks off a background worker loop that generates draft timesheets
// automatically at the end of each weekly period for all active organizations.
// Respects context cancellation for graceful shutdown.
func StartScheduler(ctx context.Context, db *gorm.DB) {
	service := NewService(db)

	go func() {
		log.Println("⚙️  timesheet scheduler started")
		for {
			periodStart, periodEnd, sleepDuration := CalculateWeeklyPeriod(time.Now().In(AddisLocation))

			log.Printf("⏰ timesheet scheduler: next execution in %s (for period %s → %s)",
				sleepDuration.Round(time.Minute),
				periodStart.Format("2006-01-02"),
				periodEnd.Format("2006-01-02"),
			)

			timer := time.NewTimer(sleepDuration)
			select {
			case <-ctx.Done():
				timer.Stop()
				log.Println("🛑 timesheet scheduler stopped gracefully")
				return
			case <-timer.C:
			}

			// Resolve active organizations at execution time
			orgIDs, err := ListActiveOrganizationIDs(db)
			if err != nil {
				log.Printf("❌ timesheet scheduler: failed to query active organizations: %v", err)
				continue
			}

			log.Printf("⚙️  timesheet scheduler: generating draft timesheets for %d organization(s)", len(orgIDs))
			for _, orgID := range orgIDs {
				func(oID uuid.UUID) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("❌ timesheet scheduler: panic recovered while processing org %s: %v", oID, r)
						}
					}()

					if genErr := service.GenerateDraftTimesheets(oID, periodStart, periodEnd); genErr != nil {
						log.Printf("❌ timesheet scheduler error for org %s: %v", oID, genErr)
					} else {
						log.Printf("✅ timesheet scheduler: drafts generated for org %s", oID)
					}
				}(orgID)
			}
		}
	}()
}
