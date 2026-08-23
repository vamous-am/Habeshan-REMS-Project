import { useEffect, useState } from "react";
import api from "../../lib/api/client";
import { getCurrentUserId } from "../../lib/api/authClient";
import { Calendar, Clock, Send, AlertCircle, CheckCircle2, RefreshCw } from "lucide-react";

interface Timesheet {
  id: string;
  period_start: string;
  period_end: string;
  total_hours: number;
  status: "draft" | "submitted" | "approved" | "rejected";
  rejection_reason?: string;
}

const statusBadgeStyles: Record<string, { label: string; className: string }> = {
  draft: {
    label: "DRAFT",
    className: "bg-ink/10 text-ink border border-ink/20",
  },
  submitted: {
    label: "SUBMITTED",
    className: "bg-ochre/15 text-ochre-dark border border-ochre/30",
  },
  approved: {
    label: "APPROVED",
    className: "bg-status-verified/15 text-status-verified border border-status-verified/30",
  },
  rejected: {
    label: "REJECTED",
    className: "bg-status-rejected/15 text-status-rejected border border-status-rejected/30",
  },
};

export default function TimesheetList() {
  const [timesheets, setTimesheets] = useState<Timesheet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submittingId, setSubmittingId] = useState<string | null>(null);

  async function loadTimesheets() {
    setLoading(true);
    setError(null);
    try {
      const userId = getCurrentUserId();
      const res = await api.get(userId ? `/timesheets?user_id=${userId}` : "/timesheets");
      const list = res.data?.data ?? res.data ?? [];
      setTimesheets(Array.isArray(list) ? list : []);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load timesheets");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadTimesheets();
  }, []);

  async function handleSubmit(id: string) {
    setSubmittingId(id);
    try {
      await api.put(`/timesheets/${id}/submit`);
      await loadTimesheets();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Failed to submit timesheet");
    } finally {
      setSubmittingId(null);
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="flex items-center gap-2 font-mono text-sm text-ink500">
          <RefreshCw className="h-4 w-4 animate-spin text-ochre" />
          <span>Loading timesheet records…</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-ink/10 pb-4">
        <div>
          <h1 className="font-display text-2xl font-bold tracking-tight text-ink">
            My Timesheets
          </h1>
          <p className="mt-1 text-sm text-ink500">
            Bi-weekly work summaries automatically calculated from attendance logs.
          </p>
        </div>
        <button
          onClick={loadTimesheets}
          className="flex items-center gap-1.5 rounded border border-ink/15 bg-paper px-3 py-1.5 text-xs font-medium text-ink transition hover:bg-ink/5"
        >
          <RefreshCw className="h-3.5 w-3.5" />
          <span>Refresh</span>
        </button>
      </div>

      {error && (
        <div className="flex items-center gap-2 rounded-lg border border-status-rejected/30 bg-status-rejected/10 p-4 text-xs text-status-rejected">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {timesheets.length === 0 ? (
        <div className="rounded-xl border border-dashed border-ink/20 bg-paper-dim/40 p-12 text-center">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-ink/10 text-ink">
            <Calendar className="h-6 w-6" />
          </div>
          <h3 className="mt-3 font-display text-base font-bold text-ink">
            No Timesheet Periods Found
          </h3>
          <p className="mt-1 text-xs text-ink500">
            Clock-in logs will automatically aggregate into timesheet review cycles.
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {timesheets.map((ts) => {
            const badge = statusBadgeStyles[ts.status] || statusBadgeStyles.draft;
            const isSubmitting = submittingId === ts.id;

            return (
              <div
                key={ts.id}
                className="flex flex-col gap-4 rounded-xl border border-ink/10 bg-paper p-5 shadow-xs transition hover:border-ink/20 sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="space-y-1.5">
                  <div className="flex items-center gap-2">
                    <span className="font-display text-sm font-semibold text-ink">
                      Period: {ts.period_start} — {ts.period_end}
                    </span>
                    <span
                      className={`rounded px-2 py-0.5 font-mono text-[10px] font-semibold ${badge.className}`}
                    >
                      {badge.label}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-ink500">
                    <Clock className="h-3.5 w-3.5" />
                    <span className="font-semibold text-ink">
                      {ts.total_hours} logged hours
                    </span>
                  </div>
                  {ts.rejection_reason && (
                    <p className="rounded bg-status-rejected/10 p-2 text-xs font-medium text-status-rejected">
                      Rejection Reason: {ts.rejection_reason}
                    </p>
                  )}
                </div>

                {/* Submit Action */}
                <div>
                  {(ts.status === "draft" || ts.status === "rejected") && (
                    <button
                      type="button"
                      disabled={isSubmitting}
                      onClick={() => handleSubmit(ts.id)}
                      className="flex items-center gap-1.5 rounded bg-ink px-4 py-2 text-xs font-semibold text-paper shadow-xs transition hover:bg-ink-light disabled:opacity-50"
                    >
                      <Send className="h-3.5 w-3.5" />
                      <span>{isSubmitting ? "Submitting…" : "Submit to Manager"}</span>
                    </button>
                  )}
                  {ts.status === "approved" && (
                    <div className="flex items-center gap-1.5 font-mono text-xs font-semibold text-status-verified">
                      <CheckCircle2 className="h-4 w-4" />
                      <span>Verified & Approved</span>
                    </div>
                  )}
                  {ts.status === "submitted" && (
                    <span className="font-mono text-xs text-ochre-dark">
                      Pending Manager Review
                    </span>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
