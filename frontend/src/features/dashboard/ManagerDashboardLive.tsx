import { useEffect, useState } from "react";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

interface DashboardData {
  team_attendance_today: { present: number; absent: number; total: number };
  task_progress: Record<string, number>;
  pending_approvals: number;
}

export default function ManagerDashboardLive() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`${API_BASE_URL}/dashboard/manager`)
      .then((r) => r.json())
      .then((json) => { setData(json.data); setLoading(false); });
  }, []);

  if (loading) return <div className="p-6 text-sm text-gray-500">Loading...</div>;
  if (!data) return null;

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-xl font-semibold">Manager Dashboard</h1>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="border rounded-lg p-4">
          <h2 className="text-sm font-medium text-gray-500">Attendance Today</h2>
          <p className="text-2xl font-semibold mt-1">
            {data.team_attendance_today.present} / {data.team_attendance_today.total}
          </p>
          <p className="text-xs text-gray-400 mt-1">{data.team_attendance_today.absent} absent</p>
        </div>
        <div className="border rounded-lg p-4">
          <h2 className="text-sm font-medium text-gray-500">Task Progress</h2>
          <ul className="mt-1 text-sm space-y-1">
            {Object.entries(data.task_progress).map(([status, count]) => (
              <li key={status}>{status.replace("_", " ")}: {count}</li>
            ))}
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
