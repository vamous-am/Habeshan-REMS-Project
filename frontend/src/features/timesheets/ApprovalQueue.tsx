import { useEffect, useState } from "react";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

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
  const [rejectReason, setRejectReason] = useState<Record<string, string>>({});

  async function fetchPending() {
    const res = await fetch(`${API_BASE_URL}/timesheets?status=submitted`);
    const json = await res.json();
    setTimesheets(json.data ?? []);
    setLoading(false);
  }

  useEffect(() => { fetchPending(); }, []);

  async function handleApprove(id: string) {
    await fetch(`${API_BASE_URL}/timesheets/${id}/approve`, { method: "PUT" });
    fetchPending();
  }

  async function handleReject(id: string) {
    const reason = rejectReason[id];
    if (!reason) return alert("Rejection reason is required.");
    await fetch(`${API_BASE_URL}/timesheets/${id}/reject`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ reason }),
    });
    fetchPending();
  }

  if (loading) return <div className="p-6 text-sm text-gray-500">Loading...</div>;

  return (
    <div className="p-6 space-y-4">
      <h1 className="text-xl font-semibold">Approval Queue</h1>
      {timesheets.length === 0 && (
        <p className="text-sm text-gray-400">No pending timesheets.</p>
      )}
      {timesheets.map((ts) => (
        <div key={ts.id} className="border rounded-lg p-4 space-y-3">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">
              {ts.period_start} to {ts.period_end}
            </p>
            <span className="text-xs bg-blue-100 text-blue-600 px-2 py-1 rounded-full font-semibold">
              submitted
            </span>
          </div>
          <p className="text-sm text-gray-500">Total hours: {ts.total_hours.toFixed(2)}</p>
          <input
            type="text"
            placeholder="Rejection reason (required to reject)"
            value={rejectReason[ts.id] ?? ""}
            onChange={(e) => setRejectReason((prev) => ({ ...prev, [ts.id]: e.target.value }))}
            className="w-full border rounded px-3 py-1 text-sm"
          />
          <div className="flex gap-2">
            <button
              onClick={() => handleApprove(ts.id)}
              className="text-xs bg-green-600 text-white px-3 py-1 rounded"
            >
              Approve
            </button>
            <button
              onClick={() => handleReject(ts.id)}
              className="text-xs bg-red-600 text-white px-3 py-1 rounded"
            >
              Reject
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
