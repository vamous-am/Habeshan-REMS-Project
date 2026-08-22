package migrations

import (
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/tasks"
	"gorm.io/gorm"
)

// ─── Deterministic seed UUIDs ────────────────────────────────────────────────
//
// Fixed UUIDs so manual testing is predictable:
//   - org      : 00000000-0000-0000-0000-000000000001
//   - admin    : 11111111-1111-1111-1111-111111111111
//   - manager  : 22222222-2222-2222-2222-222222222222
//   - employee alice : 33333333-3333-3333-3333-333333333333
//   - employee bob   : 44444444-4444-4444-4444-444444444444
//   - team     : 55555555-5555-5555-5555-555555555555
//   - tasks    : aaaaaaaa-...-0001 through 0010
//   - time_logs: bbbbbbbb-...-0001 onward
//
// For browser testing set localStorage:
//   x-user-id : <any of the user UUIDs above>
//   x-org-id  : 00000000-0000-0000-0000-000000000001

var (
	seedOrgID     = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	seedAdminID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	seedManagerID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	seedAliceID   = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	seedBobID     = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	seedTeamID    = uuid.MustParse("55555555-5555-5555-5555-555555555555")
)

func taskUUID(n byte) uuid.UUID {
	b := []byte{0xAA, 0xAA, 0xAA, 0xAA, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, n}
	return uuid.UUID(b)
}

func timeLogUUID(n byte) uuid.UUID {
	b := []byte{0xBB, 0xBB, 0xBB, 0xBB, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, n}
	return uuid.UUID(b)
}

func recordUUID(n byte) uuid.UUID {
	b := []byte{0xCC, 0xCC, 0xCC, 0xCC, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, n}
	return uuid.UUID(b)
}

