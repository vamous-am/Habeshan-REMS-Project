import React, { useEffect, useState, useCallback } from 'react';

interface AttendanceLogRecord {
  record_uuid: string;
  user_id: string;
  user_name?: string;
  clock_in: string;
  clock_out?: string;
  sync_status: string;
}

export const AttendanceHistoryScoped: React.FC = () => {
  const [records, setRecords] = useState<AttendanceLogRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const userRole = (localStorage.getItem('user_role') as 'EMPLOYEE' | 'MANAGER' | 'ADMIN') || 'EMPLOYEE';
  const [scope, setScope] = useState<'me' | 'team' | 'org'>(
    userRole === 'ADMIN' ? 'org' : userRole === 'MANAGER' ? 'team' : 'me'
  );

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const endpoint =
        scope === 'org'
          ? '/api/v1/attendance/org'
          : scope === 'team'
          ? '/api/v1/attendance/team'
          : '/api/v1/attendance/me';

      const res = await fetch(endpoint, {
        headers: { Authorization: `Bearer ${localStorage.getItem('token') || ''}` },
      });
      if (res.ok) {
        const data = await res.json();
        setRecords(data.records || data.data || []);
      }
    } catch (err) {
      console.error('Failed to load attendance logs:', err);
    } finally {
      setLoading(false);
    }
  }, [scope]);

  useEffect(() => {
    // Schedule on microtask queue to prevent synchronous state setter execution during effect setup
    const timer = setTimeout(() => {
      void fetchLogs();
    }, 0);

    return () => clearTimeout(timer);
  }, [fetchLogs]);

  return (
    <div className="w-full bg-white rounded-2xl shadow-lg border border-slate-100 p-6 font-sans space-y-4">
      <div className="flex items-center justify-between flex-wrap gap-4">
        <div>
          <span className="text-[10px] font-bold uppercase tracking-wider text-indigo-600 block">
            FR-ATT-09 View Scope
          </span>
          <h3 className="text-lg font-bold text-slate-800">Attendance History</h3>
        </div>

        <div className="flex items-center gap-1.5 bg-slate-100 p-1 rounded-xl text-xs">
          <button
            onClick={() => setScope('me')}
            className={`px-3 py-1.5 rounded-lg font-semibold transition ${
              scope === 'me' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-200'
            }`}
          >
            My Logs
          </button>
          {userRole !== 'EMPLOYEE' && (
            <button
              onClick={() => setScope('team')}
              className={`px-3 py-1.5 rounded-lg font-semibold transition ${
                scope === 'team' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-200'
              }`}
            >
              Team View
            </button>
          )}
          {userRole === 'ADMIN' && (
            <button
              onClick={() => setScope('org')}
              className={`px-3 py-1.5 rounded-lg font-semibold transition ${
                scope === 'org' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-200'
              }`}
            >
              Org View
            </button>
          )}
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-slate-600">
          <thead className="bg-slate-50 uppercase text-slate-400 font-semibold border-b border-slate-100">
            <tr>
              {scope !== 'me' && <th className="p-3">Employee</th>}
              <th className="p-3">Record UUID</th>
              <th className="p-3">Clock In</th>
              <th className="p-3">Clock Out</th>
              <th className="p-3 text-right">Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {loading ? (
              <tr>
                <td colSpan={5} className="p-6 text-center text-slate-400">
                  Loading attendance records...
                </td>
              </tr>
            ) : records.length === 0 ? (
              <tr>
                <td colSpan={5} className="p-6 text-center text-slate-400">
                  No records found for current scope.
                </td>
              </tr>
            ) : (
              records.map((rec) => (
                <tr key={rec.record_uuid} className="hover:bg-slate-50">
                  {scope !== 'me' && <td className="p-3 font-semibold text-slate-700">{rec.user_name || rec.user_id}</td>}
                  <td className="p-3 font-mono text-slate-400">{rec.record_uuid.slice(0, 8)}...</td>
                  <td className="p-3 font-semibold text-slate-700">{new Date(rec.clock_in).toLocaleString()}</td>
                  <td className="p-3">{rec.clock_out ? new Date(rec.clock_out).toLocaleString() : 'Active'}</td>
                  <td className="p-3 text-right">
                    <span className="px-2.5 py-1 rounded-full text-[10px] font-bold bg-indigo-100 text-indigo-700">
                      {rec.sync_status}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};