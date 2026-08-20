package migrations

import (
	"gorm.io/gorm"

	"github.com/habeshan-rems/backend/internal/auth"
)
// Migrate0003Teams creates or updates the teams and team_members tables.
// Run after Migrate0002Users, since teams.manager_id and
// team_members.user_id both reference users.id.
func Migrate0003Teams(db *gorm.DB) error {
	if err := db.AutoMigrate(&auth.Team{}); err != nil {
		return err
	}
	return db.AutoMigrate(&auth.TeamMember{})
}