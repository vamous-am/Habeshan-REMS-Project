package common_test

import (
	"errors"
	"testing"

	"github.com/habeshan-rems/backend/internal/common"
)

func TestValidateEthiopianPhone(t *testing.T) {
	valid := []string{
		"",
		"0912345678",
		"0712345678",
		"+251912345678",
		"+251712345678",
		" 0912345678 ",
	}
	for _, phone := range valid {
		t.Run("valid/"+phone, func(t *testing.T) {
			if err := common.ValidateEthiopianPhone(phone); err != nil {
				t.Fatalf("expected %q to be valid, got %v", phone, err)
			}
		})
	}

	invalid := []string{
		"091122334455",   // 12 digits: the case that previously passed
		"091234567",      // too short
		"09123456789",    // 11 digits
		"0812345678",     // unsupported local prefix
		"912345678",      // missing leading 0 or country code
		"+251812345678",  // unsupported mobile prefix after country code
		"+2510912345678", // country code plus local leading zero
		"+25191234567",   // international form too short
		"09123abc78",     // non-digit characters
	}
	for _, phone := range invalid {
		t.Run("invalid/"+phone, func(t *testing.T) {
			err := common.ValidateEthiopianPhone(phone)
			if !errors.Is(err, common.ErrInvalidPhone) {
				t.Fatalf("expected ErrInvalidPhone for %q, got %v", phone, err)
			}
		})
	}
}
