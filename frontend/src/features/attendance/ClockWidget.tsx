import React, { useState, useEffect } from 'react';
import { clockInApi, clockOutApi, AttendanceRecord } from './attendanceApi';
import './ClockWidget.css'; 

export const ClockWidget: React.FC = () => {
  const [activeSession, setActiveSession] = useState<AttendanceRecord | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [currentTime, setCurrentTime] = useState<Date>(new Date());

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  const handleClockIn = async () => {
    setLoading(true);
    setError(null);
    try {
      const record = await clockInApi();
      setActiveSession(record);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to record clock-in.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const handleClockOut = async () => {
    setLoading(true);
    setError(null);
    try {
      const record = await clockOutApi();
      setActiveSession(record);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to record clock-out.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const isClockedIn = Boolean(activeSession && !activeSession.clock_out);

  return (
    <div className="w-full max-w-lg mx-auto bg-white rounded-2xl shadow-lg border border-slate-100 overflow-hidden font-sans">
      <div className="bg-gradient-to-r from-slate-900 to-indigo-950 p-6 text-white flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Attendance Center</h2>
          <p className="text-xs text-slate-300 mt-1">
            {currentTime.toLocaleDateString(undefined, {
              weekday: 'long',
              year: 'numeric',
              month: 'short',
              day: 'numeric',
            })}
          </p>
        </div>
        <div className="text-right">
          <div className="text-2xl font-mono font-bold tracking-wider">
            {currentTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
          </div>
          <div className="text-[10px] uppercase tracking-widest text-indigo-300 font-semibold mt-0.5">
            Local Time
          </div>
        </div>
      </div>

      <div className="p-6 space-y-6">
        {error && (
          <div className="p-3 bg-rose-50 border border-rose-200 text-rose-700 text-xs rounded-xl flex items-center gap-2">
            <span className="font-bold">Error:</span> {error}
          </div>
        )}

        <div className="flex items-center justify-between p-4 bg-slate-50 rounded-xl border border-slate-100">
          <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
            Work Status
          </span>
          <div className="flex items-center gap-2">
            <span
              className={`h-2.5 w-2.5 rounded-full ${
                isClockedIn ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'
              }`}
            />
            <span
              className={`text-xs font-bold px-2.5 py-1 rounded-full ${
                isClockedIn
                  ? 'bg-emerald-100 text-emerald-800'
                  : 'bg-slate-200 text-slate-700'
              }`}
            >
              {isClockedIn ? 'ON THE CLOCK' : 'CLOCKED OUT'}
            </span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3 text-xs">
          <div className="p-3 bg-slate-50 rounded-xl border border-slate-100">
            <span className="text-slate-400 font-medium block mb-1">Clock In</span>
            <span className="text-slate-800 font-bold text-sm">
              {activeSession?.clock_in
                ? new Date(activeSession.clock_in).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                : '--:--'}
            </span>
          </div>

          <div className="p-3 bg-slate-50 rounded-xl border border-slate-100">
            <span className="text-slate-400 font-medium block mb-1">Clock Out</span>
            <span className="text-slate-800 font-bold text-sm">
              {activeSession?.clock_out
                ? new Date(activeSession.clock_out).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                : '--:--'}
            </span>
          </div>

          <div className="p-3 bg-slate-50 rounded-xl border border-slate-100">
            <span className="text-slate-400 font-medium block mb-1">Total Duration</span>
            <span className="text-slate-800 font-bold text-sm">
              {activeSession?.total_hours !== undefined
                ? `${activeSession.total_hours.toFixed(2)} hrs`
                : '0.00 hrs'}
            </span>
          </div>

          <div className="p-3 bg-slate-50 rounded-xl border border-slate-100">
            <span className="text-slate-400 font-medium block mb-1">Sync State</span>
            <span className="text-indigo-600 font-bold text-xs uppercase tracking-wider">
              {activeSession?.sync_status || 'ONLINE'}
            </span>
          </div>
        </div>

        {!isClockedIn ? (
          <button
            onClick={handleClockIn}
            disabled={loading}
            className="w-full py-3.5 bg-indigo-600 hover:bg-indigo-700 active:bg-indigo-800 text-white font-semibold text-sm rounded-xl shadow-md transition-all duration-150 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading ? (
              <span className="inline-block animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent" />
            ) : (
              <span>Clock In</span>
            )}
          </button>
        ) : (
          <button
            onClick={handleClockOut}
            disabled={loading}
            className="w-full py-3.5 bg-amber-600 hover:bg-amber-700 active:bg-amber-800 text-white font-semibold text-sm rounded-xl shadow-md transition-all duration-150 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading ? (
              <span className="inline-block animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent" />
            ) : (
              <span>Clock Out</span>
            )}
          </button>
        )}
      </div>
    </div>
  );
};