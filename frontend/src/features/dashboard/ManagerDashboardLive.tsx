import { useEffect, useState } from "react";
import api from "../../lib/api/client";
import { Users, CheckCircle2, Clock, ShieldCheck, AlertCircle, RefreshCw } from "lucide-react";

interface DashboardData {
  team_attendance_today: { present: number; absent: number; total: number };
  task_progress: Record<string, number>;
  pending_approvals: number;
}

export default function ManagerDashboardLive() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  async function loadDashboard() {
    setLoading(true);
    setError(null);
    try {
      const res = await api.get("/dashboard/manager");
      setData(res.data?.data ?? res.data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load dashboard metrics");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadDashboard();
  }, []);

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="flex items-center gap-2 font-mono text-sm text-ink500">
          <RefreshCw className="h-4 w-4 animate-spin text-ochre" />
          <span>Loading live management metrics…</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg border border-status-rejected/30 bg-status-rejected/10 p-6 text-center">
        <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-full bg-status-rejected/20 text-status-rejected">
          <AlertCircle className="h-5 w-5" />
        </div>
        <h3 className="mt-3 font-display text-base font-bold text-status-rejected">
          Dashboard Unavailable
        </h3>
        <p className="mt-1 text-xs text-ink500">{error}</p>
        <button
          onClick={loadDashboard}
          className="mt-4 rounded bg-ink px-4 py-1.5 text-xs font-medium text-paper hover:bg-ink-light"
        >
          Try Again
        </button>
      </div>
    );
  }

  if (!data) return null;

  const totalTasks = Object.values(data.task_progress || {}).reduce((a, b) => a + b, 0);

  return (
    <div className="space-y-6">
      {/* Title */}
      <div className="flex items-center justify-between border-b border-ink/10 pb-4">
        <div>
          <h1 className="font-display text-2xl font-bold tracking-tight text-ink">
            Management Overview
          </h1>
          <p className="mt-1 text-sm text-ink500">
            Real-time workforce attendance, task distribution, and approval queue KPIs.
          </p>
        </div>
        <button
          onClick={loadDashboard}
          className="flex items-center gap-1.5 rounded border border-ink/15 bg-paper px-3 py-1.5 text-xs font-medium text-ink transition hover:bg-ink/5"
        >
          <RefreshCw className="h-3.5 w-3.5" />
          <span>Refresh</span>
        </button>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-3">
        {/* Attendance Today */}
        <div className="rounded-xl border border-ink/10 bg-paper p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-ink500">
              Team Attendance Today
            </span>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-status-verified/10 text-status-verified">
              <Users className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="font-display text-3xl font-bold text-ink">
              {data.team_attendance_today?.present ?? 0}
            </span>
            <span className="text-sm font-medium text-ink500">
              / {data.team_attendance_today?.total ?? 0} present
            </span>
          </div>
          <div className="mt-2 flex items-center gap-2">
            <span className="inline-block h-2 w-2 rounded-full bg-status-verified"></span>
            <span className="text-xs text-ink500">
              {data.team_attendance_today?.absent ?? 0} team member(s) offline/absent
            </span>
          </div>
        </div>

        {/* Total Tasks Active */}
        <div className="rounded-xl border border-ink/10 bg-paper p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-ink500">
              Total Tracked Tasks
            </span>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-ochre/15 text-ochre-dark">
              <Clock className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="font-display text-3xl font-bold text-ink">
              {totalTasks}
            </span>
            <span className="text-sm font-medium text-ink500">tasks registered</span>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-ink500">
            <CheckCircle2 className="h-3.5 w-3.5 text-ochre" />
            <span>
              {data.task_progress?.in_progress ?? 0} active, {data.task_progress?.completed ?? 0} completed
            </span>
          </div>
        </div>

        {/* Pending Approvals */}
        <div className="rounded-xl border border-ink/10 bg-paper p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-ink500">
              Timesheet Approvals
            </span>
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-ink/10 text-ink">
              <ShieldCheck className="h-4 w-4" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="font-display text-3xl font-bold text-ink">
              {data.pending_approvals ?? 0}
            </span>
            <span className="text-sm font-medium text-ink500">pending review</span>
          </div>
          <p className="mt-2 text-xs text-ink500">
            Review and finalize submitted employee timesheets in Approvals tab.
          </p>
        </div>
      </div>

      {/* Task Distribution Breakdown */}
      <div className="rounded-xl border border-ink/10 bg-paper p-6 shadow-xs">
        <h2 className="font-display text-base font-bold text-ink">
          Task Distribution by Status
        </h2>
        <div className="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-5">
          {Object.entries(data.task_progress || {}).map(([status, count]) => {
            const label = status.replace("_", " ").toUpperCase();
            return (
              <div
                key={status}
                className="rounded-lg border border-ink/10 bg-paper-dim/40 p-4 text-center"
              >
                <span className="font-mono text-[11px] font-semibold text-ink500">
                  {label}
                </span>
                <p className="mt-1 font-display text-2xl font-bold text-ink">
                  {count}
                </p>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
