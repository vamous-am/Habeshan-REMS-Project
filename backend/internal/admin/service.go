package admin

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ── User CRUD (FR-ADMIN-01) ────────────────────────────────────────────────────

func (s *Service) ListUsers(orgID uuid.UUID) ([]auth.UserDTO, error) {
	var users []auth.User
	if err := s.db.Where("org_id = ? AND deleted_at IS NULL", orgID).Find(&users).Error; err != nil {
		return nil, common.ErrInternal
	}

	dtos := make([]auth.UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, toUserDTO(u))
	}
	return dtos, nil
}

func (s *Service) GetUser(orgID, userID uuid.UUID) (auth.UserDTO, error) {
	user, err := s.findUser(orgID, userID)
	if err != nil {
		return auth.UserDTO{}, err
	}
	return toUserDTO(user), nil
}

func (s *Service) UpdateUser(orgID, userID uuid.UUID, req UpdateUserRequest) (auth.UserDTO, error) {
	user, err := s.findUser(orgID, userID)
	if err != nil {
		return auth.UserDTO{}, err
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Status != nil {
		user.Status = auth.UserStatus(*req.Status)
	}

	if err := s.db.Save(&user).Error; err != nil {
		return auth.UserDTO{}, common.ErrInternal
	}
	return toUserDTO(user), nil
}

// DeleteUser soft-deletes via GORM's built-in behavior — User embeds
// common.SoftDeletable (gorm.DeletedAt), so .Delete() sets deleted_at
// instead of removing the row. Satisfies FR-AUTH-09.
// Rejects with 409 if deleting the last Admin would lock the org out.
func (s *Service) DeleteUser(orgID, userID uuid.UUID) error {
	user, err := s.findUser(orgID, userID)
	if err != nil {
		return err
	}

	if user.Role == auth.RoleAdmin {
		var adminCount int64
		if err := s.db.Model(&auth.User{}).
			Where("org_id = ? AND role = ? AND deleted_at IS NULL", orgID, auth.RoleAdmin).
			Count(&adminCount).Error; err != nil {
			return common.ErrInternal
		}
		if adminCount <= 1 {
			return common.ErrConflict
		}
	}

	if err := s.db.Delete(&user).Error; err != nil {
		return common.ErrInternal
	}
	return nil
}

// ── Role assignment (FR-ADMIN-02) ──────────────────────────────────────────────

// UpdateUserRole changes a user's role. Rejects the change with 409 if it
// would leave the org with zero Admins (FR-ADMIN-02).
func (s *Service) UpdateUserRole(orgID, userID uuid.UUID, role string) (auth.UserDTO, error) {
	user, err := s.findUser(orgID, userID)
	if err != nil {
		return auth.UserDTO{}, err
	}

	// Guard: if we're demoting an Admin, ensure at least one other Admin remains.
	if user.Role == auth.RoleAdmin && auth.Role(role) != auth.RoleAdmin {
		var adminCount int64
		if err := s.db.Model(&auth.User{}).
			Where("org_id = ? AND role = ? AND deleted_at IS NULL", orgID, auth.RoleAdmin).
			Count(&adminCount).Error; err != nil {
			return auth.UserDTO{}, common.ErrInternal
		}
		if adminCount <= 1 {
			return auth.UserDTO{}, common.ErrConflict // "resource already exists" sentinel → 409
		}
	}

	user.Role = auth.Role(role)
	if err := s.db.Save(&user).Error; err != nil {
		return auth.UserDTO{}, common.ErrInternal
	}
	return toUserDTO(user), nil
}

// ── Teams (FR-ADMIN-03) ─────────────────────────────────────────────────────────

func (s *Service) CreateTeam(orgID uuid.UUID, req CreateTeamRequest) (TeamDTO, error) {
	managerID, err := uuid.Parse(req.ManagerID)
	if err != nil {
		return TeamDTO{}, common.ErrBadRequest
	}

	// Confirm the manager exists in this org before creating the team
	var count int64
	s.db.Model(&auth.User{}).
		Where("id = ? AND org_id = ? AND deleted_at IS NULL", managerID, orgID).
		Count(&count)
	if count == 0 {
		return TeamDTO{}, common.ErrBadRequest
	}

	team := auth.Team{Name: req.Name, ManagerID: managerID}
	team.OrgID = orgID

	if err := s.db.Create(&team).Error; err != nil {
		return TeamDTO{}, common.ErrInternal
	}
	return toTeamDTO(team), nil
}

func (s *Service) ListTeams(orgID uuid.UUID) ([]TeamDTO, error) {
	var teams []auth.Team
	if err := s.db.Where("org_id = ?", orgID).Find(&teams).Error; err != nil {
		return nil, common.ErrInternal
	}

	dtos := make([]TeamDTO, 0, len(teams))
	for _, t := range teams {
		dtos = append(dtos, toTeamDTO(t))
	}
	return dtos, nil
}

func (s *Service) AddTeamMember(orgID, teamID uuid.UUID, req AddTeamMemberRequest) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return common.ErrBadRequest
	}

	if err := s.confirmTeamInOrg(orgID, teamID); err != nil {
		return err
	}

	var userCount int64
	s.db.Model(&auth.User{}).
		Where("id = ? AND org_id = ? AND deleted_at IS NULL", userID, orgID).
		Count(&userCount)
	if userCount == 0 {
		return common.ErrBadRequest
	}

	member := auth.TeamMember{TeamID: teamID, UserID: userID}
	if err := s.db.Create(&member).Error; err != nil {
		return common.ErrConflict // likely already a member — composite PK violation
	}
	return nil
}

