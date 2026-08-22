package notifications

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StartClockInReminderScheduler sends a daily clock-in reminder to all active
// users in the org who haven't clocked in yet by 9:30 AM Addis Ababa time.
// FR-NOTIFY-04
func StartClockInReminderScheduler(db *gorm.DB, orgID uuid.UUID) {
	sender := NewTelegramSender(db)

	go func() {
		for {
			now := time.Now().UTC()

			// Addis Ababa is UTC+3
			addisNow := now.Add(3 * time.Hour)

			// Next 9:30 AM Addis Ababa time
			next930 := time.Date(addisNow.Year(), addisNow.Month(), addisNow.Day(), 9, 30, 0, 0, addisNow.Location())
			if addisNow.After(next930) {
				next930 = next930.Add(24 * time.Hour)
			}

			sleepDuration := next930.Sub(addisNow)
			log.Printf("⏰ clock-in reminder: next run in %s", sleepDuration.Round(time.Minute))
			time.Sleep(sleepDuration)

			// Find users who haven't clocked in today
			var userIDs []uuid.UUID
			db.Raw(`
				SELECT u.id FROM users u
				WHERE u.org_id = ? AND u.status = 'active'
				AND u.id NOT IN (
					SELECT DISTINCT user_id FROM attendance_logs
					WHERE DATE(clock_in) = CURRENT_DATE
					AND sync_status = 'SYNCED_VERIFIED'
				)
			`, orgID).Scan(&userIDs)

			for _, userID := range userIDs {
				sender.Send(orgID, userID, "clock_in_reminder",
					"👋 Good morning! Don't forget to clock in for today.")
			}

			log.Printf("✅ clock-in reminders sent to %d users", len(userIDs))
		}
	}()
}
