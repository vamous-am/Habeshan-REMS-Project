import { db } from './db';

/**
 * Task 13 (Dev 2): Export pending or failed offline logs to a local JSON file.
 * Fulfills NFR-REL-06: Manual export/backup option for queued offline records.
 */
export async function exportOfflineQueueToJSON(): Promise<void> {
  try {
    const logs = await db.offlineLogs.toArray();
    if (!logs || logs.length === 0) {
      alert('No offline attendance records to export.');
      return;
    }

    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(logs, null, 2));
    const downloadAnchor = document.createElement('a');
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    
    downloadAnchor.setAttribute('href', dataStr);
    downloadAnchor.setAttribute('download', `offline_attendance_backup_${timestamp}.json`);
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
  } catch (err) {
    console.error('Failed to export offline queue backup:', err);
    alert('Error generating backup file for offline queue.');
  }
}