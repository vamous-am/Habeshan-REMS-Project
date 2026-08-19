package migrations

import "gorm.io/gorm"

// RunAll executes every migration in order. Each numbered migration file
// (0001_dev1_organizations.go, 0002_dev1_users.go, ...) will register its
// AutoMigrate call here as it's implemented in feature/dev1-schema.
//
// Deliberately empty for now — Branch 0 only wires the runner itself so
// `go run ./cmd/migrate` works end-to-end before any table exists.
func RunAll(db *gorm.DB) error {
	// TODO(dev1-schema): db.AutoMigrate(&auth.Organization{}, &auth.User{}, &auth.Team{}, &auth.TeamMember{})
	return nil
}