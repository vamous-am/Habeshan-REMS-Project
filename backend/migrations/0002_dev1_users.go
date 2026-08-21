package migrations

import (
	"gorm.io/gorm"

	"github.com/habeshan-rems/backend/internal/auth"
)

// migrateUsers creates/updates the users table. Must run after
// migrateOrganizations — users.org_id has an FK to organizations.id.
func Migrate0002Users(db *gorm.DB) error {
	return db.AutoMigrate(&auth.User{})
}
