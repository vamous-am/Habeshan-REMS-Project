package auth

// ─── Dummy / stub repositories ────────────────────────────────────────────────
//
// These stubs exist so the tasks service compiles and is testable before the
// real auth package (Dev 1) provides real User and Organization repositories.
//
// Rules:
//   1. Interface shapes MUST be stable — service.go depends on these interfaces.
//   2. Implementations do minimal DB look-ups only (primary-key lookups, team
//      membership queries needed by visibility rules FR-TASK-03).
//   3. When real auth repositories are available:
//      - delete this file
//      - update service constructor to accept the real repo interfaces
//      (the interface signatures themselves should remain identical)
//
// ⚠ DO NOT use these repositories outside the tasks package.

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── UserRepository ───────────────────────────────────────────────────────────

// UserRepository is the tasks-package interface for user look-ups.
// The real auth package must satisfy this interface.
type UserRepository interface {
	GetUserByID(userID uuid.UUID) (User, error)
	// GetTeamMembersForManager returns all user IDs that belong to any team
	// managed by managerID within orgID.  Used by FR-TASK-03 visibility query.
	GetTeamMembersForManager(managerID, orgID uuid.UUID) ([]uuid.UUID, error)
}

// OrganizationRepository is the tasks-package interface for org look-ups.
// The real auth package must satisfy this interface.
type OrganizationRepository interface {
	GetOrganizationByID(orgID uuid.UUID) (Organization, error)
}

// ─── userRepository (stub impl) ───────────────────────────────────────────────

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByID(userID uuid.UUID) (User, error) {
	var user User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// GetTeamMembersForManager returns the user IDs of every employee whose team
// is managed by managerID within orgID.
// SQL: SELECT tm.user_id FROM team_members tm
//
//	JOIN teams t ON tm.team_id = t.id
//	WHERE t.manager_id = ? AND t.org_id = ?
func (r *userRepository) GetTeamMembersForManager(managerID, orgID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.
		Table("team_members").
		Select("team_members.user_id").
		Joins("JOIN teams ON team_members.team_id = teams.id").
		Where("teams.manager_id = ? AND teams.org_id = ?", managerID, orgID).
		Scan(&userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

// ─── organizationRepository (stub impl) ───────────────────────────────────────

type organizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) GetOrganizationByID(orgID uuid.UUID) (Organization, error) {
	var org Organization
	err := r.db.Where("id = ?", orgID).First(&org).Error
	if err != nil {
		return Organization{}, err
	}
	return org, nil
}
