package notifications

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockResolver implements ManagerResolver for testing.
type mockResolver struct {
	managersFn func(orgID, employeeID uuid.UUID) ([]uuid.UUID, error)
	nameFn     func(orgID, employeeID uuid.UUID) string
}

func (m *mockResolver) GetSupervisingManagers(orgID, employeeID uuid.UUID) ([]uuid.UUID, error) {
	if m.managersFn != nil {
		return m.managersFn(orgID, employeeID)
	}
	return nil, nil
}

func (m *mockResolver) ResolveSubmitterName(orgID, employeeID uuid.UUID) string {
	if m.nameFn != nil {
		return m.nameFn(orgID, employeeID)
	}
	return ""
}

type sentNotification struct {
	orgID     uuid.UUID
	userID    uuid.UUID
	notifType string
	message   string
}

func TestNotifyTimesheetSubmitted_SingleManager(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()
	managerID := uuid.New()

	var sent []sentNotification
	sender := &TelegramSender{}
	sender.SetResolver(&mockResolver{
		managersFn: func(o, e uuid.UUID) ([]uuid.UUID, error) {
			if o == orgID && e == employeeID {
				return []uuid.UUID{managerID}, nil
			}
			return nil, nil
		},
		nameFn: func(o, e uuid.UUID) string {
			return "Jane Doe"
		},
	})
	sender.SetSendFn(func(o, u uuid.UUID, nt, msg string) {
		sent = append(sent, sentNotification{o, u, nt, msg})
	})

	err := sender.NotifyTimesheetSubmitted(orgID, employeeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sent) != 1 {
		t.Fatalf("expected exactly 1 notification sent, got %d", len(sent))
	}

	// Verify notification recipient is the MANAGER, NOT the employee
	if sent[0].userID != managerID {
		t.Errorf("expected recipient to be manager %s, got %s", managerID, sent[0].userID)
	}
	if sent[0].userID == employeeID {
		t.Errorf("recipient should NOT be the submitter employee")
	}

	if !strings.Contains(sent[0].message, "Jane Doe") {
		t.Errorf("expected message to mention Jane Doe, got %q", sent[0].message)
	}
}

