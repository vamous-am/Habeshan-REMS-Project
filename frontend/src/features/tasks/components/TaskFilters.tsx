/**
 * components/TaskFilters.tsx
 *
 * Compact, self-contained filter controls for the task list.
 * Supports:
 *   - Text search (title / description)
 *   - Status filter (uses the status badge colours)
 *   - Priority filter
 *   - "Show overdue only" toggle
 *
 * Parent controls the filter state so the list can be filtered in one place
 * (no duplicate data-source or N-list fetches).
 */

import type { Priority, TaskFilterState, TaskStatus } from "../types";
import { PRIORITY_LABELS, STATUS_LABELS } from "../types";

interface Props {
  filters: TaskFilterState;
  onChange: (next: TaskFilterState) => void;
  overdueCount?: number;
}

const STATUSES: (TaskStatus | "all")[] = [
  "all",
  "to_do",
  "in_progress",
  "paused",
  "blocked",
  "completed",
];

const PRIORITIES: (Priority | "all")[] = ["all", "high", "medium", "low"];

export default function TaskFilters({
  filters,
  onChange,
  overdueCount,
}: Props) {
  return (
    <div
      style={{
        border: "1px solid #e5e7eb",
        borderRadius: 8,
        padding: 12,
        marginBottom: 16,
        background: "#fff",
        display: "grid",
        gap: 12,
        gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
      }}
    >
      {/* Search */}
      <label
        style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: 12 }}
      >
        <span style={{ fontWeight: 500, color: "#374151" }}>Search</span>
        <input
          value={filters.search}
          onChange={(e) => onChange({ ...filters, search: e.target.value })}
          placeholder="Search title or description…"
          style={{
            padding: "6px 8px",
            border: "1px solid #d1d5db",
            borderRadius: 4,
            fontSize: 13,
          }}
        />
      </label>

      {/* Status */}
      <label
        style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: 12 }}
      >
        <span style={{ fontWeight: 500, color: "#374151" }}>Status</span>
        <select
          value={filters.status}
          onChange={(e) =>
            onChange({ ...filters, status: e.target.value as TaskStatus | "all" })
          }
          style={{
            padding: "6px 8px",
            border: "1px solid #d1d5db",
            borderRadius: 4,
            fontSize: 13,
            background: "#fff",
          }}
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {s === "all" ? "All statuses" : STATUS_LABELS[s]}
            </option>
          ))}
        </select>
      </label>

      {/* Priority */}
      <label
        style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: 12 }}
      >
        <span style={{ fontWeight: 500, color: "#374151" }}>Priority</span>
        <select
          value={filters.priority}
          onChange={(e) =>
            onChange({ ...filters, priority: e.target.value as Priority | "all" })
          }
          style={{
            padding: "6px 8px",
            border: "1px solid #d1d5db",
            borderRadius: 4,
            fontSize: 13,
            background: "#fff",
          }}
        >
          {PRIORITIES.map((p) => (
            <option key={p} value={p}>
              {p === "all" ? "All priorities" : PRIORITY_LABELS[p]}
            </option>
          ))}
        </select>
      </label>

      {/* Overdue toggle */}
      <div
        style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 13 }}
      >
        <label
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 8,
            cursor: "pointer",
            padding: "8px 10px",
            border: filters.overdueOnly
              ? "2px solid #dc2626"
              : "1px solid #fee2e2",
            borderRadius: 6,
            background: filters.overdueOnly ? "#fff1f2" : "#fff",
          }}
        >
          <input
            type="checkbox"
            checked={filters.overdueOnly}
            onChange={(e) =>
              onChange({ ...filters, overdueOnly: e.target.checked })
            }
            style={{
              accentColor: "#dc2626",
              cursor: "pointer",
              margin: 0,
            }}
          />
          <span
            style={{
              color: overdueCount ? "#991b1b" : "#dc2626",
              fontWeight: filters.overdueOnly ? 600 : 500,
            }}
            title="Show tasks that are past their due date and still open"
          >
            ⚠ Overdue
            {typeof overdueCount === "number" && overdueCount > 0 ? (
              <span
                style={{
                  marginLeft: 6,
                  padding: "1px 8px",
                  borderRadius: 10,
                  background: "#fee2e2",
                  color: "#991b1b",
                  fontSize: 11,
                }}
              >
                {overdueCount}
              </span>
            ) : null}
          </span>
        </label>
      </div>
    </div>
  );
}
