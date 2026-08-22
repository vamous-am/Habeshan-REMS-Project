import { useEffect, useState } from "react";
import { getCurrentUserId } from "../../lib/api/authClient";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

interface Timesheet {
  id: string;
  period_start: string;
  period_end: string;
  total_hours: number;
  status: "draft" | "submitted" | "approved" | "rejected";
  rejection_reason?: string;
}

const statusColors: Record<string, string> = {
  draft: "bg-gray-100 text-gray-600",
  submitted: "bg-blue-100 text-blue-600",
  approved: "bg-green-100 text-green-700",
  rejected: "bg-red-100 text-red-700",
};

export default function TimesheetList() {
  const [timesheets, setTimesheets] = useState<Timesheet[]>([]);
  const [loading, setLoading] = useState(true);

  async function fetchTimesheets() {
    const userId = getCurrentUserId();
    if (!userId) {
      setTimesheets([]);
      setLoading(false);
      return;
    }

    const res = await fetch(`${API_BASE_URL}/timesheets?user_id=${userId}`);
    const json = await res.json();
    setTimesheets(json.data ?? []);
    setLoading(false);
  }

  useEffect(() => { fetchTimesheets(); }, []);

  async function handleSubmit(id: string) {
    await fetch(`${API_BASE_URL}/timesheets/${id}/submit`, { method: "PUT" });
    fetchTimesheets();
  }

  if (loading) return <div className="p-6 text-sm text-gray-500">Loading...</div>;

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-xl font-semibold">My Timesheets</h1>
      {timesheets.length === 0 && (
        <p className="text-sm text-gray-400">No timesheets yet.</p>
      )}
      {timesheets.map((ts) => (
        <div key={ts.id} className="border rounded-lg p-4 space-y-2">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">
              {ts.period_start} to {ts.period_end}
            </p>
            <span className={`text-xs font-semibold px-2 py-1 rounded-full ${statusColors[ts.status]}`}>
              {ts.status}
            </span>
          </div>
          <p className="text-sm text-gray-500">Total hours: {ts.total_hours.toFixed(2)}</p>
          {ts.rejection_reason && (
            <p className="text-xs text-red-600">Rejection reason: {ts.rejection_reason}</p>
          )}
          {(ts.status === "draft" || ts.status === "rejected") && (
            <button
              onClick={() => handleSubmit(ts.id)}
              className="text-xs bg-blue-600 text-white px-3 py-1 rounded"
            >
              Submit
            </button>
          )}
        </div>
      ))}
    </div>
  );
}
