import { useEffect, useState } from "react";
import api from "../../lib/api/client";
import { Trophy, RefreshCw, AlertCircle } from "lucide-react";

// LeaderboardEntry mirrors the backend response shape from GET /dashboard/leaderboard.
// opted_out has been removed — opt-out is deferred as a separate backend feature
// (requires persisted org-wide preference; FR-DASH-04 addendum).
interface LeaderboardEntry {
  rank: number;
  employee_name: string;
  completed_count: number;
  total_count: number;
  completion_percentage: number;
}

interface LeaderboardData {
  entries: LeaderboardEntry[];
}

const MEDAL_COLORS: Record<number, string> = {
  1: "text-yellow-500",
  2: "text-slate-400",
  3: "text-amber-600",
};

const MEDAL_LABELS: Record<number, string> = {
  1: "🥇",
  2: "🥈",
  3: "🥉",
};

export default function Leaderboard() {
  const [data, setData] = useState<LeaderboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    api
      .get<{ data: LeaderboardData }>("/dashboard/leaderboard")
      .then((res) => {
        if (active) {
          setData(res.data?.data ?? res.data);
        }
      })
      .catch((err: unknown) => {
        if (active) {
          setError(
            err instanceof Error ? err.message : "Failed to load leaderboard",
          );
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  function load() {
    setLoading(true);
    setError(null);
    api
      .get<{ data: LeaderboardData }>("/dashboard/leaderboard")
      .then((res) => {
        setData(res.data?.data ?? res.data);
      })
      .catch((err: unknown) => {
        setError(
          err instanceof Error ? err.message : "Failed to load leaderboard",
        );
      })
      .finally(() => {
        setLoading(false);
      });
  }

  if (loading) {
    return (
      <div className="flex h-40 items-center justify-center">
        <div className="flex items-center gap-2 font-mono text-sm text-ink500">
          <RefreshCw className="h-4 w-4 animate-spin text-ochre" />
          <span>Loading leaderboard…</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-xl border border-status-rejected/30 bg-status-rejected/10 p-6 text-center">
        <AlertCircle className="mx-auto h-5 w-5 text-status-rejected" />
        <p className="mt-2 text-xs text-ink500">{error}</p>
        <button
          onClick={load}
          className="mt-3 rounded bg-ink px-4 py-1.5 text-xs font-medium text-paper hover:bg-ink-light"
        >
          Retry
        </button>
      </div>
    );
  }

  const entries = data?.entries ?? [];

  return (
    <div className="rounded-xl border border-ink/10 bg-paper p-6 shadow-xs">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-ink/10 pb-4">
        <div className="flex items-center gap-2">
          <Trophy className="h-5 w-5 text-ochre" />
          <h2 className="font-display text-base font-bold text-ink">
            Top Performers
          </h2>
        </div>
        <button
          onClick={load}
          id="leaderboard-refresh-btn"
          aria-label="Refresh leaderboard"
          className="flex items-center gap-1.5 rounded border border-ink/15 bg-paper px-3 py-1.5 text-xs font-medium text-ink transition hover:bg-ink/5"
        >
          <RefreshCw className="h-3.5 w-3.5" />
          <span>Refresh</span>
        </button>
      </div>

      {/* Sub-heading */}
      <p className="mt-3 text-xs text-ink500">
        Ranked by task completion percentage (all-time). Top 3 employees in
        your organization.
      </p>

      {entries.length === 0 ? (
        <div className="mt-6 rounded-lg border border-dashed border-ink/20 p-6 text-center">
          <p className="text-sm text-ink500">
            No assigned tasks found. Leaderboard will populate once tasks are
            assigned and completed.
          </p>
        </div>
      ) : (
        <div className="mt-4 space-y-3">
          {entries.map((entry) => {
            const pct = Math.round(entry.completion_percentage);
            return (
              <div
                key={entry.rank}
                id={`leaderboard-entry-rank-${entry.rank}`}
                className="flex items-center gap-4 rounded-lg border border-ink/8 bg-paper-dim/30 px-4 py-3 transition hover:bg-ink/3"
              >
                {/* Medal / rank */}
                <span
                  className={`w-8 text-center text-lg ${MEDAL_COLORS[entry.rank] ?? "text-ink500"}`}
                  aria-label={`Rank ${entry.rank}`}
                >
                  {MEDAL_LABELS[entry.rank] ?? `#${entry.rank}`}
                </span>

                {/* Name */}
                <span className="flex-1 truncate font-medium text-ink">
                  {entry.employee_name}
                </span>

                {/* Progress bar + percentage */}
                <div className="flex items-center gap-3">
                  <div className="hidden w-24 overflow-hidden rounded-full bg-ink/10 sm:block">
                    <div
                      className="h-1.5 rounded-full bg-ochre transition-all"
                      style={{ width: `${Math.min(pct, 100)}%` }}
                    />
                  </div>
                  <span className="w-14 text-right font-mono text-xs font-semibold text-ink">
                    {pct}%
                  </span>
                  <span className="hidden w-24 text-right text-xs text-ink500 sm:block">
                    {entry.completed_count}/{entry.total_count} tasks
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}