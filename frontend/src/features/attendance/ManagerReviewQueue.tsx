import React, { useEffect, useState } from 'react';

interface FlaggedRecord {
  record_uuid: string;
  user_id: string;
  user_name?: string;
  sync_status: string;
  timestamp?: string;
}

export const ManagerReviewQueue: React.FC = () => {
  const [flagged, setFlagged] = useState<FlaggedRecord[]>([]);

  useEffect(() => {
    fetch('/api/v1/attendance/flagged-reviews', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token') || ''}` },
    })
      .then((r) => r.json())
      .then((d) => setFlagged(d.records || d.items || []))
      .catch((e) => console.error(e));
  }, []);

  return (
    <div className="w-full bg-white rounded-2xl shadow-lg border border-rose-100 p-6 font-sans space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <span className="text-[10px] font-bold uppercase tracking-wider text-rose-600 block">
            FR-ATT-07 Flagged Queue
          </span>
          <h3 className="text-lg font-bold text-slate-800">Tampered Records Review</h3>
        </div>
        <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-rose-100 text-rose-800">
          {flagged.length} Flagged
        </span>
      </div>

      <div className="space-y-2">
        {flagged.length === 0 ? (
          <p className="text-xs text-slate-400 text-center py-4">No tampered or flagged records require review.</p>
        ) : (
          flagged.map((item) => (
            <div key={item.record_uuid} className="p-3 rounded-xl bg-rose-50/50 border border-rose-100 flex items-center justify-between text-xs">
              <div>
                <p className="font-bold text-slate-800">User: {item.user_name || item.user_id}</p>
                <p className="text-slate-400 font-mono text-[10px]">{item.record_uuid}</p>
              </div>
              <button className="px-3 py-1 bg-slate-900 text-white font-semibold rounded-lg hover:bg-slate-800">
                Review
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  );
};