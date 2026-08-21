import { db } from '../../lib/offline-db/db';

// Computes standard SHA-256 hash in browser environment
export async function generateDeviceHash(
  recordUuid: string,
  userId: string,
  actionType: string,
  timestamp: string
): Promise<string> {
  const rawString = `${recordUuid}|${userId}|${actionType}|${timestamp}`;
  const encoder = new TextEncoder();
  const data = encoder.encode(rawString);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}

export async function syncOfflineLogs(): Promise<void> {
  if (!navigator.onLine) return;

  const pendingLogs = await db.offlineLogs
    .where('sync_status')
    .equals('OFFLINE_LOGGED')
    .toArray();

  if (pendingLogs.length === 0) return;

  // Optimistically set state to PENDING_SYNC
  const logIds = pendingLogs.map((l) => l.id).filter((id): id is number => id !== undefined);
  await db.offlineLogs.where('id').anyOf(logIds).modify({ sync_status: 'PENDING_SYNC' });

  try {
    const response = await fetch('/api/v1/attendance/sync', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
      },
      body: JSON.stringify({ records: pendingLogs }),
    });

    if (response.ok) {
      const data = await response.json();
      const results = data.results || data.data?.results || [];

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
    } else {
      // Revert to OFFLINE_LOGGED if server returns error status
      await db.offlineLogs.where('id').anyOf(logIds).modify({ sync_status: 'OFFLINE_LOGGED' });
    }
  } catch (err) {
    console.error('Batch sync network request failed:', err);
    // Revert state on connection / network error
    await db.offlineLogs.where('id').anyOf(logIds).modify({ sync_status: 'OFFLINE_LOGGED' });
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('online', () => {
    syncOfflineLogs();
  });
}