package auth

import (
	"errors"
	"os"
	"time"	

    "strings" 
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)
// Service holds the DB connection and contains all auth business logic.
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}
// ----Register -------
// Register creates a new Organization and its first User in one transaction.
// Role is always forced to employee (FR-AUTH-08).
// Returns a JWT so the user is logged in immediately after registering.
func (s *Service) Register(req RegisterRequest) (LoginResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10) // cost >= 10, FR-AUTH-02
	if err != nil {
		return LoginResponse{}, common.ErrInternal
	}

	var response LoginResponse

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create the organization
		org := Organization{Name: req.OrgName}
		if err := tx.Create(&org).Error; err != nil {
			return common.ErrInternal
		}

		// 2. Create the user — role always forced to employee (FR-AUTH-08)
		user := User{
			Email:        req.Email,
			PasswordHash: string(hash),
			FullName:     req.FullName,
			Phone:        req.Phone,
			Role:         RoleEmployee,
			Status:       UserStatusActive,
		}
		user.OrgID = org.ID.ID

		// Check for duplicate email within this org before inserting
		var count int64
		if err := tx.Model(&User{}).
			Where("org_id = ? AND email = ? AND deleted_at IS NULL", org.ID.ID, req.Email).
			Count(&count).Error; err != nil {
			return common.ErrInternal
		}
		if count > 0 {
			return common.ErrConflict
		}

		if err := tx.Create(&user).Error; err != nil {
            if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "23505") {
            return common.ErrConflict
          }
        return common.ErrInternal
		}

		// 3. Issue JWT immediately so the user is logged in after registering
		token, err := issueJWT(user)
		if err != nil {
			return common.ErrInternal
		}

		response = LoginResponse{
			Token: token,
			User:  toUserDTO(user),
		}
		return nil
	})

	return response, err
}
// ----Password Reset -------
//
// ⚠️ ponytail: MVP/demo flow — token returned in API response, not emailed.
// Ceiling: anyone who can read the API response can reset any password.
// Upgrade path: send token via Telegram (Dev 4) or email before real deployment.
// Token is a UUID stored in a process-local map with a 15-min expiry.
// Not persistent across restarts — acceptable for demo purposes.

// resetTokens holds active reset tokens: token → {orgID, userID, expiry}.
// ponytail: process-local map, lost on restart, no concurrency issue at MVP scale.
var resetTokens = map[string]resetEntry{}

type resetEntry struct {
	OrgID   string
	UserID  string
	Expires time.Time
}

// ForgotPassword generates a time-limited reset token and returns it directly
// in the response. Scoped by org_id + email to match our per-org uniqueness rule.
func (s *Service) ForgotPassword(req ForgotPasswordRequest) (ForgotPasswordResponse, error) {
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return ForgotPasswordResponse{}, common.ErrBadRequest
	}

	var user User
	if err := s.db.
		Where("email = ? AND org_id = ? AND deleted_at IS NULL", req.Email, orgID).
		First(&user).Error; err != nil {
		// Return success regardless — don't leak whether the email exists
		return ForgotPasswordResponse{ResetToken: uuid.NewString()}, nil
	}

	token := uuid.NewString()
	resetTokens[token] = resetEntry{
		OrgID:   user.OrgID.String(),
		UserID:  user.ID.ID.String(),
		Expires: time.Now().Add(15 * time.Minute),
	}

	return ForgotPasswordResponse{ResetToken: token}, nil
}

