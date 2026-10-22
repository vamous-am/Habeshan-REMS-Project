package attendance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func computeTestHash(recordUUID, userID, actionType, timestamp, sep string) string {
	raw := fmt.Sprintf("%s%s%s%s%s%s%s", recordUUID, sep, userID, sep, actionType, sep, timestamp)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func TestVerifyDeviceHash_PipeFormat(t *testing.T) {
	recUUID := uuid.New().String()
	usrID := uuid.New().String()
	action := "CLOCK_IN"
	ts := "2026-10-22T14:30:00.123Z"

	validHash := computeTestHash(recUUID, usrID, action, ts, "|")

	if !VerifyDeviceHash(recUUID, usrID, action, ts, validHash) {
		t.Fatalf("expected pipe-delimited hash to be accepted")
	}
}

func TestVerifyDeviceHash_LegacyColonFormat(t *testing.T) {
	recUUID := uuid.New().String()
	usrID := uuid.New().String()
	action := "CLOCK_OUT"
	ts := "2026-10-22T17:45:00Z"

	legacyHash := computeTestHash(recUUID, usrID, action, ts, ":")

	if !VerifyDeviceHash(recUUID, usrID, action, ts, legacyHash) {
		t.Fatalf("expected legacy colon-delimited hash to be accepted for backward compatibility")
	}
}

func TestVerifyDeviceHash_TamperedPayloadRejected(t *testing.T) {
	recUUID := uuid.New().String()
	usrID := uuid.New().String()
	action := "CLOCK_IN"
	ts := "2026-10-22T14:30:00Z"

	validHash := computeTestHash(recUUID, usrID, action, ts, "|")

	// 1. Altered UUID
	tamperedUUID := uuid.New().String()
	if VerifyDeviceHash(tamperedUUID, usrID, action, ts, validHash) {
		t.Errorf("expected altered UUID to fail verification")
	}

	// 2. Altered user ID
	tamperedUserID := uuid.New().String()
	if VerifyDeviceHash(recUUID, tamperedUserID, action, ts, validHash) {
		t.Errorf("expected altered user ID to fail verification")
	}

	// 3. Altered action type
	if VerifyDeviceHash(recUUID, usrID, "CLOCK_OUT", ts, validHash) {
		t.Errorf("expected altered action type to fail verification")
	}

	// 4. Altered timestamp
	if VerifyDeviceHash(recUUID, usrID, action, "2026-10-22T14:30:01Z", validHash) {
		t.Errorf("expected altered timestamp to fail verification")
	}

	// 5. Corrupted hash string
	if VerifyDeviceHash(recUUID, usrID, action, ts, "0000000000000000000000000000000000000000000000000000000000000000") {
		t.Errorf("expected corrupted hash to fail verification")
	}
}

func TestVerifyDeviceHash_TimestampPrecisionSensitivity(t *testing.T) {
	recUUID := uuid.New().String()
	usrID := uuid.New().String()
	action := "CLOCK_IN"

	// Hash was generated with subsecond precision
	tsSigned := "2026-10-22T14:00:00.000Z"
	hash := computeTestHash(recUUID, usrID, action, tsSigned, "|")

	// Payload sent without the fractional component must be rejected because bytes differ
	tsSent := "2026-10-22T14:00:00Z"
	if VerifyDeviceHash(recUUID, usrID, action, tsSent, hash) {
		t.Errorf("expected different timestamp byte string to fail verification even if representing same instant")
	}
}

func TestVerifyDeviceHash_EmptyParameters(t *testing.T) {
	recUUID := uuid.New().String()
	usrID := uuid.New().String()
	action := "CLOCK_IN"
	ts := "2026-10-22T14:00:00Z"
	hash := computeTestHash(recUUID, usrID, action, ts, "|")

	if VerifyDeviceHash("", usrID, action, ts, hash) {
		t.Errorf("expected empty recordUUID to fail")
	}
	if VerifyDeviceHash(recUUID, "", action, ts, hash) {
		t.Errorf("expected empty userID to fail")
	}
	if VerifyDeviceHash(recUUID, usrID, "", ts, hash) {
		t.Errorf("expected empty actionType to fail")
	}
	if VerifyDeviceHash(recUUID, usrID, action, "", hash) {
		t.Errorf("expected empty timestamp to fail")
	}
	if VerifyDeviceHash(recUUID, usrID, action, ts, "") {
		t.Errorf("expected empty deviceHash to fail")
	}
}

func TestParseTimestamp_Formats(t *testing.T) {
	// Standard RFC3339
	t1, err := ParseTimestamp("2026-10-22T14:30:00Z")
	if err != nil || t1.Year() != 2026 || t1.Hour() != 14 {
		t.Fatalf("failed to parse standard RFC3339: %v", err)
	}

	// RFC3339 with timezone offset
	t2, err := ParseTimestamp("2026-10-22T17:30:00+03:00")
	if err != nil || t2.UTC().Hour() != 14 {
		t.Fatalf("failed to parse timezone offset into UTC: %v", err)
	}

	// RFC3339Nano with millisecond precision
	t3, err := ParseTimestamp("2026-10-22T14:30:00.500Z")
	if err != nil || t3.Nanosecond() != 500000000 {
		t.Fatalf("failed to parse RFC3339Nano: %v", err)
	}

	// Invalid format
	_, err = ParseTimestamp("invalid-date-string")
	if err == nil {
		t.Fatalf("expected error for invalid timestamp string")
	}
}

func TestValidateOnlineDrift_Boundaries(t *testing.T) {
	serverNow := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	maxDrift := 5 * time.Minute

	// Exactly at server time
	if !ValidateOnlineDrift(serverNow, serverNow, maxDrift) {
		t.Errorf("expected 0 drift to be valid")
	}

	// Exactly at positive drift boundary (+5m)
	boundaryPast := serverNow.Add(-5 * time.Minute)
	if !ValidateOnlineDrift(boundaryPast, serverNow, maxDrift) {
		t.Errorf("expected timestamp exactly 5 minutes in past to be valid")
	}

	// Exactly at negative drift boundary (-5m)
	boundaryFuture := serverNow.Add(5 * time.Minute)
	if !ValidateOnlineDrift(boundaryFuture, serverNow, maxDrift) {
		t.Errorf("expected timestamp exactly 5 minutes in future to be valid")
	}

	// Just inside boundary (4m 59s)
	insideBoundary := serverNow.Add(-4*time.Minute - 59*time.Second)
	if !ValidateOnlineDrift(insideBoundary, serverNow, maxDrift) {
		t.Errorf("expected timestamp within 5 minutes to be valid")
	}

	// Just outside boundary (5m 1s past)
	outsidePast := serverNow.Add(-5*time.Minute - 1*time.Second)
	if ValidateOnlineDrift(outsidePast, serverNow, maxDrift) {
		t.Errorf("expected timestamp 5m01s in past to be rejected")
	}

	// Just outside boundary (5m 1s future)
	outsideFuture := serverNow.Add(5*time.Minute + 1*time.Second)
	if ValidateOnlineDrift(outsideFuture, serverNow, maxDrift) {
		t.Errorf("expected timestamp 5m01s in future to be rejected")
	}
}

func TestConstantTimeCompare(t *testing.T) {
	s1 := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	s2 := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	s3 := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b856"
	s4 := "short"

	if !constantTimeCompare(s1, s2) {
		t.Errorf("expected identical hashes to compare true")
	}
	if constantTimeCompare(s1, s3) {
		t.Errorf("expected differing hashes to compare false")
	}
	if constantTimeCompare(s1, s4) {
		t.Errorf("expected different length strings to compare false")
	}
}
