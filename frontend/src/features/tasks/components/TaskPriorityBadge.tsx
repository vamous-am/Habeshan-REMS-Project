/**
 * components/TaskPriorityBadge.tsx
 *
 * Coloured pill badge for task priority.  Matches the three backend values:
 *   high | medium | low
 * Uses accessible colour contrast and short labels.
 */

import type { Priority } from "../types";
import { PRIORITY_LABELS } from "../types";

const COLOR_MAP: Record<Priority, { bg: string; fg: string; border?: string }> = {
  high:   { bg: "#fee2e2", fg: "#991b1b", border: "#fecaca" },
  medium: { bg: "#fef9c3", fg: "#854d0e", border: "#fef08a" },
  low:    { bg: "#dbeafe", fg: "#1d4ed8", border: "#bfdbfe" },
};

const ICON: Record<Priority, string> = {
  high: "▲",
  medium: "■",
  low: "▼",
};

interface Props {
  priority: Priority;
  size?: "sm" | "md";
}

export default function TaskPriorityBadge({ priority, size = "sm" }: Props) {
  const colors = COLOR_MAP[priority] ?? COLOR_MAP.medium;
  const padding = size === "md" ? "4px 12px" : "2px 10px";
  const fontSize = size === "md" ? 13 : 12;

  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 4,
        padding,
        borderRadius: 12,
        fontSize,
        fontWeight: 600,
        background: colors.bg,
        color: colors.fg,
        border: colors.border ? `1px solid ${colors.border}` : undefined,
      }}
      title={`Priority: ${PRIORITY_LABELS[priority]}`}
    >
      <span style={{ fontSize: fontSize - 2, opacity: 0.8 }}>
        {ICON[priority]}
      </span>
      {PRIORITY_LABELS[priority] ?? priority}
    </span>
  );
}
