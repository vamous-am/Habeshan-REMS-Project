package notifications

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNoManagerFound = errors.New("no supervising manager found for employee")

// ManagerResolver resolves supervising managers and submitter info.
type ManagerResolver interface {
	GetSupervisingManagers(orgID, employeeID uuid.UUID) ([]uuid.UUID, error)
	ResolveSubmitterName(orgID, employeeID uuid.UUID) string
}

// TelegramSender handles sending messages via the Telegram Bot API.
type TelegramSender struct {
	botToken string
	db       *gorm.DB
	resolver ManagerResolver
	sendFn   func(orgID, userID uuid.UUID, notifType, message string)
}

func NewTelegramSender(db *gorm.DB) *TelegramSender {
	return &TelegramSender{
		botToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		db:       db,
	}
}

// SetResolver allows injecting a custom ManagerResolver (primarily for testing).
func (t *TelegramSender) SetResolver(r ManagerResolver) {
	t.resolver = r
}

// SetSendFn allows overriding the send implementation (primarily for testing).
func (t *TelegramSender) SetSendFn(fn func(orgID, userID uuid.UUID, notifType, message string)) {
	t.sendFn = fn
}

// GetSupervisingManagers queries the database for all distinct managers supervising the employee within the org.
// Applies tenant isolation: joins teams, team_members, and verified users strictly matching org_id.
func (t *TelegramSender) GetSupervisingManagers(orgID, employeeID uuid.UUID) ([]uuid.UUID, error) {
	if t.resolver != nil {
		return t.resolver.GetSupervisingManagers(orgID, employeeID)
	}
	if t.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var managerIDs []uuid.UUID
	err := t.db.Table("team_members tm").
		Joins("JOIN teams t ON t.id = tm.team_id").
		Joins("JOIN users u ON u.id = tm.user_id AND u.org_id = t.org_id AND u.deleted_at IS NULL").
		Joins("JOIN users m ON m.id = t.manager_id AND m.org_id = t.org_id AND m.deleted_at IS NULL").
		Where("tm.user_id = ? AND t.org_id = ? AND t.manager_id IS NOT NULL", employeeID, orgID).
		Distinct("t.manager_id").
		Pluck("t.manager_id", &managerIDs).Error
	if err != nil {
		return nil, err
	}

	// In-memory deduplication defense-in-depth
	seen := make(map[uuid.UUID]bool)
	var unique []uuid.UUID
	for _, id := range managerIDs {
		if id != uuid.Nil && !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	return unique, nil
}

// ResolveSubmitterName returns the employee's display name or email within org_id.
// Name lookup is optional context; errors are logged and a generic fallback is used.
func (t *TelegramSender) ResolveSubmitterName(orgID, employeeID uuid.UUID) string {
	if t.resolver != nil {
		return t.resolver.ResolveSubmitterName(orgID, employeeID)
	}
	if t.db == nil {
		return ""
	}

	var submitter struct {
		FullName string
		Email    string
	}
	if err := t.db.Table("users").
		Select("full_name, email").
		Where("id = ? AND org_id = ? AND deleted_at IS NULL", employeeID, orgID).
		Scan(&submitter).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("⚠️ failed to resolve submitter name for employee %s (org %s): %v", employeeID, orgID, err)
		}
		return ""
	}

	if name := strings.TrimSpace(submitter.FullName); name != "" {
		return name
	}
	if email := strings.TrimSpace(submitter.Email); email != "" {
		return email
	}
	return ""
}

// FormatTimesheetSubmittedMessage generates the human-readable notification text.
func FormatTimesheetSubmittedMessage(submitterName string) string {
	if submitterName != "" {
		return fmt.Sprintf("📋 Timesheet submitted by %s is awaiting your review.", submitterName)
	}
	return "📋 A team timesheet has been submitted and is awaiting your review."
}

// Send sends a Telegram message to a user if they have an active subscription.
// Always persists to the notifications table first, independent of Telegram delivery.
// Best-effort delivery: failures are logged and do not roll back the calling business operation.
// FR-NOTIFY-02/03
func (t *TelegramSender) Send(orgID, userID uuid.UUID, notifType, message string) {
	if t.sendFn != nil {
		t.sendFn(orgID, userID, notifType, message)
		return
	}

	// 1. Always persist to DB first — FR-NOTIFY-05
	svc := NewService(t.db)
	if err := svc.CreateNotification(orgID, userID, notifType, message); err != nil {
		log.Printf("❌ failed to persist notification: %v", err)
	}

	// 2. Check if user has an active Telegram subscription
	sub, err := svc.GetSubscriber(userID)
	if err != nil || sub == nil || !sub.IsActive {
		return // no Telegram linked or inactive — notification already persisted above
	}

	// 3. Send via Telegram Bot API
	if t.botToken == "" {
		log.Println("⚠️  TELEGRAM_BOT_TOKEN not set, skipping Telegram delivery")
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id": {sub.ChatID},
		"text":    {message},
	})
	if err != nil {
		log.Printf("❌ telegram send failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("✅ telegram notification sent to user %s", userID)
}

// NotifyTimesheetSubmitted triggers on DRAFT → SUBMITTED — FR-NOTIFY-02.
// Resolves supervising managers for the employee and routes notifications to the managers.
//
// Delivery Policy:
// - Supervising managers are deduplicated and nil IDs are filtered out at this boundary.
// - Returns ErrNoManagerFound if no managers exist (does not fallback to admin or submitter).
// - Individual delivery failures are logged by Send and do not abort notifying other managers.
func (t *TelegramSender) NotifyTimesheetSubmitted(orgID, employeeID uuid.UUID) error {
	rawManagers, err := t.GetSupervisingManagers(orgID, employeeID)
	if err != nil {
		log.Printf("❌ failed to resolve supervising managers for employee %s (org %s): %v", employeeID, orgID, err)
		return err
	}

	// Defense-in-depth: deduplicate and filter uuid.Nil at final routing boundary
	seen := make(map[uuid.UUID]bool)
	var uniqueManagers []uuid.UUID
	for _, id := range rawManagers {
		if id != uuid.Nil && !seen[id] {
			seen[id] = true
			uniqueManagers = append(uniqueManagers, id)
		}
	}

	if len(uniqueManagers) == 0 {
		log.Printf("⚠️ timesheet submitted by employee %s (org %s) but no supervising manager was found", employeeID, orgID)
		return ErrNoManagerFound
	}

	submitterName := t.ResolveSubmitterName(orgID, employeeID)
	message := FormatTimesheetSubmittedMessage(submitterName)

	for _, managerID := range uniqueManagers {
		t.Send(orgID, managerID, "timesheet_submitted", message)
	}
	return nil
}

// NotifyTimesheetApproved triggers on SUBMITTED → APPROVED — FR-NOTIFY-02
func (t *TelegramSender) NotifyTimesheetApproved(orgID, userID uuid.UUID) {
	t.Send(orgID, userID, "timesheet_approved",
		"✅ Your timesheet has been approved!")
}

// NotifyTimesheetRejected triggers on SUBMITTED → REJECTED — FR-NOTIFY-03
func (t *TelegramSender) NotifyTimesheetRejected(orgID, userID uuid.UUID, reason string) {
	t.Send(orgID, userID, "timesheet_rejected",
		fmt.Sprintf("❌ Your timesheet was rejected. Reason: %s", reason))
}
