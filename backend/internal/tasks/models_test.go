package tasks

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
)

func TestTaskTimeLog_TenantScopedEmbedded(t *testing.T) {
	timeLog := TaskTimeLog{}
	val := reflect.ValueOf(timeLog)
	typ := val.Type()

	// 1. Verify TenantScoped is an embedded field
	field, found := typ.FieldByName("TenantScoped")
	if !found || !field.Anonymous {
		t.Fatalf("expected TaskTimeLog to embed common.TenantScoped anonymously")
	}

	// 2. Verify OrgID field exists and has UUID type
	orgIDField, found := typ.FieldByName("OrgID")
	if !found {
		t.Fatalf("expected TaskTimeLog to have promoted OrgID field")
	}
	if orgIDField.Type != reflect.TypeOf(uuid.UUID{}) {
		t.Fatalf("expected OrgID to be of type uuid.UUID, got %v", orgIDField.Type)
	}

	// 3. Verify JSON serialization includes org_id
	orgID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()
	recUUID := uuid.New()
	now := time.Now().UTC()

	entry := TaskTimeLog{
		TenantScoped: common.TenantScoped{OrgID: orgID},
		TaskID:       taskID,
		UserID:       userID,
		StartedAt:    now,
		SyncStatus:   SyncSyncedVerified,
		DeviceHash:   "test-device-hash",
		RecordUUID:   recUUID,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal TaskTimeLog: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed["org_id"] != orgID.String() {
		t.Errorf("expected JSON to contain org_id = %s, got %v", orgID.String(), parsed["org_id"])
	}
}
