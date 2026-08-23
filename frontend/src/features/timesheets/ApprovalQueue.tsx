import { useEffect, useState } from "react";
import api from "../../lib/api/client";
import { ShieldCheck, Check, X, AlertCircle, RefreshCw, Calendar, Clock } from "lucide-react";

interface Timesheet {
  id: string;
  user_id: string;
  period_start: string;
  period_end: string;
  total_hours: number;
  status: string;
}

export default function ApprovalQueue() {
  const [timesheets, setTimesheets] = useState<Timesheet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<Record<string, boolean>>({});
  const [rejectReason, setRejectReason] = useState<Record<string, string>>({});
  const [showRejectBox, setShowRejectBox] = useState<Record<string, boolean>>({});

  async function loadPending() {
    setLoading(true);
    setError(null);
    try {
      const res = await api.get("/timesheets?status=submitted");
      const list = res.data?.data ?? res.data ?? [];
      setTimesheets(Array.isArray(list) ? list : []);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load approval queue");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    let active = true;
    api.get("/timesheets?status=submitted")
      .then((res) => {
        if (active) {
          const list = res.data?.data ?? res.data ?? [];
          setTimesheets(Array.isArray(list) ? list : []);
        }
      })
      .catch((err: unknown) => {
        if (active) setError(err instanceof Error ? err.message : "Failed to load approval queue");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function handleApprove(id: string) {
    setActionLoading((prev) => ({ ...prev, [id]: true }));
    try {
      await api.put(`/timesheets/${id}/approve`);
      await loadPending();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Failed to approve timesheet");
    } finally {
      setActionLoading((prev) => ({ ...prev, [id]: false }));
    }
  }

  async function handleReject(id: string) {
    const reason = rejectReason[id];
    if (!reason || !reason.trim()) {
      alert("Rejection reason is required.");
      return;
    }
    setActionLoading((prev) => ({ ...prev, [id]: true }));
    try {
      await api.put(`/timesheets/${id}/reject`, { reason: reason.trim() });
      await loadPending();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Failed to reject timesheet");
    } finally {
      setActionLoading((prev) => ({ ...prev, [id]: false }));
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="flex items-center gap-2 font-mono text-sm text-ink500">
          <RefreshCw className="h-4 w-4 animate-spin text-ochre" />
          <span>Loading approval queue…</span>
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
            Timesheet Approvals
          </h1>
          <p className="mt-1 text-sm text-ink500">
            Review, verify, and approve employee timesheet submissions.
          </p>
        </div>
        <button
          onClick={loadPending}
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
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-status-verified/10 text-status-verified">
            <ShieldCheck className="h-6 w-6" />
          </div>
          <h3 className="mt-3 font-display text-base font-bold text-ink">
            All Caught Up!
          </h3>
          <p className="mt-1 text-xs text-ink500">
            There are no pending timesheets requiring manager or admin review.
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {timesheets.map((ts) => {
            const isProcessing = actionLoading[ts.id];
            const isRejectOpen = showRejectBox[ts.id];

            return (
              <div
                key={ts.id}
                className="rounded-xl border border-ink/10 bg-paper p-5 shadow-xs transition hover:border-ink/20"
              >
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div className="space-y-1.5">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs font-bold text-ink">
                        User ID: {ts.user_id.slice(0, 8)}…
                      </span>
                      <span className="rounded bg-ochre/15 px-2 py-0.5 font-mono text-[10px] font-semibold text-ochre-dark">
                        SUBMITTED
                      </span>
                    </div>
                    <div className="flex flex-wrap items-center gap-4 text-xs text-ink500">
                      <div className="flex items-center gap-1.5">
                        <Calendar className="h-3.5 w-3.5" />
                        <span>
                          {ts.period_start} → {ts.period_end}
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <Clock className="h-3.5 w-3.5" />
                        <span className="font-semibold text-ink">
                          {ts.total_hours} hrs logged
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      disabled={isProcessing}
                      onClick={() => handleApprove(ts.id)}
                      className="flex items-center gap-1.5 rounded bg-status-verified px-3.5 py-1.5 text-xs font-semibold text-paper shadow-xs transition hover:bg-status-verified/90 disabled:opacity-50"
                    >
                      <Check className="h-3.5 w-3.5" />
                      <span>{isProcessing ? "Approving…" : "Approve"}</span>
                    </button>

                    <button
                      type="button"
                      disabled={isProcessing}
                      onClick={() =>
                        setShowRejectBox((prev) => ({
                          ...prev,
                          [ts.id]: !prev[ts.id],
                        }))
                      }
                      className="flex items-center gap-1.5 rounded border border-status-rejected/30 bg-status-rejected/10 px-3.5 py-1.5 text-xs font-semibold text-status-rejected transition hover:bg-status-rejected/20 disabled:opacity-50"
                    >
                      <X className="h-3.5 w-3.5" />
                      <span>Reject</span>
                    </button>
                  </div>
                </div>

                {/* Rejection input drawer */}
                {isRejectOpen && (
                  <div className="mt-4 border-t border-ink/10 pt-4">
                    <label className="mb-1 block text-xs font-medium text-ink500">
                      Rejection Reason (Required)
                    </label>
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={rejectReason[ts.id] || ""}
                        onChange={(e) =>
                          setRejectReason((prev) => ({
                            ...prev,
                            [ts.id]: e.target.value,
                          }))
                        }
                        placeholder="Explain why this timesheet was rejected…"
                        className="flex-1 rounded border border-ink/15 bg-paper px-3 py-1.5 text-xs text-ink outline-none focus:border-status-rejected"
                      />
                      <button
                        type="button"
                        disabled={isProcessing || !rejectReason[ts.id]?.trim()}
                        onClick={() => handleReject(ts.id)}
                        className="rounded bg-status-rejected px-3.5 py-1.5 text-xs font-semibold text-paper shadow-xs hover:bg-status-rejected/90 disabled:opacity-50"
                      >
                        Confirm Rejection
                      </button>
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
