package auth

import (
	"github.com/google/uuid"

	"github.com/habeshan-rems/backend/internal/common"
)

type SubscriptionStatus string

const (
	SubscriptionTrial     SubscriptionStatus = "trial"
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionSuspended SubscriptionStatus = "suspended"
)

// Organization is a tenant workspace. It has no OrgID of its own — it does
// not embed common.TenantScoped — but it is soft-deletable
type Organization struct {
	common.ID
	common.Timestamps
	common.SoftDeletable

	Name               string             `gorm:"type:varchar(150);not null"                            json:"name"`
	Currency           string             `gorm:"type:varchar(3);not null;default:ETB"                  json:"currency"`
	Timezone           string             `gorm:"type:varchar(50);not null;default:Africa/Addis_Ababa"  json:"timezone"`
	SeatCount          int                `gorm:"not null;default:0"                                    json:"seat_count"`
	SubscriptionStatus SubscriptionStatus `gorm:"type:varchar(20);not null;default:trial;check:subscription_status IN ('trial','active','suspended')" json:"subscription_status"`
}

func (Organization) TableName() string { return "organizations" }

// Role is a user's permission level within their organization.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleManager  Role = "manager"
	RoleEmployee Role = "employee"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

// User is an account belonging to exactly one organization. Email is unique
// per organization, excluding soft-deleted rows, so a deleted account's
// email can be reused.
type User struct {
	common.ID
	common.Timestamps
	common.TenantScoped
	common.SoftDeletable

	Email        string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_org_email,where:deleted_at IS NULL" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	FullName     string     `gorm:"type:varchar(100);not null" json:"full_name"`
	Phone        string     `gorm:"type:varchar(20)"           json:"phone,omitempty"`
	Role         Role       `gorm:"type:varchar(20);not null;default:employee;check:role IN ('admin','manager','employee')" json:"role"`
	Status       UserStatus `gorm:"type:varchar(20);not null;default:active;check:status IN ('active','inactive')"          json:"status"`

	Organization Organization `gorm:"foreignKey:OrgID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (User) TableName() string { return "users" }

// Team groups users under one manager within an organization. Teams are
// never soft-deleted.
type Team struct {
	common.ID
	common.Timestamps
	common.TenantScoped

	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	ManagerID uuid.UUID `gorm:"type:uuid;not null;index"    json:"manager_id"`

	Organization Organization `gorm:"foreignKey:OrgID;references:ID;constraint:OnDelete:CASCADE"     json:"-"`
	Manager      User         `gorm:"foreignKey:ManagerID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
}

func (Team) TableName() string { return "teams" }


// TeamMember links a user to a team. Composite primary key, no own id or
// timestamps.
type TeamMember struct {
	TeamID uuid.UUID `gorm:"type:uuid;primaryKey" json:"team_id"`
	UserID uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`

	Team Team `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (TeamMember) TableName() string { return "team_members" }