func TestNotifyTimesheetSubmitted_DeduplicatesSameManagerAcrossTeams(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()
	managerID := uuid.New()

	var sent []sentNotification
	sender := &TelegramSender{}
	sender.SetResolver(&mockResolver{
		managersFn: func(o, e uuid.UUID) ([]uuid.UUID, error) {
			// Simulates query or membership returning duplicate manager ID from multiple teams
			return []uuid.UUID{managerID, managerID, managerID}, nil
		},
		nameFn: func(o, e uuid.UUID) string {
			return "Alice"
		},
	})
	sender.SetSendFn(func(o, u uuid.UUID, nt, msg string) {
		sent = append(sent, sentNotification{o, u, nt, msg})
	})

	err := sender.NotifyTimesheetSubmitted(orgID, employeeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Crucial assertion: exactly 1 notification sent despite 3 duplicate manager entries
	if len(sent) != 1 {
		t.Fatalf("expected exactly 1 notification sent after deduplication, got %d", len(sent))
	}
	if sent[0].userID != managerID {
		t.Errorf("expected recipient to be %s, got %s", managerID, sent[0].userID)
	}
}

func TestNotifyTimesheetSubmitted_FiltersNilManagerID(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()
	managerID := uuid.New()

	var sent []sentNotification
	sender := &TelegramSender{}
	sender.SetResolver(&mockResolver{
		managersFn: func(o, e uuid.UUID) ([]uuid.UUID, error) {
			// Returns uuid.Nil alongside valid manager
			return []uuid.UUID{uuid.Nil, managerID, uuid.Nil}, nil
		},
		nameFn: func(o, e uuid.UUID) string {
			return "Bob"
		},
	})
	sender.SetSendFn(func(o, u uuid.UUID, nt, msg string) {
		sent = append(sent, sentNotification{o, u, nt, msg})
	})

	err := sender.NotifyTimesheetSubmitted(orgID, employeeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// uuid.Nil entries must be filtered out; exactly 1 notification sent to real manager
	if len(sent) != 1 {
		t.Fatalf("expected exactly 1 notification sent (nil filtered), got %d", len(sent))
	}
	if sent[0].userID != managerID {
		t.Errorf("expected recipient to be manager %s, got %s", managerID, sent[0].userID)
	}
}

func TestNotifyTimesheetSubmitted_MultipleDistinctManagers(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()
	manager1 := uuid.New()
	manager2 := uuid.New()

	var sent []sentNotification
	sender := &TelegramSender{}
	sender.SetResolver(&mockResolver{
		managersFn: func(o, e uuid.UUID) ([]uuid.UUID, error) {
			return []uuid.UUID{manager1, manager2}, nil
		},
		nameFn: func(o, e uuid.UUID) string {
			return "Bob Smith"
		},
	})
	sender.SetSendFn(func(o, u uuid.UUID, nt, msg string) {
		sent = append(sent, sentNotification{o, u, nt, msg})
	})

	err := sender.NotifyTimesheetSubmitted(orgID, employeeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sent) != 2 {
		t.Fatalf("expected 2 notifications sent, got %d", len(sent))
	}

	recipients := map[uuid.UUID]bool{sent[0].userID: true, sent[1].userID: true}
	if !recipients[manager1] || !recipients[manager2] {
		t.Errorf("expected both manager1 and manager2 to receive notifications")
	}
	if recipients[employeeID] {
		t.Errorf("employee must not be among recipients")
	}
}

func TestNotifyTimesheetSubmitted_NoManagerFound(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()

	var sent []sentNotification
	sender := &TelegramSender{}
	sender.SetResolver(&mockResolver{
		managersFn: func(o, e uuid.UUID) ([]uuid.UUID, error) {
			return nil, nil // No manager found
		},
	})
	sender.SetSendFn(func(o, u uuid.UUID, nt, msg string) {
		sent = append(sent, sentNotification{o, u, nt, msg})
	})

	err := sender.NotifyTimesheetSubmitted(orgID, employeeID)
	if !errors.Is(err, ErrNoManagerFound) {
		t.Fatalf("expected ErrNoManagerFound, got %v", err)
	}

	// Must NOT send any notification to submitter or admins as fallback
	if len(sent) != 0 {
		t.Errorf("expected 0 notifications sent when no manager exists, got %d", len(sent))
	}
}

func TestNotifyTimesheetSubmitted_ResolverError(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()
	expectedErr := errors.New("db connection failure")

	var sent []sentNotification
	sender := &TelegramSender{}
	sender.SetResolver(&mockResolver{
		managersFn: func(o, e uuid.UUID) ([]uuid.UUID, error) {
			return nil, expectedErr
		},
	})
	sender.SetSendFn(func(o, u uuid.UUID, nt, msg string) {
		sent = append(sent, sentNotification{o, u, nt, msg})
	})

	err := sender.NotifyTimesheetSubmitted(orgID, employeeID)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if len(sent) != 0 {
		t.Errorf("expected 0 notifications sent on error, got %d", len(sent))
	}
}

func TestFormatTimesheetSubmittedMessage(t *testing.T) {
	// 1. With employee full name
	msg1 := FormatTimesheetSubmittedMessage("John Doe")
	if msg1 != "📋 Timesheet submitted by John Doe is awaiting your review." {
		t.Errorf("unexpected message with name: %q", msg1)
	}

	// 2. With fallback email
	msg2 := FormatTimesheetSubmittedMessage("john@example.com")
	if msg2 != "📋 Timesheet submitted by john@example.com is awaiting your review." {
		t.Errorf("unexpected message with email: %q", msg2)
	}

	// 3. With empty string fallback
	msg3 := FormatTimesheetSubmittedMessage("")
	if msg3 != "📋 A team timesheet has been submitted and is awaiting your review." {
		t.Errorf("unexpected generic message: %q", msg3)
	}
}

func TestNotifyTimesheetApprovedAndRejected_RecipientIsEmployee(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()

	var sent []sentNotification
	sender := &TelegramSender{}
	sender.SetSendFn(func(o, u uuid.UUID, nt, msg string) {
		sent = append(sent, sentNotification{o, u, nt, msg})
	})

	// 1. Approval notifies the employee
	sender.NotifyTimesheetApproved(orgID, employeeID)
	if len(sent) != 1 || sent[0].userID != employeeID || sent[0].notifType != "timesheet_approved" {
		t.Errorf("expected approval to notify employee %s", employeeID)
	}

	// 2. Rejection notifies the employee with reason
	sender.NotifyTimesheetRejected(orgID, employeeID, "missing overtime breakdown")
	if len(sent) != 2 || sent[1].userID != employeeID || sent[1].notifType != "timesheet_rejected" {
		t.Errorf("expected rejection to notify employee %s", employeeID)
	}
	if !strings.Contains(sent[1].message, "missing overtime breakdown") {
		t.Errorf("expected rejection reason in message, got %q", sent[1].message)
	}
}
