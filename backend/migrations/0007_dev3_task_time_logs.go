package migrations

import (
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"gorm.io/gorm"
)

func MigrateTaskTimeLogs(db *gorm.DB) error {
	// org_id was added in migration 0011. Keep it out of this historical
	// migration so existing rows can be backfilled before the column becomes
	// NOT NULL.
	type taskTimeLogV1 struct {
		common.ID
		common.Timestamps

		TaskID          uuid.UUID  `gorm:"type:uuid;not null;index"`
		UserID          uuid.UUID  `gorm:"type:uuid;not null;index"`
		StartedAt       time.Time  `gorm:"not null"`
		StoppedAt       *time.Time `gorm:"default:null"`
		DurationMinutes *int       `gorm:"default:null"`
		PauseReason     *string    `gorm:"type:varchar(50);default:null"`
		SyncStatus      string     `gorm:"type:varchar(30);not null;default:'pending_sync'"`
		DeviceHash      string     `gorm:"type:varchar(128);not null"`
		RecordUUID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"`
	}

	return db.Table("task_time_logs").AutoMigrate(&taskTimeLogV1{})
}
