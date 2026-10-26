import { db, OfflineAttendanceLog, SyncStatus } from './db';
import { base64ToBuffer } from './crypto';

const BACKUP_FORMAT_IDENTIFIER = 'habeshan-rems-offline-backup';
const BACKUP_SCHEMA_VERSION = 1;
const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const ALLOWED_SYNC_STATUSES: SyncStatus[] = [
  'OFFLINE_LOGGED',
  'PENDING_SYNC',
  'SYNCED_VERIFIED',
  'REJECTED_TAMPERED',
];

export interface EncryptedBackupRecord {
  record_uuid: string;
  org_id: string;
  user_id: string;
  sync_status: SyncStatus;
  created_at: string;
  iv: string;
  ciphertext: string;
  crypto_version: number;
  decrypt_error?: string;
}

export interface EncryptedBackupEnvelope {
  format: typeof BACKUP_FORMAT_IDENTIFIER;
  version: typeof BACKUP_SCHEMA_VERSION;
  exported_at: string;
  record_count: number;
  records: EncryptedBackupRecord[];
}

/**
 * Task 13 (Dev 2) & Step 9: Export offline queue to an encrypted backup JSON file.
 * Preserves ciphertext, IV, and metadata. Does NOT export decrypted plaintext.
 * Fulfills NFR-REL-06: Manual export/backup option for queued offline records.
 */
export async function exportOfflineQueueToJSON(): Promise<void> {
  try {
    const logs = await db.offlineLogs.toArray();
    if (!logs || logs.length === 0) {
      alert('No offline attendance records to export.');
      return;
    }

    const backupRecords: EncryptedBackupRecord[] = logs.map((log) => ({
      record_uuid: log.record_uuid,
      org_id: log.org_id,
      user_id: log.user_id,
      sync_status: log.sync_status,
      created_at: log.created_at,
      iv: log.iv || '',
      ciphertext: log.ciphertext || '',
      crypto_version: log.crypto_version || 1,
      decrypt_error: log.decrypt_error,
    }));

    const envelope: EncryptedBackupEnvelope = {
      format: BACKUP_FORMAT_IDENTIFIER,
      version: BACKUP_SCHEMA_VERSION,
      exported_at: new Date().toISOString(),
      record_count: backupRecords.length,
      records: backupRecords,
    };

    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(envelope, null, 2));
    const downloadAnchor = document.createElement('a');
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');

    downloadAnchor.setAttribute('href', dataStr);
    downloadAnchor.setAttribute('download', `offline_attendance_encrypted_backup_${timestamp}.json`);
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
  } catch (err) {
    console.error('Failed to export encrypted offline queue backup:', err);
    alert('Error generating backup file for offline queue.');
  }
}

export interface RestoreResult {
  restored: number;
  skipped: number;
  errors: string[];
}

/**
 * Validates and restores offline records from an encrypted backup JSON file.
 * Does not overwrite or delete existing records unless record_uuid already exists.
 */
export async function restoreOfflineQueueFromJSON(jsonString: string): Promise<RestoreResult> {
  const result: RestoreResult = { restored: 0, skipped: 0, errors: [] };

  let parsed: unknown;
  try {
    parsed = JSON.parse(jsonString);
  } catch {
    result.errors.push('Invalid JSON file format.');
    return result;
  }

  if (
    typeof parsed !== 'object' ||
    parsed === null ||
    (parsed as EncryptedBackupEnvelope).format !== BACKUP_FORMAT_IDENTIFIER ||
    (parsed as EncryptedBackupEnvelope).version !== BACKUP_SCHEMA_VERSION ||
    !Array.isArray((parsed as EncryptedBackupEnvelope).records)
  ) {
    result.errors.push('Unrecognized backup format or unsupported schema version.');
    return result;
  }

  const envelope = parsed as EncryptedBackupEnvelope;

  for (let idx = 0; idx < envelope.records.length; idx++) {
    const rec = envelope.records[idx];

    // 1. UUID validation
    if (!rec.record_uuid || !UUID_REGEX.test(rec.record_uuid)) {
      result.skipped++;
      result.errors.push(`Record ${idx + 1}: Invalid or missing record_uuid format.`);
      continue;
    }

    if (!rec.org_id || !rec.user_id) {
      result.skipped++;
      result.errors.push(`Record ${rec.record_uuid}: Missing org_id or user_id.`);
      continue;
    }

    // 2. Status validation
    if (!ALLOWED_SYNC_STATUSES.includes(rec.sync_status)) {
      result.skipped++;
      result.errors.push(`Record ${rec.record_uuid}: Invalid sync_status '${rec.sync_status}'.`);
      continue;
    }

    // 3. IV validation
    if (!rec.iv) {
      result.skipped++;
      result.errors.push(`Record ${rec.record_uuid}: Missing IV.`);
      continue;
    }

    try {
      const ivBuffer = base64ToBuffer(rec.iv);
      if (ivBuffer.byteLength !== 12) {
        result.skipped++;
        result.errors.push(`Record ${rec.record_uuid}: Invalid IV length (${ivBuffer.byteLength} bytes, expected 12).`);
        continue;
      }
    } catch {
      result.skipped++;
      result.errors.push(`Record ${rec.record_uuid}: Malformed base64 IV.`);
      continue;
    }

    // 4. Ciphertext validation
    if (!rec.ciphertext) {
      result.skipped++;
      result.errors.push(`Record ${rec.record_uuid}: Missing ciphertext.`);
      continue;
    }

    try {
      base64ToBuffer(rec.ciphertext);
    } catch {
      result.skipped++;
      result.errors.push(`Record ${rec.record_uuid}: Malformed base64 ciphertext.`);
      continue;
    }

    // Check if record already exists in IndexedDB
    const existing = await db.offlineLogs.where('record_uuid').equals(rec.record_uuid).first();
    if (existing) {
      result.skipped++;
      continue;
    }

    const newLog: OfflineAttendanceLog = {
      record_uuid: rec.record_uuid,
      org_id: rec.org_id,
      user_id: rec.user_id,
      sync_status: rec.sync_status,
      created_at: rec.created_at || new Date().toISOString(),
      iv: rec.iv,
      ciphertext: rec.ciphertext,
      crypto_version: rec.crypto_version || 1,
      decrypt_error: rec.decrypt_error,
    };

    await db.offlineLogs.add(newLog);
    result.restored++;
  }

  return result;
}