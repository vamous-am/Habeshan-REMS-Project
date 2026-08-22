/**
 * components/TaskStatsCard.tsx
 *
 * Summary card showing task counts grouped by status (FR-TASK-09 aggregates).
 * Shows 5 buckets with the count / total and click-to-filter callback.
 */

import type { TaskStatus, TaskStatusCounts } from "../types";
import { STATUS_LABELS } from "../types";
import TaskStatusBadge from "./TaskStatusBadge";

interface Props {
  counts: TaskStatusCounts | null;
  loading?: boolean;
  error?: string | null;
  /** When a status is clicked, this callback fires — parent can filter the list. */
  onStatusClick?: (status: TaskStatus | "all") => void;
  activeStatus?: TaskStatus | "all";
}

const ORDER: TaskStatus[] = [
  "to_do",
  "in_progress",
  "paused",
  "blocked",
  "completed",
];

export default function TaskStatsCard({
  counts,
  loading,
  error,
  onStatusClick,
  activeStatus = "all",
}: Props) {
  const total = counts
    ? counts.to_do +
      counts.in_progress +
      counts.paused +
      counts.blocked +
      counts.completed
    : 0;

  return (
    <div
      style={{
        border: "1px solid #e5e7eb",
        borderRadius: 8,
        padding: 16,
        marginBottom: 20,
        background: "#fff",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 12,
        }}
      >
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: "#111827" }}>
          Task Overview
        </h3>
        {!loading && counts && (
          <span style={{ fontSize: 13, color: "#6b7280" }}>
            {total} total
          </span>
        )}
      </div>

      {loading && (
        <p style={{ margin: 0, fontSize: 13, color: "#9ca3af" }}>
          Loading counts…
        </p>
      )}

      {error && (
        <p style={{ margin: 0, fontSize: 13, color: "#dc2626" }}>{error}</p>
      )}

      {!loading && !error && counts && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns:
              "repeat(auto-fit, minmax(130px, 1fr))",
            gap: 8,
          }}
        >
          <Bucket
            label="All"
            count={total}
            active={activeStatus === "all"}
            onClick={() => onStatusClick?.("all")}
          />
          {ORDER.map((s) => {
            const count = counts[s];
            const pct = total > 0 ? Math.round((count / total) * 100) : 0;
            return (
              <Bucket
                key={s}
                label={STATUS_LABELS[s]}
                count={count}
                pct={pct}
                active={activeStatus === s}
                onClick={() => onStatusClick?.(s)}
                badge={<TaskStatusBadge status={s} />}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

// ─── sub components ──────────────────────────────────────────────────────────

interface BucketProps {
  label: string;
  count: number;
  pct?: number;
  active?: boolean;
  onClick?: () => void;
  badge?: React.ReactNode;
}

function Bucket({ label, count, pct, active, onClick, badge }: BucketProps) {
  const clickable = !!onClick;
  return (
    <button
      type="button"
      disabled={!clickable}
      onClick={onClick}
      style={{
        textAlign: "left" as const,
        padding: 12,
        border: active ? "2px solid #2563eb" : "1px solid #e5e7eb",
        borderRadius: 8,
        background: active ? "#eff6ff" : "#fafafa",
        cursor: clickable ? "pointer" : "default",
        display: "flex",
        flexDirection: "column" as const,
        gap: 4,
        minHeight: 72,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <span style={{ fontSize: 12, color: "#6b7280" }}>{label}</span>
        {badge}
      </div>
      <div style={{ display: "flex", alignItems: "baseline", gap: 8 }}>
        <span style={{ fontSize: 22, fontWeight: 700, color: "#111827" }}>
          {count}
        </span>
        {pct != null && (
          <span style={{ fontSize: 12, color: "#9ca3af" }}>{pct}%</span>
        )}
      </div>
    </button>
  );
}