func (s *Service) RemoveTeamMember(orgID, teamID, userID uuid.UUID) error {
	if err := s.confirmTeamInOrg(orgID, teamID); err != nil {
		return err
	}
	if err := s.db.Delete(&auth.TeamMember{}, "team_id = ? AND user_id = ?", teamID, userID).Error; err != nil {
		return common.ErrInternal
	}
	return nil
}

// ── Org settings (FR-ADMIN-04) ──────────────────────────────────────────────────

func (s *Service) GetOrgSettings(orgID uuid.UUID) (OrgDTO, error) {
	var org auth.Organization
	if err := s.db.Where("id = ? AND deleted_at IS NULL", orgID).First(&org).Error; err != nil {
		return OrgDTO{}, common.ErrNotFound
	}
	return toOrgDTO(org), nil
}

func (s *Service) UpdateOrgSettings(orgID uuid.UUID, req UpdateOrgRequest) (OrgDTO, error) {
	var org auth.Organization
	if err := s.db.Where("id = ? AND deleted_at IS NULL", orgID).First(&org).Error; err != nil {
		return OrgDTO{}, common.ErrNotFound
	}

	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Currency != nil {
		org.Currency = *req.Currency
	}
	if req.Timezone != nil {
		org.Timezone = *req.Timezone
	}

	if err := s.db.Save(&org).Error; err != nil {
		return OrgDTO{}, common.ErrInternal
	}
	return toOrgDTO(org), nil
}

// ── Bootstrap (seed the first Admin) ────────────────────────────────────────────

// ErrAdminAlreadyExists is returned when the target organization already
// has an active Admin — PromoteAdmin refuses to create a second one,
// regardless of which user is being promoted.
var ErrAdminAlreadyExists = errors.New("an admin already exists for this organization")

type PromoteAdminInput struct {
	OrgID uuid.UUID
	Email string
}

// PromoteAdmin promotes an already-registered user to Admin. Deliberately
// does NOT create a user or set a password — the target must have already
// registered via POST /auth/register, proving they control that
// email/password pair through the normal bcrypt-verified login flow.
// This only flips one field: role.
//
// Why promote, not create: creating a new user here would mean handling a
// password on the CLI (shell history, `ps aux`, CI logs — all real leak
// vectors), and it'd be a second, parallel path for setting a password
// outside Register()'s bcrypt-cost-checked flow. Promoting an existing user
// needs zero secrets on the command line.
//
// Refuses if the org already has ANY admin (checked by role, not by this
// specific user) — running this twice against two different emails in the
// same org is a safe no-op on the second run, never a second admin.
func (s *Service) PromoteAdmin(in PromoteAdminInput) (auth.User, error) {
	var org auth.Organization
	if err := s.db.First(&org, "id = ? AND deleted_at IS NULL", in.OrgID).Error; err != nil {
		return auth.User{}, fmt.Errorf("organization %s not found: %w", in.OrgID, err)
	}

	var adminCount int64
	if err := s.db.Model(&auth.User{}).
		Where("org_id = ? AND role = ? AND deleted_at IS NULL", in.OrgID, auth.RoleAdmin).
		Count(&adminCount).Error; err != nil {
		return auth.User{}, err
	}
	if adminCount > 0 {
		return auth.User{}, ErrAdminAlreadyExists
	}

	var user auth.User
	if err := s.db.Where("org_id = ? AND email = ? AND deleted_at IS NULL", in.OrgID, in.Email).
		First(&user).Error; err != nil {
		return auth.User{}, fmt.Errorf("no registered user with email %s in org %s — they must register via POST /auth/register first: %w", in.Email, in.OrgID, err)
	}

	user.Role = auth.RoleAdmin
	if err := s.db.Save(&user).Error; err != nil {
		return auth.User{}, err
	}
	return user, nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (s *Service) findUser(orgID, userID uuid.UUID) (auth.User, error) {
	var user auth.User
	err := s.db.Where("id = ? AND org_id = ? AND deleted_at IS NULL", userID, orgID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.User{}, common.ErrNotFound
		}
		return auth.User{}, common.ErrInternal
	}
	return user, nil
}

func (s *Service) confirmTeamInOrg(orgID, teamID uuid.UUID) error {
	var count int64
	s.db.Model(&auth.Team{}).Where("id = ? AND org_id = ?", teamID, orgID).Count(&count)
	if count == 0 {
		return common.ErrNotFound
	}
	return nil
}

func toUserDTO(u auth.User) auth.UserDTO {
	return auth.UserDTO{
		ID:       u.ID.ID.String(),
		OrgID:    u.OrgID.String(),
		Email:    u.Email,
		FullName: u.FullName,
		Phone:    u.Phone,
		Role:     string(u.Role),
		Status:   string(u.Status),
	}
}

func toTeamDTO(t auth.Team) TeamDTO {
	return TeamDTO{
		ID:        t.ID.ID.String(),
		OrgID:     t.OrgID.String(),
		Name:      t.Name,
		ManagerID: t.ManagerID.String(),
	}
}

func toOrgDTO(o auth.Organization) OrgDTO {
	return OrgDTO{
		ID:                 o.ID.ID.String(),
		Name:               o.Name,
		Currency:           o.Currency,
		Timezone:           o.Timezone,
		SeatCount:          o.SeatCount,
		SubscriptionStatus: string(o.SubscriptionStatus),
	}
}