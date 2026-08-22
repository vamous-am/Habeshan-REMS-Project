import { useEffect, useState } from "react";

interface LeaderboardEntry {
  rank: number;
  employee_name: string;
  completion_percentage: number;
}

interface LeaderboardData {
  opted_out: boolean;
  entries: LeaderboardEntry[];
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

export default function Leaderboard() {
  const [data, setData] = useState<LeaderboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [optedOut, setOptedOut] = useState(false);

  useEffect(() => {
    async function fetchLeaderboard() {
      try {
        const res = await fetch(
          `${API_BASE_URL}/dashboard/leaderboard`,
        );

        if (!res.ok) {
          throw new Error(`Request failed with status ${res.status}`);
        }

        const json = await res.json();

        setData(json.data);
        setOptedOut(json.data.opted_out);
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "Unknown error",
        );
      } finally {
        setLoading(false);
      }
    }

    fetchLeaderboard();
  }, []);

  if (loading) {
    return (
      <div className="p-6 text-sm text-gray-500">
        Loading leaderboard…
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6 text-sm text-red-600">
        Failed to load leaderboard: {error}
      </div>
    );
  }

  if (!data) {
    return null;
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">
            Leaderboard
          </h1>

          <p className="text-sm text-gray-500 mt-1">
            Top performers based on task completion.
          </p>
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={optedOut}
            onChange={(event) =>
              setOptedOut(event.target.checked)
            }
          />

          Opt out of leaderboard
        </label>
      </div>

      {optedOut ? (
        <div className="border rounded-lg p-6 text-sm text-gray-500">
          You are opted out of the leaderboard.
        </div>
      ) : (
        <div className="border rounded-lg overflow-hidden">
          <div className="grid grid-cols-3 gap-4 border-b bg-gray-50 px-4 py-3 text-sm font-medium text-gray-500">
            <span>Rank</span>
            <span>Employee</span>
            <span>Completion</span>
          </div>

          {data.entries.map((entry) => (
            <div
              key={entry.rank}
              className="grid grid-cols-3 gap-4 border-b last:border-b-0 px-4 py-4 text-sm"
            >
              <span className="font-medium">
                #{entry.rank}
              </span>

              <span>
                {entry.employee_name}
              </span>

              <span className="font-medium">
                {entry.completion_percentage}%
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}