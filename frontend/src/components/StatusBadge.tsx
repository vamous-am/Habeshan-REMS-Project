import type { StatusVariant } from "./statusBadgeUtils";

interface StatusBadgeProps {
  label: string;
  variant: StatusVariant;
}

const variantClasses: Record<StatusVariant, string> = {
  verified: "bg-status-verified/10 text-status-verified border-status-verified/20",
  pending: "bg-status-pending/10 text-status-pending border-status-pending/20",
  rejected: "bg-status-rejected/10 text-status-rejected border-status-rejected/20",
  offline: "bg-status-offline/10 text-status-offline border-status-offline/20",
};

export function StatusBadge({ label, variant }: StatusBadgeProps) {
  return (
    <span
      className={[
        "inline-flex items-center rounded border px-2 py-0.5",
        "text-xs font-medium uppercase tracking-wide",
        variantClasses[variant],
      ].join(" ")}
    >
      {label}
    </span>
  );
}
