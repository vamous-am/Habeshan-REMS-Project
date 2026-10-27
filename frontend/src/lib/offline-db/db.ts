import Dexie, { Table } from 'dexie';
import {
  encryptAttendancePayload,
  getOrCreateDeviceKey,
} from './crypto';

export type SyncStatus =
  | 'OFFLINE_LOGGED'
  | 'PENDING_SYNC'
  | 'SYNCED_VERIFIED'
  | 'REJECTED_TAMPERED';

export interface OfflineAttendanceLog {
  id?: number;
  record_uuid: string; // Idempotent client-generated UUID (indexed)
  org_id: string; // Plaintext routing & AAD construction (indexed)
  user_id: string; // Plaintext routing & AAD construction (indexed)
  sync_status: SyncStatus; // Queue state (indexed)
  created_at: string; // Creation timestamp (indexed)

  // Encrypted sensitive fields (AES-256-GCM)
  iv: string; // Base64 12-byte random IV
  ciphertext: string; // Base64 encrypted { action_type, timestamp, device_hash }
  crypto_version: number; // Encryption schema version (1)

  // Local-only diagnostic/recovery field (never submitted to backend API)
  decrypt_error?: string;

  // Legacy plaintext fields kept optional for safe v1 migration fallback
  action_type?: 'CLOCK_IN' | 'CLOCK_OUT';
  timestamp?: string;
  device_hash?: string;
}

export class AttendanceDatabase extends Dexie {
  offlineLogs!: Table<OfflineAttendanceLog>;

  constructor() {
    super('HabeshanREMS_Attendance');

    // Version 1: Legacy schema with plaintext sensitive fields
    this.version(1).stores({
      offlineLogs: '++id, record_uuid, sync_status, action_type, timestamp',
    });

    // Version 2: Hybrid schema with plaintext routing & AES-GCM encrypted payload
    this.version(2)
      .stores({
        offlineLogs: '++id, record_uuid, user_id, org_id, sync_status, created_at',
      })
      .upgrade(async (tx) => {
        try {
          const key = await getOrCreateDeviceKey();
          const table = tx.table<OfflineAttendanceLog>('offlineLogs');
          const records = await table.toArray();

          for (const record of records) {
            // Check if record is legacy plaintext without ciphertext
            if (!record.ciphertext && record.action_type && record.timestamp && record.device_hash) {
              if (record.record_uuid && record.org_id && record.user_id) {
                try {
                  const encrypted = await encryptAttendancePayload(
                    record.record_uuid,
                    record.org_id,
                    record.user_id,
                    {
                      action_type: record.action_type,
                      timestamp: record.timestamp,
                      device_hash: record.device_hash,
                    },
                    key
                  );

                  record.iv = encrypted.iv;
                  record.ciphertext = encrypted.ciphertext;
                  record.crypto_version = encrypted.crypto_version;

                  // Clear sensitive plaintext fields after encryption
                  delete record.action_type;
                  delete record.timestamp;
                  delete record.device_hash;

                  if (record.id) {
                    await table.put(record);
                  }
                } catch (encErr) {
                  console.error('[DB Upgrade] Failed to encrypt legacy row:', encErr);
                  record.decrypt_error = 'LEGACY_MIGRATION_ENCRYPTION_FAILED';
                  if (record.id) {
                    await table.put(record);
                  }
                }
              } else {
                // Incomplete routing metadata — preserve row in legacy recovery state
                record.decrypt_error = 'LEGACY_MIGRATION_MISSING_METADATA';
                if (record.id) {
                  await table.put(record);
                }
              }
            }
          }
        } catch (upgradeErr) {
          console.error('[DB Upgrade] Error during version 2 migration:', upgradeErr);
          // Never throw destructive errors that wipe database or drop tables
        }
      });
  }
}

export const db = new AttendanceDatabase();