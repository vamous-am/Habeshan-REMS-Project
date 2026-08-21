/**
 * features/tasks/types.ts
 *
 * TypeScript types mirroring the Go DTOs in backend/internal/tasks/dto.go
 * and models in backend/internal/tasks/models.go.
 * Keep these in sync when the backend DTOs change.
 */

// ─── Enums ───────────────────────────────────────────────────────────────────

export type Priority = "high" | "medium" | "low";
export type TaskStatus =
  | "to_do"
  | "in_progress"
  | "paused"
  | "blocked"
  | "completed";
export type SyncStatus =
  | "offline_logged"
  | "pending_sync"
  | "synced_verified"
  | "rejected_tampered";
export type PauseReason =
  | "Power Outage"
  | "Personal Break"
  | "Commute/Travel"
  | "Waiting on Dependency"
  | "Other";

export const PAUSE_REASONS: PauseReason[] = [
  "Power Outage",
  "Personal Break",
  "Commute/Travel",
  "Waiting on Dependency",
  "Other",
];

export const STATUS_LABELS: Record<TaskStatus, string> = {
  to_do: "To Do",
  in_progress: "In Progress",
  paused: "Paused",
  blocked: "Blocked",
  completed: "Completed",
};

export const PRIORITY_LABELS: Record<Priority, string> = {
  high: "High",
  medium: "Medium",
  low: "Low",
};

// ─── Response shapes ──────────────────────────────────────────────────────────

export interface Task {
  id: string;
  org_id: string;
  title: string;
  description?: string | null;
  priority: Priority;
  status: TaskStatus;
  created_by: string;
  due_date?: string | null;
  created_at: string;
  updated_at: string;
  assigned_to?: string[];
}

export interface TimeLog {
  id: string;
  task_id: string;
  user_id: string;
  started_at: string;
  stopped_at?: string | null;
  duration_minutes?: number | null;
  /** null when caller is not permitted to see it (FR-TASK-07) */
  pause_reason?: string | null;
  sync_status: SyncStatus;
  record_uuid: string;
  created_at: string;
}

export interface TaskStatusCounts {
  to_do: number;
  in_progress: number;
  paused: number;
  blocked: number;
  completed: number;
}

// ─── Request shapes ───────────────────────────────────────────────────────────

export interface CreateTaskPayload {
  title: string;
  description?: string;
  priority?: Priority;
  due_date?: string;
}

export interface AssignTaskPayload {
  user_ids: string[];
}

export interface ChangeStatusPayload {
  status: TaskStatus;
}

export interface TimerStartPayload {
  started_at: string;
  device_hash: string;
  record_uuid: string;
  sync_status?: SyncStatus;
}

export interface TimerPausePayload {
  paused_at: string;
  pause_reason: PauseReason;
  device_hash: string;
  record_uuid: string;
  sync_status?: SyncStatus;
}

export interface TimerResumePayload {
  resumed_at: string;
  device_hash: string;
  record_uuid: string;
  sync_status?: SyncStatus;
}

export interface TimerStopPayload {
  stopped_at: string;
  device_hash: string;
  record_uuid: string;
  sync_status?: SyncStatus;
}

// ─── Frontend-only derived types ──────────────────────────────────────────────

export type TimerState = "idle" | "running" | "paused";

export interface TaskFilterState {
  status: TaskStatus | "all";
  priority: Priority | "all";
  search: string;
  overdueOnly: boolean;
}
