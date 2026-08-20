import { db } from '../../lib/offline-db/db';

export async function syncOfflineLogs() {
  if (!navigator.onLine) return;

  const pendingLogs = await db.offlineLogs
    .where('sync_status')
    .equals('OFFLINE_LOGGED')
    .toArray();

  if (pendingLogs.length === 0) return;

  for (const log of pendingLogs) {
    try {
      // POST to backend batch/sync endpoint
      const response = await fetch('/api/v1/attendance/sync', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(log)
      });

      if (response.ok) {
        // Update local status to SYNCED_VERIFIED
        await db.offlineLogs.update(log.id!, { sync_status: 'SYNCED_VERIFIED' });
      }
    } catch (err) {
      console.error("Sync failed for log:", log.record_uuid, err);
    }
  }
}

// Auto-trigger sync when coming back online
window.addEventListener('online', () => {
  syncOfflineLogs();
});