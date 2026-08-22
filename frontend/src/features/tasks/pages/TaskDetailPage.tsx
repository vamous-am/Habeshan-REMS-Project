/**
 * TaskDetailPage.tsx
 *
 * Shows a single task's full details, assignment management, status controls,
 * timer controls, and timer audit history.
 */

import { useEffect, useMemo, useState } from "react";
import { useParams, Link } from "react-router-dom";
import type { Task, TaskStatus, TimeLog } from "../types";
import { STATUS_LABELS } from "../types";
import {
  extractApiErrorMessage,
  fetchTaskByID,
  changeTaskStatus,
  fetchTimerHistory,
} from "../api";
import TaskStatusBadge from "../components/TaskStatusBadge";
import TaskPriorityBadge from "../components/TaskPriorityBadge";
import TimerControls from "../components/TimerControls";
import TaskAssignmentPanel from "../components/TaskAssignmentPanel";
import { useTimer, formatDuration } from "../useTimer";

// Statuses reachable from each current status — mirrors the backend transition
// table.  Used only to populate buttons; actual validation still happens on
// the backend.
const NEXT_STATUSES: Record<TaskStatus, TaskStatus[]> = {
  to_do:       ["in_progress", "blocked"],
  in_progress: ["paused", "blocked", "completed"],
  paused:      ["in_progress", "blocked"],
  blocked:     ["in_progress", "to_do"],
  completed:   [],
};

function isOverdue(task: Task): boolean {
  if (!task.due_date) return false;
  if (task.status === "completed") return false;
  const due = new Date(task.due_date);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  due.setHours(0, 0, 0, 0);
  return due.getTime() < today.getTime();
}