// ResetPassword consumes a valid reset token and sets a new bcrypt password.
// Token is single-use — deleted immediately on consumption.
func (s *Service) ResetPassword(req ResetPasswordRequest) error {
	entry, ok := resetTokens[req.ResetToken]
	if !ok || time.Now().After(entry.Expires) {
		delete(resetTokens, req.ResetToken) // clean up expired token if present
		return common.ErrUnauthorized
	}
	delete(resetTokens, req.ResetToken) // single-use

	userID, err := uuid.Parse(entry.UserID)
	if err != nil {
		return common.ErrInternal
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		return common.ErrInternal
	}

	if err := s.db.Model(&User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("password_hash", string(hash)).Error; err != nil {
		return common.ErrInternal
	}

	return nil
}

// ----Lookup -------

// Lookup returns the list of organizations an email belongs to (FR-AUTH-01).
// Step 1 of the two-step login flow. Returns an empty list (not 404) when
// the email is not found — the login step returns the same 401 either way,
// so leaking "email exists" here adds no extra information to an attacker.
func (s *Service) Lookup(req LookupRequest) (LookupResponse, error) {
	var users []User
	if err := s.db.
		Where("email = ? AND deleted_at IS NULL", req.Email).
		Find(&users).Error; err != nil {
		return LookupResponse{}, common.ErrInternal
	}

	if len(users) == 0 {
		return LookupResponse{Orgs: []OrgSummary{}}, nil
	}

	// Collect org IDs, then fetch org names in one query
	orgIDs := make([]string, len(users))
	for i, u := range users {
		orgIDs[i] = u.OrgID.String()
	}

	var orgs []Organization
	if err := s.db.
		Where("id IN ? AND deleted_at IS NULL", orgIDs).
		Find(&orgs).Error; err != nil {
		return LookupResponse{}, common.ErrInternal
	}

	// Build a name map for O(1) lookup
	nameByID := make(map[string]string, len(orgs))
	for _, o := range orgs {
		nameByID[o.ID.ID.String()] = o.Name
	}

	summaries := make([]OrgSummary, 0, len(users))
	for _, u := range users {
		summaries = append(summaries, OrgSummary{
			OrgID:   u.OrgID.String(),
			OrgName: nameByID[u.OrgID.String()],
		})
	}

	return LookupResponse{Orgs: summaries}, nil
}

// ----Login -------

// Login verifies email + password and returns a JWT (FR-AUTH-01/02/03).
// org_id is required — scopes the lookup to a single organization so the
// same email can exist under multiple orgs without ambiguity (BR-15).
// Always returns ErrUnauthorized for wrong email, wrong org, and wrong
// password to avoid user enumeration.
func (s *Service) Login(req LoginRequest) (LoginResponse, error) {
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return LoginResponse{}, common.ErrBadRequest
	}

	var user User
	if err := s.db.
		Where("email = ? AND org_id = ? AND deleted_at IS NULL", req.Email, orgID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoginResponse{}, common.ErrUnauthorized
		}
		return LoginResponse{}, common.ErrInternal
	}

	// bcrypt.CompareHashAndPassword handles cost >= 10 automatically since
	// the cost is encoded in the hash itself — FR-AUTH-02
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResponse{}, common.ErrUnauthorized
	}

	token, err := issueJWT(user)
	if err != nil {
		return LoginResponse{}, common.ErrInternal
	}

	return LoginResponse{
		Token: token,
		User:  toUserDTO(user),
	}, nil
}
// ----JWT -------

// Claims is the JWT payload. Published in contracts/api/auth.md.
type Claims struct {
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// issueJWT creates a signed JWT containing user_id, org_id and role (FR-AUTH-03).
func issueJWT(user User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	expiry := 24 * time.Hour // matches JWT_EXPIRES_IN default in .env.example

	claims := Claims{
		UserID: user.ID.ID.String(),
		OrgID:  user.OrgID.String(),
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			ID:        uuid.NewString(), // jti — unique per token, useful for future revocation
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
// --- Helpers -----

// toUserDTO converts a User model to the safe public DTO (no password).
func toUserDTO(u User) UserDTO {
	return UserDTO{
		ID:       u.ID.ID.String(),
		OrgID:    u.OrgID.String(),
		Email:    u.Email,
		FullName: u.FullName,
		Phone:    u.Phone,
		Role:     string(u.Role),
		Status:   string(u.Status),
	}
}