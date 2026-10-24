package notifications

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

// CalculateNext930Am computes the next 9:30 AM execution time in Addis Ababa time and the sleep duration.
func CalculateNext930Am(now time.Time) (nextRun time.Time, sleepDuration time.Duration) {
	localNow := now.In(AddisLocation)
	nextRun = time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 9, 30, 0, 0, AddisLocation)
	if !localNow.Before(nextRun) {
		nextRun = nextRun.Add(24 * time.Hour)
	}
	sleepDuration = nextRun.Sub(localNow)
	if sleepDuration <= 0 {
		sleepDuration = 1 * time.Minute
	}
	return nextRun, sleepDuration
}

// CalculateDayBoundsUTC returns the [start, end) time in UTC corresponding to the full local day in Addis Ababa.
func CalculateDayBoundsUTC(t time.Time) (startUTC, endUTC time.Time) {
	local := t.In(AddisLocation)
	dayStartLocal := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, AddisLocation)
	startUTC = dayStartLocal.UTC()
	endUTC = startUTC.Add(24 * time.Hour)
	return startUTC, endUTC
}

// listActiveOrgIDs queries all active, non-suspended organization IDs.
func listActiveOrgIDs(db *gorm.DB) ([]uuid.UUID, error) {
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

// StartClockInReminderScheduler sends a daily clock-in reminder to all active
// users across active organizations who haven't clocked in yet by 9:30 AM Addis Ababa time.
// Excludes users who already clocked in or who have already received a reminder for the day.
// FR-NOTIFY-04
func StartClockInReminderScheduler(ctx context.Context, db *gorm.DB) {
	sender := NewTelegramSender(db)

	go func() {
		log.Println("⚙️  clock-in reminder scheduler started")
		for {
			next930, sleepDuration := CalculateNext930Am(time.Now().In(AddisLocation))
			log.Printf("⏰ clock-in reminder scheduler: next execution in %s (at %s)",
				sleepDuration.Round(time.Minute),
				next930.Format("2006-01-02 15:04:05 MST"),
			)

			timer := time.NewTimer(sleepDuration)
			select {
			case <-ctx.Done():
				timer.Stop()
				log.Println("🛑 clock-in reminder scheduler stopped gracefully")
				return
			case <-timer.C:
			}

			// Execution time: compute today's Addis Ababa day bounds in UTC
			dayStartUTC, dayEndUTC := CalculateDayBoundsUTC(time.Now().In(AddisLocation))

			orgIDs, err := listActiveOrgIDs(db)
			if err != nil {
				log.Printf("❌ clock-in reminder scheduler: failed to query active organizations: %v", err)
				continue
			}

			for _, orgID := range orgIDs {
				func(oID uuid.UUID) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("❌ clock-in reminder: panic recovered for org %s: %v", oID, r)
						}
					}()

					// Find active users who have not clocked in today AND have not already been sent a reminder today
					var userIDs []uuid.UUID
					queryErr := db.Raw(`
						SELECT u.id FROM users u
						WHERE u.org_id = ? AND u.status = 'active' AND u.deleted_at IS NULL
						  AND NOT EXISTS (
						      SELECT 1 FROM attendance_logs a
						      WHERE a.user_id = u.id AND a.org_id = u.org_id
						        AND a.clock_in >= ? AND a.clock_in < ?
						        AND a.sync_status = 'SYNCED_VERIFIED'
						  )
						  AND NOT EXISTS (
						      SELECT 1 FROM notifications n
						      WHERE n.user_id = u.id AND n.org_id = u.org_id
						        AND n.type = 'clock_in_reminder'
						        AND n.created_at >= ? AND n.created_at < ?
						  )
					`, oID, dayStartUTC, dayEndUTC, dayStartUTC, dayEndUTC).Scan(&userIDs).Error

					if queryErr != nil {
						log.Printf("❌ clock-in reminder: failed to query eligible users for org %s: %v", oID, queryErr)
						return
					}

					for _, userID := range userIDs {
						sender.Send(oID, userID, "clock_in_reminder",
							"👋 Good morning! Don't forget to clock in for today.")
					}

					if len(userIDs) > 0 {
						log.Printf("✅ clock-in reminder: sent %d reminder(s) for org %s", len(userIDs), oID)
					}
				}(orgID)
			}
		}
	}()
}
