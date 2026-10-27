import { db, OfflineAttendanceLog } from '../../lib/offline-db/db';
import {
  generateDeviceHash,
  encryptAttendancePayload,
  getOrCreateDeviceKey,
} from '../../lib/offline-db/crypto';

export async function recordOfflineAttendance(
  orgId: string,
  userId: string,
  actionType: 'CLOCK_IN' | 'CLOCK_OUT'
): Promise<OfflineAttendanceLog> {
  const recordUuid = crypto.randomUUID();
  const timestamp = new Date().toISOString();

  // 1. Generate SHA-256 tamper-evident hash over plaintext fields
  const deviceHash = await generateDeviceHash(recordUuid, userId, actionType, timestamp);

  // 2. Retrieve or initialize the device-held AES-256-GCM key
  const key = await getOrCreateDeviceKey();

  // 3. Encrypt the sensitive attendance payload with canonical AAD
  const encryptedBundle = await encryptAttendancePayload(
    recordUuid,
    orgId,
    userId,
    {
      action_type: actionType,
      timestamp,
      device_hash: deviceHash,
    },
    key
  );

  const newLog: OfflineAttendanceLog = {
    record_uuid: recordUuid,
    org_id: orgId,
    user_id: userId,
    sync_status: 'OFFLINE_LOGGED',
    iv: encryptedBundle.iv,
    ciphertext: encryptedBundle.ciphertext,
    crypto_version: encryptedBundle.crypto_version,
    created_at: timestamp,
  };

  // 4. Persist encrypted payload to Dexie IndexedDB
  await db.offlineLogs.add(newLog);
  return newLog;
}