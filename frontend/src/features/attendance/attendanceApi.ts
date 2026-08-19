const API_BASE = 'http://localhost:8080/api/v1/attendance';

export interface AttendanceRecord {
  id: string;
  user_id: string;
  clock_in: string;
  clock_out?: string;
  total_hours?: number;
  sync_status: string;
  device_hash: string;
  record_uuid: string;
}

// Generates a browser device fingerprint SHA-256 hash
export async function getDeviceHash(): Promise<string> {
  const userAgent = navigator.userAgent;
  const screenRes = `${window.screen.width}x${window.screen.height}`;
  const rawString = `${userAgent}-${screenRes}`;
  const encoder = new TextEncoder();
  const data = encoder.encode(rawString);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}

export async function clockInApi(): Promise<AttendanceRecord> {
  const deviceHash = await getDeviceHash();
  const recordUuid = typeof crypto.randomUUID === 'function' 
    ? crypto.randomUUID() 
    : 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
        const r = (Math.random() * 16) | 0;
        const v = c === 'x' ? r : (r & 0x3) | 0x8;
        return v.toString(16);
      });

  const response = await fetch(`${API_BASE}/clock-in`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      device_hash: deviceHash,
      record_uuid: recordUuid,
      timestamp: new Date().toISOString(),
    }),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.message || 'Failed to record clock-in');
  }
  return data.data;
}

export async function clockOutApi(): Promise<AttendanceRecord> {
  const deviceHash = await getDeviceHash();

  const response = await fetch(`${API_BASE}/clock-out`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      device_hash: deviceHash,
      timestamp: new Date().toISOString(),
    }),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.message || 'Failed to record clock-out');
  }
  return data.data;
}