// MigrateSeedTasks inserts deterministic seed data for the Tasks feature.
// Uses FirstOrCreate-style logic so the seed is idempotent when re-run.
func MigrateSeedTasks(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		pastDay := now.AddDate(0, 0, -7)
		futureDay := now.AddDate(0, 0, 14)
		tomorrow := now.AddDate(0, 0, 1)
		yesterday := now.AddDate(0, 0, -1)
		longAgo := now.AddDate(0, 0, -30)

		// ── 1. Organization ──────────────────────────────────────────────────
		org := auth.Organization{
			ID:                 common.ID{ID: seedOrgID},
			Name:               "Habeshan Tech Ltd",
			Currency:           "ETB",
			Timezone:           "Africa/Addis_Ababa",
			SeatCount:          4,
			SubscriptionStatus: auth.SubscriptionTrial,
			Timestamps:         common.Timestamps{CreatedAt: longAgo, UpdatedAt: longAgo},
		}
		if err := upsertOrg(tx, &org); err != nil {
			return err
		}

		// ── 2. Users ─────────────────────────────────────────────────────────
		pwHash := "$2a$10$placeholder.password.hash.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

		users := []auth.User{
			{
				ID:           common.ID{ID: seedAdminID},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Email:        "admin@habeshan.local", PasswordHash: pwHash,
				FullName: "Admin User", Phone: "+251911000001",
				Role: auth.RoleAdmin, Status: auth.UserStatusActive,
				Timestamps: common.Timestamps{CreatedAt: longAgo, UpdatedAt: longAgo},
			},
			{
				ID:           common.ID{ID: seedManagerID},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Email:        "manager@habeshan.local", PasswordHash: pwHash,
				FullName: "Manager User", Phone: "+251911000002",
				Role: auth.RoleManager, Status: auth.UserStatusActive,
				Timestamps: common.Timestamps{CreatedAt: longAgo, UpdatedAt: longAgo},
			},
			{
				ID:           common.ID{ID: seedAliceID},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Email:        "alice@habeshan.local", PasswordHash: pwHash,
				FullName: "Alice Employee", Phone: "+251911000003",
				Role: auth.RoleEmployee, Status: auth.UserStatusActive,
				Timestamps: common.Timestamps{CreatedAt: longAgo, UpdatedAt: longAgo},
			},
			{
				ID:           common.ID{ID: seedBobID},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Email:        "bob@habeshan.local", PasswordHash: pwHash,
				FullName: "Bob Employee", Phone: "+251911000004",
				Role: auth.RoleEmployee, Status: auth.UserStatusActive,
				Timestamps: common.Timestamps{CreatedAt: longAgo, UpdatedAt: longAgo},
			},
		}
		for i := range users {
			if err := upsertUser(tx, &users[i]); err != nil {
				return err
			}
		}

		// ── 3. Team + TeamMembers ────────────────────────────────────────────
		team := auth.Team{
			ID:           common.ID{ID: seedTeamID},
			TenantScoped: common.TenantScoped{OrgID: seedOrgID},
			Name:         "Engineering Team",
			ManagerID:    seedManagerID,
			Timestamps:   common.Timestamps{CreatedAt: longAgo, UpdatedAt: longAgo},
		}
		if err := upsertTeam(tx, &team); err != nil {
			return err
		}
		if err := upsertTeamMember(tx, seedTeamID, seedAliceID); err != nil {
			return err
		}
		if err := upsertTeamMember(tx, seedTeamID, seedBobID); err != nil {
			return err
		}

		// ── 4. Tasks ─────────────────────────────────────────────────────────
		desc := func(s string) *string { return &s }

		seedTasks := []tasks.Task{
			// 01: in_progress, high, Alice, due tomorrow
			{
				ID:           common.ID{ID: taskUUID(1)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Design new dashboard layout",
				Description:  desc("Create mockups for the dashboard redesign including KPIs, charts, and responsive breakpoints."),
				Priority:     tasks.PriorityHigh,
				Status:       tasks.StatusInProgress,
				CreatedBy:    seedManagerID,
				DueDate:      &tomorrow,
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 02: to_do, medium, Alice, due in 14 days
			{
				ID:           common.ID{ID: taskUUID(2)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Write API documentation for tasks endpoints",
				Description:  desc("Document all /api/v1/tasks endpoints with request/response examples and auth requirements."),
				Priority:     tasks.PriorityMedium,
				Status:       tasks.StatusToDo,
				CreatedBy:    seedManagerID,
				DueDate:      &futureDay,
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 03: paused, high, Bob, due yesterday (overdue)
			{
				ID:           common.ID{ID: taskUUID(3)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Fix login page authentication bug",
				Description:  desc("Users report being logged out unexpectedly after 2-3 minutes. Investigate session cookie settings."),
				Priority:     tasks.PriorityHigh,
				Status:       tasks.StatusPaused,
				CreatedBy:    seedManagerID,
				DueDate:      &yesterday,
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 04: blocked, low, Alice + Bob, due in 21 days
			{
				ID:           common.ID{ID: taskUUID(4)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Integrate Telegram notifications",
				Description:  desc("Blocked on Telegram bot token provisioning. Once obtained wire up to notifications service."),
				Priority:     tasks.PriorityLow,
				Status:       tasks.StatusBlocked,
				CreatedBy:    seedManagerID,
				DueDate:      timePtr(now.AddDate(0, 0, 21)),
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 05: completed, medium, Alice, due in past (on time)
			{
				ID:           common.ID{ID: taskUUID(5)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Implement login page UI",
				Description:  desc("Build responsive login form with email/password fields and error states."),
				Priority:     tasks.PriorityMedium,
				Status:       tasks.StatusCompleted,
				CreatedBy:    seedManagerID,
				DueDate:      timePtr(now.AddDate(0, 0, -3)),
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 06: to_do, high, UNASSIGNED, due yesterday (overdue)
			{
				ID:           common.ID{ID: taskUUID(6)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Fix production CORS configuration",
				Description:  desc("Wildcard origin is currently allowed. Restrict to the production frontend domain."),
				Priority:     tasks.PriorityHigh,
				Status:       tasks.StatusToDo,
				CreatedBy:    seedAdminID,
				DueDate:      &yesterday,
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 07: in_progress, medium, Bob, due in 3 days, timer running
			{
				ID:           common.ID{ID: taskUUID(7)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Write unit tests for task service layer",
				Description:  desc("Cover status transitions, assignment rules, and timer state machine with tests."),
				Priority:     tasks.PriorityMedium,
				Status:       tasks.StatusInProgress,
				CreatedBy:    seedManagerID,
				DueDate:      timePtr(now.AddDate(0, 0, 3)),
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 08: to_do, low, Alice, due in 30 days
			{
				ID:           common.ID{ID: taskUUID(8)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Refactor seed migration into reusable helper",
				Description:  desc("Low priority cleanup. Extract helpers so future features can seed data consistently."),
				Priority:     tasks.PriorityLow,
				Status:       tasks.StatusToDo,
				CreatedBy:    seedManagerID,
				DueDate:      timePtr(now.AddDate(0, 0, 30)),
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 09: blocked, high, Manager
			{
				ID:           common.ID{ID: taskUUID(9)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Procure staging environment hardware",
				Description:  desc("Blocked on finance approval. Once PO is issued proceed with procurement."),
				Priority:     tasks.PriorityHigh,
				Status:       tasks.StatusBlocked,
				CreatedBy:    seedAdminID,
				DueDate:      timePtr(now.AddDate(0, 0, 7)),
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
			// 10: completed, medium, Bob, completed late
			{
				ID:           common.ID{ID: taskUUID(10)},
				TenantScoped: common.TenantScoped{OrgID: seedOrgID},
				Title:        "Fix date formatting in task list",
				Description:  desc("Due dates showed server timezone instead of organization timezone."),
				Priority:     tasks.PriorityMedium,
				Status:       tasks.StatusCompleted,
				CreatedBy:    seedManagerID,
				DueDate:      timePtr(now.AddDate(0, 0, -10)),
				Timestamps:   common.Timestamps{CreatedAt: pastDay, UpdatedAt: pastDay},
			},
		}
		for i := range seedTasks {
			if err := upsertTask(tx, &seedTasks[i]); err != nil {
				return err
			}
		}

		// ── 5. Task assignments ──────────────────────────────────────────────
		assignments := []tasks.TaskAssignment{
			{TaskID: taskUUID(1), UserID: seedAliceID, AssignedAt: pastDay},
			{TaskID: taskUUID(2), UserID: seedAliceID, AssignedAt: pastDay},
			{TaskID: taskUUID(3), UserID: seedBobID, AssignedAt: pastDay},
			{TaskID: taskUUID(4), UserID: seedAliceID, AssignedAt: pastDay},
			{TaskID: taskUUID(4), UserID: seedBobID, AssignedAt: pastDay},
			{TaskID: taskUUID(5), UserID: seedAliceID, AssignedAt: pastDay},
			{TaskID: taskUUID(7), UserID: seedBobID, AssignedAt: pastDay},
			{TaskID: taskUUID(8), UserID: seedAliceID, AssignedAt: pastDay},
			{TaskID: taskUUID(9), UserID: seedManagerID, AssignedAt: pastDay},
			{TaskID: taskUUID(10), UserID: seedBobID, AssignedAt: pastDay},
		}
		for i := range assignments {
			if err := upsertAssignment(tx, &assignments[i]); err != nil {
				return err
			}
		}

		// ── 6. Task time logs ────────────────────────────────────────────────
		devHash := "seed-device-hash-012345678901234567890123456789012345678901234567890123"

		timeLogs := []tasks.TaskTimeLog{
			// Task 1 (Alice): Start → Pause (Power Outage 75 min) → Resume → Stop (75 min) = 150 min total
			{
				ID: common.ID{ID: timeLogUUID(1)}, TaskID: taskUUID(1), UserID: seedAliceID,
				StartedAt:       now.AddDate(0, 0, -1).Add(-3 * time.Hour),
				StoppedAt:       timePtr(now.AddDate(0, 0, -1).Add(-2*time.Hour - 45*time.Minute)),
				DurationMinutes: intPtr(75),
				PauseReason:     strPtr("Power Outage"),
				SyncStatus:      tasks.SyncSyncedVerified,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(1),
				Timestamps:      common.Timestamps{CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.AddDate(0, 0, -1)},
			},
			{
				ID: common.ID{ID: timeLogUUID(2)}, TaskID: taskUUID(1), UserID: seedAliceID,
				StartedAt:       now.AddDate(0, 0, -1).Add(-2 * time.Hour),
				StoppedAt:       timePtr(now.AddDate(0, 0, -1).Add(-45 * time.Minute)),
				DurationMinutes: intPtr(75),
				SyncStatus:      tasks.SyncSyncedVerified,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(2),
				Timestamps:      common.Timestamps{CreatedAt: now.AddDate(0, 0, -1), UpdatedAt: now.AddDate(0, 0, -1)},
			},
			// Task 3 (Bob): Start → Paused with "Personal Break" — closed (100 min)
			{
				ID: common.ID{ID: timeLogUUID(3)}, TaskID: taskUUID(3), UserID: seedBobID,
				StartedAt:       now.AddDate(0, 0, -2).Add(-4 * time.Hour),
				StoppedAt:       timePtr(now.AddDate(0, 0, -2).Add(-2*time.Hour - 20*time.Minute)),
				DurationMinutes: intPtr(100),
				PauseReason:     strPtr("Personal Break"),
				SyncStatus:      tasks.SyncSyncedVerified,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(3),
				Timestamps:      common.Timestamps{CreatedAt: now.AddDate(0, 0, -2), UpdatedAt: now.AddDate(0, 0, -2)},
			},
			// Task 5 (Alice): Completed — 2 segments 240 min each
			{
				ID: common.ID{ID: timeLogUUID(4)}, TaskID: taskUUID(5), UserID: seedAliceID,
				StartedAt:       now.AddDate(0, 0, -6).Add(-6 * time.Hour),
				StoppedAt:       timePtr(now.AddDate(0, 0, -6).Add(-2 * time.Hour)),
				DurationMinutes: intPtr(240),
				SyncStatus:      tasks.SyncSyncedVerified,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(4),
				Timestamps:      common.Timestamps{CreatedAt: now.AddDate(0, 0, -6), UpdatedAt: now.AddDate(0, 0, -6)},
			},
			{
				ID: common.ID{ID: timeLogUUID(5)}, TaskID: taskUUID(5), UserID: seedAliceID,
				StartedAt:       now.AddDate(0, 0, -5).Add(-7 * time.Hour),
				StoppedAt:       timePtr(now.AddDate(0, 0, -5).Add(-3 * time.Hour)),
				DurationMinutes: intPtr(240),
				SyncStatus:      tasks.SyncSyncedVerified,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(5),
				Timestamps:      common.Timestamps{CreatedAt: now.AddDate(0, 0, -5), UpdatedAt: now.AddDate(0, 0, -5)},
			},
			// Task 7 (Bob): RUNNING — open segment (stopped_at = NULL, ~35 min elapsed)
			{
				ID: common.ID{ID: timeLogUUID(6)}, TaskID: taskUUID(7), UserID: seedBobID,
				StartedAt:       now.Add(-35 * time.Minute),
				StoppedAt:       nil,
				DurationMinutes: nil,
				SyncStatus:      tasks.SyncPendingSync,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(6),
				Timestamps:      common.Timestamps{CreatedAt: now.Add(-35 * time.Minute), UpdatedAt: now.Add(-35 * time.Minute)},
			},
			// Task 10 (Bob): Completed late — 2 segments 180 min each
			{
				ID: common.ID{ID: timeLogUUID(7)}, TaskID: taskUUID(10), UserID: seedBobID,
				StartedAt:       now.AddDate(0, 0, -8).Add(-5 * time.Hour),
				StoppedAt:       timePtr(now.AddDate(0, 0, -8).Add(-2 * time.Hour)),
				DurationMinutes: intPtr(180),
				SyncStatus:      tasks.SyncSyncedVerified,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(7),
				Timestamps:      common.Timestamps{CreatedAt: now.AddDate(0, 0, -8), UpdatedAt: now.AddDate(0, 0, -8)},
			},
			{
				ID: common.ID{ID: timeLogUUID(8)}, TaskID: taskUUID(10), UserID: seedBobID,
				StartedAt:       now.AddDate(0, 0, -7).Add(-6 * time.Hour),
				StoppedAt:       timePtr(now.AddDate(0, 0, -7).Add(-3 * time.Hour)),
				DurationMinutes: intPtr(180),
				SyncStatus:      tasks.SyncSyncedVerified,
				DeviceHash:      devHash,
				RecordUUID:      recordUUID(8),
				Timestamps:      common.Timestamps{CreatedAt: now.AddDate(0, 0, -7), UpdatedAt: now.AddDate(0, 0, -7)},
			},
		}
		for i := range timeLogs {
			if err := upsertTimeLog(tx, &timeLogs[i]); err != nil {
				return err
			}
		}

		return nil
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func timePtr(t time.Time) *time.Time { return &t }
func intPtr(i int) *int              { return &i }
func strPtr(s string) *string        { return &s }

func upsertOrg(db *gorm.DB, org *auth.Organization) error {
	var existing auth.Organization
	err := db.Where("id = ?", org.ID.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(org).Error
	}
	return err
}

func upsertUser(db *gorm.DB, user *auth.User) error {
	var existing auth.User
	err := db.Unscoped().Where("id = ?", user.ID.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(user).Error
	}
	return err
}

func upsertTeam(db *gorm.DB, team *auth.Team) error {
	var existing auth.Team
	err := db.Where("id = ?", team.ID.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(team).Error
	}
	return err
}

func upsertTeamMember(db *gorm.DB, teamID, userID uuid.UUID) error {
	var existing auth.TeamMember
	err := db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(&auth.TeamMember{TeamID: teamID, UserID: userID}).Error
	}
	return err
}

func upsertTask(db *gorm.DB, task *tasks.Task) error {
	var existing tasks.Task
	err := db.Where("id = ?", task.ID.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(task).Error
	}
	return err
}

func upsertAssignment(db *gorm.DB, a *tasks.TaskAssignment) error {
	var existing tasks.TaskAssignment
	err := db.Where("task_id = ? AND user_id = ?", a.TaskID, a.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(a).Error
	}
	return err
}

func upsertTimeLog(db *gorm.DB, l *tasks.TaskTimeLog) error {
	var existing tasks.TaskTimeLog
	err := db.Where("id = ?", l.ID.ID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(l).Error
	}
	return err
}
