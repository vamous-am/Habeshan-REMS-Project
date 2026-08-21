import { useEffect, useState } from "react";

// Shape of the JSON your Go endpoint returns — TypeScript's version of a
// "contract" for the data, similar to a Python dataclass or a JS JSDoc
// type comment, but enforced at compile time.
interface DashboardData {
  team_attendance_today: {
    present: number;
    absent: number;
    total: number;
  };
  task_progress: {
    to_do: number;
    in_progress: number;
    completed: number;
  };
  pending_approvals: number;
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

export default function ManagerDashboard() {
  // useState is React's way of storing a variable that, when changed,
  // makes the component re-render. Three separate pieces of state here:
  // the data itself, whether we're still loading, and any error message.
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // useEffect runs side-effecty code (like a network fetch) after the
  // component renders. The empty array [] at the end means "only run
  // this once, when the component first mounts" — not on every render.
  useEffect(() => {
    async function fetchDashboard() {
      try {
        const res = await fetch(`${API_BASE_URL}/dashboard/manager`);
        if (!res.ok) {
          throw new Error(`Request failed with status ${res.status}`);
        }
        const json = await res.json();
        setData(json.data); // matches the Envelope shape: { status, data, message }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown error");
      } finally {
        setLoading(false);
      }
    }

    fetchDashboard();
  }, []);

  if (loading) {
    return <div className="p-6 text-sm text-gray-500">Loading dashboard…</div>;
  }

  if (error) {
    return (
      <div className="p-6 text-sm text-red-600">
        Failed to load dashboard: {error}
      </div>
    );
  }

  if (!data) {
    return null;
  }

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-xl font-semibold">Manager Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="border rounded-lg p-4">
          <h2 className="text-sm font-medium text-gray-500">Attendance Today</h2>
          <p className="text-2xl font-semibold mt-1">
            {data.team_attendance_today.present} / {data.team_attendance_today.total}
          </p>
          <p className="text-xs text-gray-400 mt-1">
            {data.team_attendance_today.absent} absent
          </p>
        </div>

        <div className="border rounded-lg p-4">
          <h2 className="text-sm font-medium text-gray-500">Task Progress</h2>
          <ul className="mt-1 text-sm space-y-1">
            <li>To Do: {data.task_progress.to_do}</li>
            <li>In Progress: {data.task_progress.in_progress}</li>
            <li>Completed: {data.task_progress.completed}</li>
          </ul>
        </div>

        <div className="border rounded-lg p-4">
          <h2 className="text-sm font-medium text-gray-500">Pending Approvals</h2>
          <p className="text-2xl font-semibold mt-1">{data.pending_approvals}</p>
        </div>
      </div>
    </div>
  );
}