/**
 * TaskStatusBadge.tsx
 *
 * Coloured pill badge for task status.  Minimal inline styles — no UI library.
 */

import type { TaskStatus } from "../types";
import { STATUS_LABELS } from "../types";

const COLOR_MAP: Record<TaskStatus, { bg: string; fg: string }> = {
  to_do:       { bg: "#e5e7eb", fg: "#374151" },
  in_progress: { bg: "#dbeafe", fg: "#1d4ed8" },
  paused:      { bg: "#fef9c3", fg: "#854d0e" },
  blocked:     { bg: "#fee2e2", fg: "#991b1b" },
  completed:   { bg: "#dcfce7", fg: "#166534" },
};

interface Props {
  status: TaskStatus;
}

export default function TaskStatusBadge({ status }: Props) {
  const colors = COLOR_MAP[status] ?? COLOR_MAP.to_do;
  return (
    <span
      style={{
        display: "inline-block",
        padding: "2px 10px",
        borderRadius: 12,
        fontSize: 12,
        fontWeight: 600,
        background: colors.bg,
        color: colors.fg,
      }}
    >
      {STATUS_LABELS[status] ?? status}
    </span>
  );
}
