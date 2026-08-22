/**
 * TaskListPage.tsx
 *
 * Displays the visibility-scoped task list for the authenticated user.
 * Backend enforces visibility (FR-TASK-03) — client-side filtering is
 * supplemental (search / priority / overdue-only) and layered on top.
 */

import { useEffect, useMemo, useState, useCallback } from "react";
import { Link } from "react-router-dom";
import type { Task, TaskFilterState, TaskStatus, TaskStatusCounts } from "../types";
import {
  extractApiErrorMessage,
  fetchMyTasks,
  fetchStatusCounts,
} from "../api";
import TaskStatusBadge from "../components/TaskStatusBadge";
import TaskPriorityBadge from "../components/TaskPriorityBadge";
import CreateTaskForm from "../components/CreateTaskForm";
import TaskStatsCard from "../components/TaskStatsCard";
import TaskFilters from "../components/TaskFilters";

const DEFAULT_FILTERS: TaskFilterState = {
  status: "all",
  priority: "all",
  search: "",
  overdueOnly: false,
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

export default function TaskListPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [counts, setCounts] = useState<TaskStatusCounts | null>(null);
  const [loading, setLoading] = useState(true);
  const [countsLoading, setCountsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [countsError, setCountsError] = useState<string | null>(null);
  const [filters, setFilters] = useState<TaskFilterState>(DEFAULT_FILTERS);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchMyTasks();
      setTasks(data ?? []);
    } catch (e: unknown) {
      setError(extractApiErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }, []);

  const loadCounts = useCallback(async () => {
    setCountsLoading(true);
    setCountsError(null);
    try {
      const c = await fetchStatusCounts();
      setCounts(c);
    } catch (e: unknown) {
      // Status counts are only available to admin/manager — failing quietly for
      // employees is expected and fine; show the error in the card only.
      setCountsError(extractApiErrorMessage(e));
      setCounts(null);
    } finally {
      setCountsLoading(false);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    queueMicrotask(() => {
      if (cancelled) return;
      load();
      loadCounts();
    });
    return () => {
      cancelled = true;
    };
  }, [load, loadCounts]);

  // Compute derived counters for filter UI.
  const overdueCount = useMemo(
    () => tasks.filter((t) => isOverdue(t)).length,
    [tasks]
  );

  // Client-side filter application.
  const visibleTasks = useMemo(() => {
    const q = filters.search.trim().toLowerCase();
    return tasks.filter((t) => {
      if (filters.status !== "all" && t.status !== filters.status) return false;
      if (filters.priority !== "all" && t.priority !== filters.priority) return false;
      if (filters.overdueOnly && !isOverdue(t)) return false;
      if (q) {
        const hay =
          `${t.title} ${t.description ?? ""}`.toLowerCase();
        if (!hay.includes(q)) return false;
      }
      return true;
    });
  }, [tasks, filters]);

  function handleStatusClick(status: TaskStatus | "all") {
    setFilters((prev) => ({ ...prev, status }));
  }

  return (
    <div
      style={{
        padding: 24,
        fontFamily:
          'ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, sans-serif',
        maxWidth: 1100,
        margin: "0 auto",
      }}
    >
      <header style={{ marginBottom: 20 }}>
        <h1 style={{ margin: "0 0 4px", fontSize: 24, fontWeight: 700 }}>
          Tasks
        </h1>
        <p style={{ color: "#6b7280", margin: 0, fontSize: 14 }}>
          Showing tasks you have access to based on your role.
        </p>
      </header>

      <TaskStatsCard
        counts={counts}
        loading={countsLoading}
        error={countsError}
        onStatusClick={handleStatusClick}
        activeStatus={filters.status}
      />

      <CreateTaskForm onCreated={load} />

      <TaskFilters
        filters={filters}
        onChange={setFilters}
        overdueCount={overdueCount}
      />

      {loading && <p style={{ fontSize: 14, color: "#6b7280" }}>Loading tasks…</p>}
      {error && (
        <p
          style={{
            color: "#dc2626",
            fontSize: 14,
            padding: 10,
            background: "#fef2f2",
            borderRadius: 6,
            border: "1px solid #fecaca",
          }}
        >
          {error}
        </p>
      )}

      {!loading && !error && visibleTasks.length === 0 && (
        <div
          style={{
            padding: 32,
            textAlign: "center",
            border: "1px dashed #d1d5db",
            borderRadius: 8,
            color: "#6b7280",
            fontSize: 14,
          }}
        >
          {tasks.length === 0
            ? "No tasks found. Create one to get started."
            : "No tasks match the current filters."}
        </div>
      )}

      {!loading && !error && visibleTasks.length > 0 && (
        <div style={{ overflowX: "auto" }}>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: 14,
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
                <th style={thStyle}>Title</th>
                <th style={thStyle}>Priority</th>
                <th style={thStyle}>Status</th>
                <th style={thStyle}>Due Date</th>
                <th style={thStyle}>Assignees</th>
                <th style={thStyle}></th>
              </tr>
            </thead>
            <tbody>
              {visibleTasks.map((task) => {
                const overdue = isOverdue(task);
                const assignedCount = (task.assigned_to ?? []).length;
                return (
                  <tr
                    key={task.id}
                    style={{
                      borderBottom: "1px solid #f3f4f6",
                      background: overdue ? "#fff7ed" : undefined,
                    }}
                  >
                    <td style={tdStyle}>
                      <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
                        <Link
                          to={`/tasks/${task.id}`}
                          style={{
                            color: "#1d4ed8",
                            textDecoration: "none",
                            fontWeight: 500,
                          }}
                        >
                          {task.title}
                          {overdue && (
                            <span
                              style={{
                                marginLeft: 8,
                                color: "#9a3412",
                                fontSize: 12,
                                fontWeight: 600,
                              }}
                              title="Overdue"
                            >
                              ⚠ Overdue
                            </span>
                          )}
                        </Link>
                        {task.description && (
                          <span
                            style={{
                              fontSize: 12,
                              color: "#6b7280",
                              maxWidth: 420,
                              overflow: "hidden",
                              textOverflow: "ellipsis",
                              whiteSpace: "nowrap",
                            }}
                          >
                            {task.description}
                          </span>
                        )}
                      </div>
                    </td>
                    <td style={tdStyle}>
                      <TaskPriorityBadge priority={task.priority} />
                    </td>
                    <td style={tdStyle}>
                      <TaskStatusBadge status={task.status} />
                    </td>
                    <td
                      style={{
                        ...tdStyle,
                        color: overdue ? "#9a3412" : "#374151",
                        fontWeight: overdue ? 600 : 400,
                      }}
                    >
                      {task.due_date
                        ? new Date(task.due_date).toLocaleDateString()
                        : "—"}
                    </td>
                    <td style={tdStyle}>
                      {assignedCount > 0 ? (
                        <span
                          style={{
                            padding: "2px 10px",
                            borderRadius: 12,
                            background: "#eff6ff",
                            color: "#1d4ed8",
                            fontSize: 12,
                            fontWeight: 500,
                          }}
                          title={`Assigned to ${assignedCount} user${assignedCount > 1 ? "s" : ""}`}
                        >
                          {assignedCount} assigned
                        </span>
                      ) : (
                        <span
                          style={{
                            fontSize: 12,
                            color: "#9ca3af",
                            fontStyle: "italic",
                          }}
                        >
                          Unassigned
                        </span>
                      )}
                    </td>
                    <td style={{ ...tdStyle, textAlign: "right" }}>
                      <Link
                        to={`/tasks/${task.id}`}
                        style={{
                          fontSize: 12,
                          color: "#2563eb",
                          textDecoration: "none",
                          fontWeight: 500,
                        }}
                      >
                        View →
                      </Link>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

const thStyle: React.CSSProperties = {
  padding: "10px 14px",
  fontWeight: 600,
  color: "#374151",
  fontSize: 13,
  whiteSpace: "nowrap",
};

const tdStyle: React.CSSProperties = {
  padding: "12px 14px",
  verticalAlign: "middle",
  whiteSpace: "nowrap",
};
