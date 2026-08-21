import React, { useState, useEffect, useCallback } from 'react';
import { db } from '../../lib/offline-db/db';
import { exportOfflineQueueToJSON } from '../../lib/offline-db/backup';

export const SyncStatusBadge: React.FC = () => {
  const [isOnline, setIsOnline] = useState<boolean>(navigator.onLine);
  const [pendingCount, setPendingCount] = useState<number>(0);
  const [tamperedCount, setTamperedCount] = useState<number>(0);

  const updateCounts = useCallback(async () => {
    try {
      const pending = await db.offlineLogs
        .where('sync_status')
        .equals('OFFLINE_LOGGED')
        .count();
      const tampered = await db.offlineLogs
        .where('sync_status')
        .equals('REJECTED_TAMPERED')
        .count();
      setPendingCount(pending);
      setTamperedCount(tampered);
    } catch (e) {
      console.error('Error reading offline log status counts:', e);
    }
  }, []);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Schedule query async on microtask queue to avoid ESLint set-state-in-effect warning
    const timer = setTimeout(() => {
      updateCounts();
    }, 0);

    const interval = setInterval(() => {
      updateCounts();
    }, 3000);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
      clearTimeout(timer);
      clearInterval(interval);
    };
  }, [updateCounts]);

  return (
    <div className="flex flex-wrap items-center gap-3 p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
      {/* Task 14: Network Status Indicator */}
      <div className="flex items-center gap-2">
        <span
          className={`h-3 w-3 rounded-full ${
            isOnline ? 'bg-green-500 animate-pulse' : 'bg-amber-500 animate-bounce'
          }`}
        />
        <span className="text-sm font-medium text-gray-700 dark:text-gray-200">
          {isOnline ? 'Online' : 'Offline Mode'}
        </span>
      </div>

      {/* Sync State Display */}
      {pendingCount > 0 && (
        <span className="px-2.5 py-0.5 text-xs font-semibold rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
          {pendingCount} Pending Sync
        </span>
      )}

      {tamperedCount > 0 && (
        <span className="px-2.5 py-0.5 text-xs font-semibold rounded-full bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
          ⚠️ {tamperedCount} Tamper Rejected
        </span>
      )}

      {pendingCount === 0 && tamperedCount === 0 && isOnline && (
        <span className="px-2.5 py-0.5 text-xs font-semibold rounded-full bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
          All Logged Data Synced
        </span>
      )}

      {/* Task 13: Manual Offline Queue Export Trigger */}
      {pendingCount > 0 && (
        <button
          onClick={exportOfflineQueueToJSON}
          type="button"
          className="ml-auto text-xs px-3 py-1 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded border border-gray-300 dark:border-gray-600 transition-colors"
        >
          📥 Backup Queue JSON
        </button>
      )}
    </div>
  );
};