package auth

// DTOs (Data Transfer Objects) — the exact shapes the API accepts and returns.

// RegisterRequest is the body for POST /auth/register.
// Role is never accepted from the client (FR-AUTH-08).
type RegisterRequest struct {
	OrgName  string `json:"org_name" validate:"required,min=2,max=150"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8"`
	Phone    string `json:"phone" validate:"omitempty,max=20"`
}

// LookupRequest is the body for POST /auth/lookup.
// Step 1 of the two-step login flow — returns which orgs have this email.
type LookupRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// OrgSummary is one entry in a LookupResponse — only what the login screen needs.
type OrgSummary struct {
	OrgID   string `json:"org_id"`
	OrgName string `json:"org_name"`
}

// LookupResponse is the body returned by POST /auth/lookup.
type LookupResponse struct {
	Orgs []OrgSummary `json:"orgs"`
}

// LoginRequest is the body for POST /auth/login.
// org_id is required — use POST /auth/lookup first to discover it.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
	OrgID    string `json:"org_id"   validate:"required,uuid"`
}

// ForgotPasswordRequest is the body for POST /auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
	OrgID string `json:"org_id" validate:"required,uuid"`
}

// ForgotPasswordResponse returns the reset token directly in the response.
// ⚠️ MVP/demo only — NOT safe for production. Token must be delivered
// out-of-band (email/SMS) before real deployment. See contracts/api/auth.md.
type ForgotPasswordResponse struct {
	ResetToken string `json:"reset_token"`
}

// ResetPasswordRequest is the body for POST /auth/reset-password.
type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token"  validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// UserDTO is the public representation of a user (no password).
type UserDTO struct {
	ID       string `json:"id"`
	OrgID    string `json:"org_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone,omitempty"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}
