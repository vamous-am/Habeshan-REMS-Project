package auth_test

import (
	"errors"
	"testing"

	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
)

func TestRegister_InvalidPhoneRejected(t *testing.T) {
	svc := auth.NewService(nil)
	_, err := svc.Register(auth.RegisterRequest{
		OrgName:  "Test Organization",
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "hunter2abc",
		Phone:    "091122334455",
	})
	if !errors.Is(err, common.ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}
}
