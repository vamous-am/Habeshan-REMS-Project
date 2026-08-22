import { AlertCircle } from "lucide-react";

interface ErrorStateProps {
  title?: string;
  message: string;
}

export function ErrorState({
  title = "Something went wrong",
  message,
}: ErrorStateProps) {
  return (
    <div
      role="alert"
      className="flex gap-3 rounded border border-status-rejected/30 bg-status-rejected/5 p-4"
    >
      <AlertCircle
        className="mt-0.5 h-5 w-5 shrink-0 text-status-rejected"
        aria-hidden
      />
      <div>
        <p className="text-sm font-medium text-status-rejected">{title}</p>
        <p className="mt-1 text-sm text-ink">{message}</p>
      </div>
    </div>
  );
}