export default function TaskDetailPage() {
  const { id } = useParams<{ id: string }>();
  const taskID = id ?? "";

  const [task, setTask] = useState<Task | null>(null);
  const [loadingTask, setLoadingTask] = useState(true);
  const [loadingLogs, setLoadingLogs] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [statusMsg, setStatusMsg] = useState<{ text: string; error?: boolean } | null>(null);

  // Timer state — seed from server history when available.
  const [initialLogs, setInitialLogs] = useState<TimeLog[]>([]);
  const timer = useTimer(taskID, {
    initialLogs,
    onStopped: () => {
      // No-op; we refresh below explicitly on every action completion.
    },
  });

  async function loadTask() {
    if (!taskID) return;
    setLoadingTask(true);
    setError(null);
    try {
      const t = await fetchTaskByID(taskID);
      setTask(t);
    } catch (e: unknown) {
      setError(extractApiErrorMessage(e));
    } finally {
      setLoadingTask(false);
    }
  }

  async function loadLogs() {
    if (!taskID) return;
    setLoadingLogs(true);
    try {
      const logs = await fetchTimerHistory(taskID);
      const safeLogs = logs ?? [];
      setInitialLogs(safeLogs);
      timer.setLogs(safeLogs);
    } catch {
      // Timer history is best-effort.
      setInitialLogs([]);
      timer.setLogs([]);
    } finally {
      setLoadingLogs(false);
    }
  }

  useEffect(() => {
    let cancelled = false;
    queueMicrotask(() => {
      if (cancelled) return;
      loadTask();
      loadLogs();
    });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [taskID]);

  async function handleStatusChange(next: TaskStatus) {
    if (!taskID || !task) return;
    setStatusMsg(null);
    try {
      await changeTaskStatus(taskID, { status: next });
      setTask({ ...task, status: next });
      setStatusMsg({ text: `Status changed to ${STATUS_LABELS[next]}.` });
      setTimeout(() => setStatusMsg(null), 3500);
    } catch (e: unknown) {
      setStatusMsg({
        text: extractApiErrorMessage(e),
        error: true,
      });
    }
  }

  const totalMinutesFromLogs = useMemo(
    () =>
      timer.logs.reduce((acc, l) => acc + (l.duration_minutes ?? 0), 0) +
      (timer.timerState === "running"
        ? Math.floor(timer.elapsedSeconds / 60)
        : 0),
    [timer.logs, timer.timerState, timer.elapsedSeconds]
  );

  if (loadingTask)
    return (
      <div style={page}>
        <p style={{ color: "#6b7280" }}>Loading task…</p>
      </div>
    );
  if (error)
    return (
      <div style={page}>
        <Link to="/tasks" style={{ fontSize: 13, color: "#6b7280" }}>
          ← Back to tasks
        </Link>
        <div
          style={{
            marginTop: 16,
            padding: 16,
            background: "#fef2f2",
            borderRadius: 8,
            border: "1px solid #fecaca",
            color: "#991b1b",
          }}
        >
          <strong>Error loading task:</strong>
          <div style={{ marginTop: 4 }}>{error}</div>
        </div>
      </div>
    );
  if (!task) return null;

  const nextStatuses = NEXT_STATUSES[task.status] ?? [];
  const overdue = isOverdue(task);

  return (
    <div style={{ ...page, maxWidth: 900 }}>
      <Link
        to="/tasks"
        style={{
          fontSize: 13,
          color: "#6b7280",
          textDecoration: "none",
          display: "inline-flex",
          alignItems: "center",
          gap: 4,
        }}
      >
        ← Back to tasks
      </Link>

      <header style={{ marginTop: 12, marginBottom: 16 }}>
        <h1 style={{ margin: "0 0 8px", fontSize: 24, fontWeight: 700 }}>
          {task.title}
        </h1>
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: 10,
            alignItems: "center",
          }}
        >
          <TaskStatusBadge status={task.status} />
          <TaskPriorityBadge priority={task.priority} size="md" />
          {task.due_date && (
            <span
              style={{
                fontSize: 13,
                color: overdue ? "#9a3412" : "#6b7280",
                fontWeight: overdue ? 600 : 500,
                padding: overdue ? "4px 10px" : 0,
                background: overdue ? "#fff7ed" : undefined,
                borderRadius: 6,
                border: overdue ? "1px solid #fed7aa" : undefined,
              }}
            >
              📅 Due:{" "}
              <strong>{new Date(task.due_date).toLocaleDateString()}</strong>
              {overdue && " ⚠ Overdue"}
            </span>
          )}
          <span style={{ fontSize: 13, color: "#9ca3af" }}>
            Created {new Date(task.created_at).toLocaleDateString()}
          </span>
        </div>
      </header>

      {/* Description */}
      <section style={section}>
        <h2 style={sectionHeading}>Description</h2>
        {task.description ? (
          <p style={bodyText}>{task.description}</p>
        ) : (
          <p style={{ ...bodyText, color: "#9ca3af", fontStyle: "italic" }}>
            No description provided.
          </p>
        )}
      </section>

      {/* Status change */}
      <section style={section}>
        <h2 style={sectionHeading}>Change Status</h2>
        {nextStatuses.length === 0 ? (
          <p style={{ ...bodyText, color: "#6b7280", fontSize: 13 }}>
            This task is in a terminal state (<strong>Completed</strong>) and
            cannot be transitioned further.
          </p>
        ) : (
          <>
            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                gap: 8,
                marginBottom: 8,
              }}
            >
              {nextStatuses.map((s) => (
                <button
                  key={s}
                  type="button"
                  onClick={() => handleStatusChange(s)}
                  style={{
                    padding: "6px 14px",
                    fontSize: 13,
                    fontWeight: 500,
                    background: "#fff",
                    color: "#111827",
                    border: "1px solid #d1d5db",
                    borderRadius: 6,
                    cursor: "pointer",
                  }}
                >
                  → {STATUS_LABELS[s]}
                </button>
              ))}
            </div>
            {statusMsg && (
              <div
                style={{
                  fontSize: 13,
                  padding: "8px 12px",
                  borderRadius: 6,
                  background: statusMsg.error ? "#fef2f2" : "#ecfdf5",
                  color: statusMsg.error ? "#991b1b" : "#065f46",
                  border: statusMsg.error
                    ? "1px solid #fecaca"
                    : "1px solid #a7f3d0",
                }}
              >
                {statusMsg.text}
              </div>
            )}
          </>
        )}
      </section>

      {/* Assignments panel */}
      <TaskAssignmentPanel task={task} onChanged={loadTask} />

      {/* Timer controls */}
      <section style={section}>
        <h2 style={sectionHeading}>Timer</h2>
        <TimerControls
          taskID={taskID}
          timerState={timer.timerState}
          loading={timer.loading}
          error={timer.error}
          elapsedSeconds={timer.elapsedSeconds}
          completedMinutes={timer.completedMinutes}
          onStart={async () => {
            await timer.handleStart();
            await loadLogs();
          }}
          onPause={async (reason) => {
            await timer.handlePause(reason);
            await loadLogs();
          }}
          onResume={async () => {
            await timer.handleResume();
            await loadLogs();
          }}
          onStop={async () => {
            await timer.handleStop();
            await loadLogs();
          }}
        />
      </section>

      {/* Timer history */}
      <section style={section}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 10,
          }}
        >
          <h2 style={{ ...sectionHeading, margin: 0 }}>
            Timer History
            {loadingLogs && (
              <span style={{ fontSize: 12, color: "#9ca3af", marginLeft: 8 }}>
                loading…
              </span>
            )}
          </h2>
          <div style={{ fontSize: 13, color: "#6b7280" }}>
            Total logged:{" "}
            <strong style={{ color: "#111827" }}>
              {formatDuration(totalMinutesFromLogs)}
            </strong>
            {" · "}
            <button
              type="button"
              onClick={loadLogs}
              style={{
                fontSize: 12,
                color: "#2563eb",
                border: "none",
                background: "transparent",
                padding: 0,
                cursor: "pointer",
              }}
            >
              refresh
            </button>
          </div>
        </div>

        {timer.logs.length === 0 && !loadingLogs && (
          <p
            style={{
              fontSize: 13,
              color: "#6b7280",
              padding: 16,
              textAlign: "center",
              border: "1px dashed #d1d5db",
              borderRadius: 6,
            }}
          >
            No timer segments yet. Start the timer above to log time against
            this task.
          </p>
        )}

        {timer.logs.length > 0 && (
          <div style={{ overflowX: "auto" }}>
            <table
              style={{
                width: "100%",
                borderCollapse: "collapse",
                fontSize: 13,
                background: "#fff",
                border: "1px solid #e5e7eb",
                borderRadius: 8,
                overflow: "hidden",
              }}
            >
              <thead>
                <tr
                  style={{
                    borderBottom: "2px solid #e5e7eb",
                    textAlign: "left",
                    background: "#f9fafb",
                  }}
                >
                  <th style={thStyle}>User</th>
                  <th style={thStyle}>Started</th>
                  <th style={thStyle}>Stopped</th>
                  <th style={thStyle}>Duration</th>
                  <th style={thStyle}>Pause Reason</th>
                </tr>
              </thead>
              <tbody>
                {[...timer.logs]
                  .sort(
                    (a, b) =>
                      new Date(b.started_at).getTime() -
                      new Date(a.started_at).getTime()
                  )
                  .map((log) => {
                    const open = !log.stopped_at;
                    return (
                      <tr
                        key={log.id}
                        style={{
                          borderBottom: "1px solid #f3f4f6",
                          background: open ? "#f0fdf4" : undefined,
                        }}
                      >
                        <td style={tdStyle}>
                          <span
                            style={{
                              fontFamily:
                                'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
                              fontSize: 11,
                              color: "#374151",
                            }}
                            title={log.user_id}
                          >
                            {log.user_id.slice(0, 8)}…
                          </span>
                        </td>
                        <td style={tdStyle}>
                          {new Date(log.started_at).toLocaleString()}
                        </td>
                        <td style={tdStyle}>
                          {open ? (
                            <span
                              style={{
                                color: "#047857",
                                fontWeight: 600,
                                fontSize: 12,
                              }}
                            >
                              ● RUNNING
                            </span>
                          ) : (
                            new Date(log.stopped_at!).toLocaleString()
                          )}
                        </td>
                        <td style={tdStyle}>
                          {open ? (
                            <span style={{ color: "#6b7280" }}>in progress…</span>
                          ) : log.duration_minutes != null ? (
                            formatDuration(log.duration_minutes)
                          ) : (
                            "—"
                          )}
                        </td>
                        <td style={tdStyle}>
                          {log.pause_reason ? (
                            <span
                              style={{
                                padding: "2px 8px",
                                background: "#fef9c3",
                                color: "#854d0e",
                                borderRadius: 10,
                                fontSize: 12,
                              }}
                            >
                              {log.pause_reason}
                            </span>
                          ) : open ? (
                            <span style={{ color: "#9ca3af", fontSize: 12 }}>
                              (none yet)
                            </span>
                          ) : (
                            <span style={{ color: "#9ca3af" }}>—</span>
                          )}
                        </td>
                      </tr>
                    );
                  })}
              </tbody>
              <tfoot>
                <tr style={{ background: "#f9fafb" }}>
                  <td
                    colSpan={3}
                    style={{
                      ...tdStyle,
                      textAlign: "right",
                      fontWeight: 600,
                      color: "#374151",
                    }}
                  >
                    Total completed:
                  </td>
                  <td
                    style={{
                      ...tdStyle,
                      fontWeight: 700,
                      color: "#111827",
                    }}
                  >
                    {formatDuration(
                      timer.logs.reduce(
                        (a, l) => a + (l.duration_minutes ?? 0),
                        0
                      )
                    )}
                  </td>
                  <td style={tdStyle}></td>
                </tr>
              </tfoot>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}

// ─── layout helpers ───────────────────────────────────────────────────────────

const page: React.CSSProperties = {
  padding: 24,
  fontFamily:
    'ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, sans-serif',
  maxWidth: 900,
  margin: "0 auto",
};

const section: React.CSSProperties = {
  border: "1px solid #e5e7eb",
  borderRadius: 8,
  padding: 16,
  marginBottom: 20,
  background: "#fff",
};

const sectionHeading: React.CSSProperties = {
  margin: "0 0 10px",
  fontSize: 14,
  fontWeight: 600,
  color: "#111827",
};

const bodyText: React.CSSProperties = {
  margin: 0,
  color: "#374151",
  fontSize: 14,
  lineHeight: 1.5,
};

const thStyle: React.CSSProperties = {
  padding: "8px 12px",
  fontWeight: 600,
  color: "#374151",
  fontSize: 12,
  whiteSpace: "nowrap",
};

const tdStyle: React.CSSProperties = {
  padding: "8px 12px",
  verticalAlign: "middle",
  fontSize: 13,
};
