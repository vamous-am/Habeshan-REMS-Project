/**
 * features/tasks/api.ts
 *
 * All HTTP calls for the tasks feature.  Every function returns the unwrapped
 * data payload from the server's { status, data } envelope.
 *
 * Route contract (all under /api/v1 — configured via VITE_API_BASE_URL):
 *   GET    /tasks                          — visibility-scoped task list
 *   POST   /tasks                          — create task
 *   GET    /tasks/stats/status-counts      — status counts (admin/manager)
 *   GET    /tasks/stats/overdue            — overdue tasks (admin/manager)
 *   GET    /tasks/:id                      — single task detail
 *   PATCH  /tasks/:id/status             — change status
 *   POST   /tasks/:id/assignments          — assign users
 *   DELETE /tasks/:id/assignments/:userID  — unassign user
 *   POST   /tasks/:id/timer/start        — start timer
 *   POST   /tasks/:id/timer/pause        — pause timer
 *   POST   /tasks/:id/timer/resume       — resume timer
 *   POST   /tasks/:id/timer/stop         — stop timer
 *   GET    /tasks/:id/timer                — timer history
 */

import api from "../../lib/api/client";
import type {
  Task,
  TimeLog,
  TaskStatusCounts,
  CreateTaskPayload,
  AssignTaskPayload,
  ChangeStatusPayload,
  TimerStartPayload,
  TimerPausePayload,
  TimerResumePayload,
  TimerStopPayload,
} from "./types";

// ─── helpers ─────────────────────────────────────────────────────────────────

/**
 * The backend always wraps successful responses in:
 *   { "status": "success", "data": ... }
 *
 * Failures look like:
 *   { "status": "fail", "message": "..." }
 *
 * unwrap() reads response.data.data and throws a typed error on any problem.
 */
interface Envelope<T> {
  status: "success" | "fail";
  data?: T;
  message?: string;
}

export class ApiError extends Error {
  readonly status?: number;
  constructor(message: string, status?: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

function unwrap<T>(response: { data: Envelope<T>; status: number }): T {
  const env = response.data;
  if (env && env.status === "success" && env.data !== undefined) {
    return env.data;
  }
  if (env && env.status === "fail" && env.message) {
    throw new ApiError(env.message, response.status);
  }
  // Malformed envelope — best-effort message from axios
  throw new ApiError("Unexpected response format from server", response.status);
}

/**
 * Extracts a user-friendly message from any thrown value.
 */
export function extractApiErrorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message;
  if (
    err &&
    typeof err === "object" &&
    "response" in (err as Record<string, unknown>)
  ) {
    const resp = (err as { response?: { data?: { message?: string } } }).response;
    const msg = resp?.data?.message;
    if (msg && typeof msg === "string") return msg;
  }
  if (err instanceof Error) return err.message;
  return "An unexpected error occurred";
}

const BASE = "/tasks";

// ─── task CRUD ────────────────────────────────────────────────────────────────

export async function fetchMyTasks(): Promise<Task[]> {
  return unwrap<Task[]>(await api.get(BASE));
}

export async function fetchTaskByID(taskID: string): Promise<Task> {
  return unwrap<Task>(await api.get(`${BASE}/${taskID}`));
}

export async function createTask(payload: CreateTaskPayload): Promise<Task> {
  return unwrap<Task>(await api.post(BASE, payload));
}

// ─── status ───────────────────────────────────────────────────────────────────

export async function changeTaskStatus(
  taskID: string,
  payload: ChangeStatusPayload
): Promise<{ task_id: string; status: string }> {
  return unwrap(await api.patch(`${BASE}/${taskID}/status`, payload));
}

// ─── assignments ─────────────────────────────────────────────────────────────

export async function assignTask(
  taskID: string,
  payload: AssignTaskPayload
): Promise<{ task_id: string; assigned_to: string[] }> {
  return unwrap(await api.post(`${BASE}/${taskID}/assignments`, payload));
}

export async function unassignTask(
  taskID: string,
  userID: string
): Promise<{ message: string }> {
  return unwrap(await api.delete(`${BASE}/${taskID}/assignments/${userID}`));
}

// ─── timer ────────────────────────────────────────────────────────────────────

export async function startTimer(
  taskID: string,
  payload: TimerStartPayload
): Promise<TimeLog> {
  return unwrap<TimeLog>(
    await api.post(`${BASE}/${taskID}/timer/start`, payload)
  );
}

export async function pauseTimer(
  taskID: string,
  payload: TimerPausePayload
): Promise<TimeLog> {
  return unwrap<TimeLog>(
    await api.post(`${BASE}/${taskID}/timer/pause`, payload)
  );
}

export async function resumeTimer(
  taskID: string,
  payload: TimerResumePayload
): Promise<TimeLog> {
  return unwrap<TimeLog>(
    await api.post(`${BASE}/${taskID}/timer/resume`, payload)
  );
}

export async function stopTimer(
  taskID: string,
  payload: TimerStopPayload
): Promise<TimeLog> {
  return unwrap<TimeLog>(
    await api.post(`${BASE}/${taskID}/timer/stop`, payload)
  );
}

export async function fetchTimerHistory(taskID: string): Promise<TimeLog[]> {
  return unwrap<TimeLog[]>(await api.get(`${BASE}/${taskID}/timer`));
}

// ─── aggregates ───────────────────────────────────────────────────────────────

export async function fetchStatusCounts(): Promise<TaskStatusCounts> {
  return unwrap<TaskStatusCounts>(
    await api.get(`${BASE}/stats/status-counts`)
  );
}

export async function fetchOverdueTasks(): Promise<Task[]> {
  return unwrap<Task[]>(await api.get(`${BASE}/stats/overdue`));
}
