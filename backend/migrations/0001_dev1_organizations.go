package migrations

import (
	"gorm.io/gorm"

	"github.com/habeshan-rems/backend/internal/auth"
)

// migrateOrganizations creates/updates the organizations table.
func Migrate0001Organizations(db *gorm.DB) error {
	return db.AutoMigrate(&auth.Organization{})
}
