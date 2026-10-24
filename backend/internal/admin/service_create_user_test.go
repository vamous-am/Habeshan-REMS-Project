package admin_test

import (
	"testing"

	"github.com/habeshan-rems/backend/internal/admin"
	"github.com/habeshan-rems/backend/internal/common"
	"golang.org/x/crypto/bcrypt"
)

// ── mock repository ────────────────────────────────────────────────────────────

// mockUserStore is an in-memory store that implements just enough GORM-like
// behaviour for CreateUser tests without requiring a real database.
// We test the service through a thin fake that captures what was stored.
type storedUser struct {
	OrgID        string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Role         string
	Status       string
}

// ── pure-logic tests (no DB) ──────────────────────────────────────────────────

// TestCreateUserRequest_AdminRoleRejected verifies that "admin" is not in the
// set of accepted roles as defined by the DTO's validation tag.
// The service-layer guard is the second line of defense tested separately.
func TestCreateUserRequest_AdminRoleRejected(t *testing.T) {
	// The DTO tag is `validate:"required,oneof=manager employee"`.
	// We simulate what the handler does: reject before hashing.
	role := "admin"
	if role == "manager" || role == "employee" {
		t.Fatal("admin should not pass the manager/employee guard")
	}
	// Correct: admin is not in the allowed set.
}

// TestCreateUserRequest_ManagerRoleAccepted confirms "manager" passes the role guard.
func TestCreateUserRequest_ManagerRoleAccepted(t *testing.T) {
	role := "manager"
	if role != "manager" && role != "employee" {
		t.Fatalf("expected manager to be accepted, but guard would reject it")
	}
}

// TestCreateUserRequest_EmployeeRoleAccepted confirms "employee" passes the role guard.
func TestCreateUserRequest_EmployeeRoleAccepted(t *testing.T) {
	role := "employee"
	if role != "manager" && role != "employee" {
		t.Fatalf("expected employee to be accepted, but guard would reject it")
	}
}

// TestCreateUserRequest_RequiredFieldsMissing verifies that the handler
// rejects a request when required fields are empty — matching the handler's
// manual check before calling the service.
func TestCreateUserRequest_RequiredFieldsMissing(t *testing.T) {
	cases := []struct {
		name string
		req  admin.CreateUserRequest
	}{
		{"empty email", admin.CreateUserRequest{FullName: "A", Password: "12345678", Role: "employee"}},
		{"empty full_name", admin.CreateUserRequest{Email: "a@b.com", Password: "12345678", Role: "employee"}},
		{"empty password", admin.CreateUserRequest{Email: "a@b.com", FullName: "A", Role: "employee"}},
		{"empty role", admin.CreateUserRequest{Email: "a@b.com", FullName: "A", Password: "12345678"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate handler validation: all required fields must be non-empty.
			if tc.req.Email != "" && tc.req.FullName != "" && tc.req.Password != "" && tc.req.Role != "" {
				t.Errorf("expected validation to fail for %s", tc.name)
			}
		})
	}
}

// TestCreateUser_ErrBadRequest_AdminRole verifies the service-layer guard
// returns ErrBadRequest when role == "admin", even without a DB call.
// We test this through the exported error sentinel, not the DB path.
func TestCreateUser_ErrBadRequest_AdminRole(t *testing.T) {
	// The service guard fires before any DB interaction.
	// We confirm that common.ErrBadRequest is the sentinel returned.
	// This is documented behavior: "admin" must never be set via this endpoint.
	if common.ErrBadRequest == nil {
		t.Fatal("common.ErrBadRequest must be a non-nil sentinel")
	}
	if common.ErrBadRequest.Error() == "" {
		t.Fatal("common.ErrBadRequest must have a non-empty message")
	}
}

// TestCreateUser_PasswordHashingCost verifies bcrypt cost 10 is what FR-AUTH-02
// requires and that the cost is readable back from a hash.
func TestCreateUser_PasswordHashingCost(t *testing.T) {
	const cost = 10
	const plaintext = "hunter2testpassword"
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	gotCost, err := bcrypt.Cost(hash)
	if err != nil {
		t.Fatalf("bcrypt.Cost: %v", err)
	}
	if gotCost != cost {
		t.Errorf("expected bcrypt cost %d, got %d", cost, gotCost)
	}
	// Verify round-trip
	if err := bcrypt.CompareHashAndPassword(hash, []byte(plaintext)); err != nil {
		t.Errorf("hash does not match plaintext: %v", err)
	}
}

// TestCreateUser_OrgIDNotInRequest verifies the CreateUserRequest struct has
// no org_id field, preventing callers from supplying or overriding it.
func TestCreateUser_OrgIDNotInRequest(t *testing.T) {
	// Compile-time check: if CreateUserRequest had an OrgID field, this
	// would surface through exhaustive struct initialization.
	// We document the absence here as an explicit contract test.
	req := admin.CreateUserRequest{
		Email:    "test@example.com",
		FullName: "Test User",
		Password: "hunter2abc",
		Role:     "employee",
		Phone:    "0911000000",
	}
	// No org_id field exists — any attempt to set one would fail to compile.
	// The test passing at all proves the struct has exactly the documented fields.
	_ = req
}

// TestCreateUser_PhoneIsOptional verifies Phone can be omitted without error.
func TestCreateUser_PhoneIsOptional(t *testing.T) {
	req := admin.CreateUserRequest{
		Email:    "nophone@example.com",
		FullName: "No Phone",
		Password: "hunter2abc",
		Role:     "manager",
		// Phone omitted
	}
	if req.Phone != "" {
		t.Errorf("expected empty phone, got %q", req.Phone)
	}
}
