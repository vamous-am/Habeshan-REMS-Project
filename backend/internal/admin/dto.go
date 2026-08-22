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