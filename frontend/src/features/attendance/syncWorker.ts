import { db } from '../../lib/offline-db/db';

export async function syncOfflineLogs(): Promise<void> {
  if (!navigator.onLine) return;

  const pendingLogs = await db.offlineLogs
    .where('sync_status')
    .equals('OFFLINE_LOGGED')
    .toArray();

  if (pendingLogs.length === 0) return;

  try {
    const response = await fetch('/api/v1/attendance/sync', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${localStorage.getItem('token') || ''}`
      },
      body: JSON.stringify({ records: pendingLogs })
    });

    if (response.ok) {
      const data = await response.json();
      const results = data.data?.results || [];

      for (const res of results) {
        const localLog = pendingLogs.find((l) => l.record_uuid === res.record_uuid);
        if (localLog && localLog.id) {
          if (res.status === 'SYNCED_VERIFIED' || res.status === 'ALREADY_SYNCED') {
            await db.offlineLogs.update(localLog.id, { sync_status: 'SYNCED_VERIFIED' });
          } else if (res.status === 'REJECTED_TAMPERED') {
            await db.offlineLogs.update(localLog.id, { sync_status: 'REJECTED_TAMPERED' });
          }
        }
      }
    }
  } catch (err) {
    console.error('Batch sync network request failed:', err);
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('online', () => {
    syncOfflineLogs();
  });
}