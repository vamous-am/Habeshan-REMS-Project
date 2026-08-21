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

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is returned after a successful login.
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
