package admin

// ── User management ────────

type UpdateUserRequest struct {
	FullName *string `json:"full_name,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Status   *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin manager employee"`
}

// CreateUserRequest is the request body for POST /admin/users.
// org_id is intentionally absent — it is always taken from the JWT claim
// so that a caller cannot create users in a foreign org.
// role is restricted to "manager" or "employee" — admin promotion remains
// CLI-only through the seed-admin command (see admin/service.go:PromoteAdmin).
type CreateUserRequest struct {
	Email    string `json:"email"     validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Password string `json:"password"  validate:"required,min=8"`
	Role     string `json:"role"      validate:"required,oneof=manager employee"`
	Phone    string `json:"phone,omitempty"`
}

// ── Teams ────

type CreateTeamRequest struct {
	Name      string `json:"name"       validate:"required,min=2,max=100"`
	ManagerID string `json:"manager_id" validate:"required,uuid"`
}

type AddTeamMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type TeamDTO struct {
	ID        string `json:"id"`
	OrgID     string `json:"org_id"`
	Name      string `json:"name"`
	ManagerID string `json:"manager_id"`
}

// ── Org settings ──────────────────────────────────────────────────────────────

type UpdateOrgRequest struct {
	Name     *string `json:"name,omitempty"`
	Currency *string `json:"currency,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
}

type OrgDTO struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Currency           string `json:"currency"`
	Timezone           string `json:"timezone"`
	SeatCount          int    `json:"seat_count"`
	SubscriptionStatus string `json:"subscription_status"`
}