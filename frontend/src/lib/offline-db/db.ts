import Dexie, { Table } from 'dexie';

export interface OfflineAttendanceLog {
  id?: number;
  record_uuid: string; // Idempotent client-generated UUID
  org_id: string;
  user_id: string;
  action_type: 'CLOCK_IN' | 'CLOCK_OUT';
  timestamp: string;  // ISO string timestamp
  sync_status: 'OFFLINE_LOGGED' | 'PENDING_SYNC' | 'SYNCED_VERIFIED' | 'REJECTED_TAMPERED';
  device_hash: string; // Cryptographic SHA-256 hash
  created_at: string;
}

export class AttendanceDatabase extends Dexie {
  offlineLogs!: Table<OfflineAttendanceLog>;

  constructor() {
    super('HabeshanREMS_Attendance');
    this.version(1).stores({
      // Primary Key ++id, Indexes on record_uuid and sync_status for sync workers
      offlineLogs: '++id, record_uuid, sync_status, action_type, timestamp'
    });
  }
}

export const db = new AttendanceDatabase();