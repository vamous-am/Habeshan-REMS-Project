import { db, OfflineAttendanceLog } from '../../lib/offline-db/db';
import {
  generateDeviceHash,
  decryptAttendancePayload,
  getExistingDeviceKey,
  AttendanceSensitivePayload,
} from '../../lib/offline-db/crypto';
import { getAuthToken, getCurrentUserId, getCurrentOrgId } from '../../lib/api/authClient';

export { generateDeviceHash };

interface DecryptedSyncCandidate {
  localLog: OfflineAttendanceLog;
  payload: AttendanceSensitivePayload;
}

export async function syncOfflineLogs(): Promise<void> {
  if (!navigator.onLine) return;

  const currentUserId = getCurrentUserId();
  const currentOrgId = getCurrentOrgId();

  // If user is not authenticated, cannot authorize or route sync records
  if (!currentUserId || !currentOrgId) {
    return;
  }

  // Retrieve device encryption key without generating a replacement
  const key = await getExistingDeviceKey();

  // Pull all records awaiting sync
  const pendingLogs = await db.offlineLogs
    .where('sync_status')
    .equals('OFFLINE_LOGGED')
    .toArray();

  if (pendingLogs.length === 0) return;

  const candidates: DecryptedSyncCandidate[] = [];

  for (const log of pendingLogs) {
    // Application-level safeguard: process ONLY records belonging to the current session owner
    if (log.user_id !== currentUserId || log.org_id !== currentOrgId) {
      // Leave quarantined for the actual owner session; never attempt cross-user decryption
      continue;
    }

    // Handle encrypted v2 record
    if (log.ciphertext && log.iv) {
      if (!key) {
        console.warn(`[SyncWorker] Device key unavailable for record ${log.record_uuid}; preserving ciphertext.`);
        if (log.id && log.decrypt_error !== 'KEY_UNAVAILABLE') {
          await db.offlineLogs.update(log.id, { decrypt_error: 'KEY_UNAVAILABLE' });
        }
        continue;
      }

      try {
        const decrypted = await decryptAttendancePayload(
          log.record_uuid,
          log.org_id,
          log.user_id,
          log.iv,
          log.ciphertext,
          key
        );

        candidates.push({
          localLog: log,
          payload: decrypted,
        });
      } catch (decErr) {
        console.error(`[SyncWorker] Decryption failed for record ${log.record_uuid}; preserving record:`, decErr);
        if (log.id) {
          await db.offlineLogs.update(log.id, {
            decrypt_error: decErr instanceof Error ? decErr.message : 'DECRYPTION_FAILED',
          });
        }
        // Do NOT drop or mark synced; preserve intact for diagnosis/recovery
        continue;
      }
    } else if (log.action_type && log.timestamp && log.device_hash) {
      // Legacy v1 plaintext record
      candidates.push({
        localLog: log,
        payload: {
          action_type: log.action_type,
          timestamp: log.timestamp,
          device_hash: log.device_hash,
        },
      });
    }
  }

  if (candidates.length === 0) return;

  // Optimistically set state to PENDING_SYNC for candidates to transmit
  const transmittingIds = candidates
    .map((c) => c.localLog.id)
    .filter((id): id is number => id !== undefined);

  await db.offlineLogs.where('id').anyOf(transmittingIds).modify({ sync_status: 'PENDING_SYNC' });

  const syncRecords = candidates.map((c) => ({
    record_uuid: c.localLog.record_uuid,
    org_id: c.localLog.org_id,
    user_id: c.localLog.user_id,
    action_type: c.payload.action_type,
    timestamp: c.payload.timestamp,
    device_hash: c.payload.device_hash,
  }));

  try {
    const token = getAuthToken() || localStorage.getItem('token') || '';
    const response = await fetch('/api/v1/attendance/sync', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ records: syncRecords }),
    });

    if (response.ok) {
      const data = await response.json();
      const results = data.results || data.data?.results || [];

      for (const res of results) {
        const candidate = candidates.find((c) => c.localLog.record_uuid === res.record_uuid);
        if (candidate && candidate.localLog.id) {
          if (res.status === 'SYNCED_VERIFIED' || res.status === 'ALREADY_SYNCED') {
            await db.offlineLogs.update(candidate.localLog.id, {
              sync_status: 'SYNCED_VERIFIED',
              decrypt_error: undefined,
            });
          } else if (res.status === 'REJECTED_TAMPERED') {
            await db.offlineLogs.update(candidate.localLog.id, {
              sync_status: 'REJECTED_TAMPERED',
            });
          }
        }
      }
    } else {
      // Revert to OFFLINE_LOGGED if server returns error status
      await db.offlineLogs.where('id').anyOf(transmittingIds).modify({ sync_status: 'OFFLINE_LOGGED' });
    }
  } catch (err) {
    console.error('Batch sync network request failed:', err);
    // Revert state on connection / network error
    await db.offlineLogs.where('id').anyOf(transmittingIds).modify({ sync_status: 'OFFLINE_LOGGED' });
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('online', () => {
    syncOfflineLogs();
  });
}