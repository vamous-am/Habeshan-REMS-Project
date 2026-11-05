package common

import (
	"regexp"
	"strings"
)

// EthiopianPhoneMessage is the user-facing validation message shared by the
// API responses for an invalid Ethiopian mobile number.
const EthiopianPhoneMessage = "Enter a valid Ethiopian phone number"

// Ethiopian mobile numbers are accepted in either local form
// (09XXXXXXXX or 07XXXXXXXX) or international form
// (+2519XXXXXXXX or +2517XXXXXXXX).
var ethiopianPhonePattern = regexp.MustCompile(`^(?:0[79][0-9]{8}|\+251[79][0-9]{8})$`)

// ValidateEthiopianPhone validates an optional Ethiopian mobile number.
// An empty value is valid because phone is optional in the user flows that
// use this helper.
func ValidateEthiopianPhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil
	}
	if !ethiopianPhonePattern.MatchString(phone) {
		return ErrInvalidPhone
	}
	return nil
}
