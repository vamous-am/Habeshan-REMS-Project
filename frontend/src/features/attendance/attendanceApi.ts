import { getAuthToken, getCurrentUserId } from '../../lib/api/authClient';
import { generateDeviceHash } from '../../lib/offline-db/crypto';

export interface AttendanceRecord {
  id?: string;
  record_uuid: string;
  user_id: string;
  clock_in: string;
  clock_out?: string;
  total_hours?: number;
  sync_status: 'OFFLINE_LOGGED' | 'SYNCED_VERIFIED' | 'REJECTED_TAMPERED';
  device_hash?: string;
}

const API_BASE = '/api/v1/attendance';

function getAuthHeader(): Record<string, string> {
  const token =
    getAuthToken() ||
    localStorage.getItem('token') ||
    localStorage.getItem('jwt_token') ||
    localStorage.getItem('access_token') ||
    '';
  return token ? { Authorization: `Bearer ${token}` } : {};
}

export async function clockInApi(): Promise<AttendanceRecord> {
  const recordUUID = crypto.randomUUID();
  const userId = getCurrentUserId() || '';
  const nowISO = new Date().toISOString();
  const deviceHash = await generateDeviceHash(recordUUID, userId, 'CLOCK_IN', nowISO);

  const res = await fetch(`${API_BASE}/clock-in`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeader(),
    },
    body: JSON.stringify({
      record_uuid: recordUUID,
      device_hash: deviceHash,
      timestamp: nowISO,
    }),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.message || 'Failed to clock in');
  }

  const responseJson = await res.json();
  return responseJson.data ?? responseJson;
}

export async function clockOutApi(): Promise<AttendanceRecord> {
  const recordUUID = crypto.randomUUID();
  const userId = getCurrentUserId() || '';
  const nowISO = new Date().toISOString();
  const deviceHash = await generateDeviceHash(recordUUID, userId, 'CLOCK_OUT', nowISO);

  const res = await fetch(`${API_BASE}/clock-out`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeader(),
    },
    body: JSON.stringify({
      record_uuid: recordUUID,
      device_hash: deviceHash,
      timestamp: nowISO,
    }),
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.message || 'Failed to clock out');
  }

  const responseJson = await res.json();
  return responseJson.data ?? responseJson;
}

/**
 * Task 13: Triggers attendance report CSV export & download
 */
export async function exportAttendanceCsvApi(
  startDate?: string,
  endDate?: string
): Promise<void> {
  const params = new URLSearchParams();
  if (startDate) params.append('start_date', startDate);
  if (endDate) params.append('end_date', endDate);

  const res = await fetch(`${API_BASE}/export?${params.toString()}`, {
    method: 'GET',
    headers: {
      ...getAuthHeader(),
    },
  });

  if (!res.ok) {
    throw new Error('Failed to export attendance CSV report');
  }

  const blob = await res.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `attendance_report_${new Date().toISOString().slice(0, 10)}.csv`;
  document.body.appendChild(a);
  a.click();
  a.remove();
  window.URL.revokeObjectURL(url);
}