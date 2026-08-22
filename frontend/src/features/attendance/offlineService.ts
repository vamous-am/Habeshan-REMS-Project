import { db, OfflineAttendanceLog } from '../../lib/offline-db/db';
import { generateDeviceHash } from '../../lib/offline-db/crypto';

export async function recordOfflineAttendance(
  orgId: string,
  userId: string,
  actionType: 'CLOCK_IN' | 'CLOCK_OUT'
): Promise<OfflineAttendanceLog> {
  const recordUuid = crypto.randomUUID();
  const timestamp = new Date().toISOString();

  // Generate SHA-256 tamper verification hash
  const deviceHash = await generateDeviceHash(recordUuid, userId, actionType, timestamp);

  const newLog: OfflineAttendanceLog = {
    record_uuid: recordUuid,
    org_id: orgId,
    user_id: userId,
    action_type: actionType,
    timestamp: timestamp,
    sync_status: 'OFFLINE_LOGGED',
    device_hash: deviceHash,
    created_at: timestamp
  };

  // Add payload to Dexie IndexedDB
  await db.offlineLogs.add(newLog);
  return newLog;
}