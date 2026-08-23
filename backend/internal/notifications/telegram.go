package notifications

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TelegramSender handles sending messages via the Telegram Bot API.
type TelegramSender struct {
	botToken string
	db       *gorm.DB
}

func NewTelegramSender(db *gorm.DB) *TelegramSender {
	return &TelegramSender{
		botToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		db:       db,
	}
}

// Send sends a Telegram message to a user if they have an active subscription.
// Always persists to the notifications table first, independent of Telegram delivery.
// FR-NOTIFY-02/03
func (t *TelegramSender) Send(orgID, userID uuid.UUID, notifType, message string) {
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

// NotifyTimesheetSubmitted triggers on DRAFT → SUBMITTED — FR-NOTIFY-02
func (t *TelegramSender) NotifyTimesheetSubmitted(orgID, userID uuid.UUID) {
	t.Send(orgID, userID, "timesheet_submitted",
		"📋 Your timesheet has been submitted and is awaiting manager review.")
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
