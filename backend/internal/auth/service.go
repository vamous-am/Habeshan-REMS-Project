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
// ----Login -------

// Login verifies email + password and returns a JWT (FR-AUTH-01/02/03).
// Always returns ErrUnauthorized for both wrong email and wrong password
// to avoid user enumeration.
func (s *Service) Login(req LoginRequest) (LoginResponse, error) {
	var user User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", req.Email).
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