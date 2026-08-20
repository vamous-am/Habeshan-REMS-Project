/**
 * features/tasks/useTimer.ts
 *
 * Hook for task timer operations.
 *
 * Responsibilities:
 *   - Start / pause / resume / stop via the backend API
 *   - Live-updating elapsed / total time display based on timer history
 *   - Sync initial timer state from the backend history (FR-TASK-08 audit log)
 *   - Generate idempotency keys (record_uuid) and device_hash per event so
 *     the backend duplicate-submission guard (FR-TASK-10) always gets a
 *     unique key per submission.
 */

import { useEffect, useMemo, useRef, useState } from "react";
import { v4 as uuidv4 } from "uuid";
import type { PauseReason, TimeLog, TimerState } from "./types";
import {
  extractApiErrorMessage,
  fetchTimerHistory,
  pauseTimer,
  resumeTimer,
  startTimer,
  stopTimer,
} from "./api";

// Stable device fingerprint — recomputed once per page load, not per event.
// Mirrors the attendance feature (btoa of userAgent, capped to 128 chars).
const DEVICE_HASH =
  typeof navigator !== "undefined"
    ? btoa(navigator.userAgent || "unknown-device").slice(0, 128)
    : "server-rendered-device";

// Polling interval for live elapsed-time ticks (seconds).
const TICK_MS = 1000;

// ─── time helpers ─────────────────────────────────────────────────────────────

function nowISO(): string {
  return new Date().toISOString();
}

export function formatDuration(totalMinutes: number): string {
  if (totalMinutes < 1) return "< 1 min";
  const h = Math.floor(totalMinutes / 60);
  const m = Math.round(totalMinutes % 60);
  if (h === 0) return `${m} min`;
  if (m === 0) return `${h} h`;
  return `${h}h ${m}m`;
}

export function formatElapsedSeconds(elapsedSec: number): string {
  if (elapsedSec < 0) elapsedSec = 0;
  const h = Math.floor(elapsedSec / 3600);
  const m = Math.floor((elapsedSec % 3600) / 60);
  const s = Math.floor(elapsedSec % 60);
  const pad = (n: number) => n.toString().padStart(2, "0");
  return `${pad(h)}:${pad(m)}:${pad(s)}`;
}

/**
 * Sums completed-duration minutes from a list of time logs.
 * Ignores open (running) segments.
 */
export function totalLoggedMinutes(logs: TimeLog[]): number {
  let total = 0;
  for (const l of logs) {
    if (l.duration_minutes != null) {
      total += l.duration_minutes;
    }
  }
  return total;
}

/**
 * Given a list of logs, returns the currently OPEN (running) segment for the
 * current user, or null if none.
 *
 * Open = stopped_at is null/undefined.
 *
 * (The backend enforces at most one open segment per (task, user); we take
 * the most-recent-by-started_at just in case.)
 */
function findActiveLog(
  logs: TimeLog[],
  currentUserID: string | null
): TimeLog | null {
  let active: TimeLog | null = null;
  for (const l of logs) {
    if (l.stopped_at) continue;
    if (currentUserID && l.user_id !== currentUserID) continue;
    if (!active || new Date(l.started_at) > new Date(active.started_at)) {
      active = l;
    }
  }
  return active;
}

/**
 * Determines whether the most recent event was a PAUSE (segment closed with
 * a pause_reason).  Used to seed timerState = "paused" so the UI matches
 * the backend's understanding before the first manual action on this page.
 */
function lastEventWasPause(logs: TimeLog[]): boolean {
  if (logs.length === 0) return false;
  const sorted = [...logs].sort(
    (a, b) => new Date(a.started_at).getTime() - new Date(b.started_at).getTime()
  );
  const last = sorted[sorted.length - 1];
  return !!last.stopped_at && !!last.pause_reason;
}

// ─── hook ─────────────────────────────────────────────────────────────────────

export function useTimer(
  taskID: string,
  options: { initialLogs?: TimeLog[]; onStopped?: () => void } = {}
) {
  const { initialLogs = [], onStopped } = options;

  // Identity — read from the same localStorage stub that api/client.ts uses.
  // We need user id to attribute open segments to "me" vs other assignees.
  const currentUserID: string | null =
    typeof localStorage !== "undefined"
      ? localStorage.getItem("x-user-id")
      : null;

  // ── state ────────────────────────────────────────────────────────────────
  const [timerState, setTimerState] = useState<TimerState>(() => {
    if (findActiveLog(initialLogs, currentUserID)) return "running";
    if (lastEventWasPause(initialLogs)) return "paused";
    return "idle";
  });
  const [logs, setLogs] = useState<TimeLog[]>(initialLogs);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // `nowMs` is a state value updated on every tick (once per second while a
  // segment is open).  Keeping it as real state (instead of a ref) lets us
  // read it during render without violating react-hooks purity/refs rules.
  const [nowMs, setNowMs] = useState<number>(() => new Date().getTime());
  const tickIntervalRef = useRef<number | null>(null);

  // ── derived: running segment + elapsed seconds ────────────────────────────
  const activeLog = useMemo(
    () => findActiveLog(logs, currentUserID),
    [logs, currentUserID]
  );

  const elapsedSeconds = useMemo(() => {
    if (!activeLog) return 0;
    const start = new Date(activeLog.started_at).getTime();
    return Math.max(0, Math.floor((nowMs - start) / 1000));
  }, [activeLog, nowMs]);

  const completedMinutes = useMemo(() => totalLoggedMinutes(logs), [logs]);

  // ── ticking: start/stop an interval based on whether a segment is open ───
  useEffect(() => {
    if (activeLog) {
      if (tickIntervalRef.current == null) {
        tickIntervalRef.current = window.setInterval(() => {
          setNowMs(new Date().getTime());
        }, TICK_MS);
      }
    } else if (tickIntervalRef.current != null) {
      clearInterval(tickIntervalRef.current);
      tickIntervalRef.current = null;
    }
    return () => {
      if (tickIntervalRef.current != null) {
        clearInterval(tickIntervalRef.current);
        tickIntervalRef.current = null;
      }
    };
  }, [activeLog]);

  // ── external refresh ──────────────────────────────────────────────────────
  async function refreshLogs() {
    if (!taskID) return;
    try {
      const fresh = await fetchTimerHistory(taskID);
      setLogs(fresh ?? []);
      if (findActiveLog(fresh, currentUserID)) {
        setTimerState("running");
      } else if (lastEventWasPause(fresh)) {
        setTimerState("paused");
      } else {
        setTimerState("idle");
      }
    } catch {
      // history refresh is best-effort; ignore errors here.
    }
  }

  // ── actions ───────────────────────────────────────────────────────────────
  async function handleStart() {
    setLoading(true);
    setError(null);
    try {
      const log = await startTimer(taskID, {
        started_at: nowISO(),
        device_hash: DEVICE_HASH,
        record_uuid: uuidv4(),
        sync_status: "pending_sync",
      });
      setLogs((prev) => [...prev, log]);
      setTimerState("running");
    } catch (e: unknown) {
      setError(extractApiErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }

  async function handlePause(reason: PauseReason) {
    setLoading(true);
    setError(null);
    try {
      const log = await pauseTimer(taskID, {
        paused_at: nowISO(),
        pause_reason: reason,
        device_hash: DEVICE_HASH,
        record_uuid: uuidv4(),
        sync_status: "pending_sync",
      });
      // The backend closes the currently active segment; refresh to pick up
      // the updated duration + pause reason.
      setLogs((prev) =>
        prev.map((l) => (l.id === log.id ? log : l))
      );
      setTimerState("paused");
    } catch (e: unknown) {
      setError(extractApiErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }

  async function handleResume() {
    setLoading(true);
    setError(null);
    try {
      const log = await resumeTimer(taskID, {
        resumed_at: nowISO(),
        device_hash: DEVICE_HASH,
        record_uuid: uuidv4(),
        sync_status: "pending_sync",
      });
      setLogs((prev) => [...prev, log]);
      setTimerState("running");
    } catch (e: unknown) {
      setError(extractApiErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }

  async function handleStop() {
    setLoading(true);
    setError(null);
    try {
      const log = await stopTimer(taskID, {
        stopped_at: nowISO(),
        device_hash: DEVICE_HASH,
        record_uuid: uuidv4(),
        sync_status: "pending_sync",
      });
      setLogs((prev) =>
        prev.map((l) => (l.id === log.id ? log : l))
      );
      setTimerState("idle");
      onStopped?.();
    } catch (e: unknown) {
      setError(extractApiErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }

  return {
    timerState,
    loading,
    error,
    logs,
    activeLog,
    elapsedSeconds,
    completedMinutes,
    handleStart,
    handlePause,
    handleResume,
    handleStop,
    refreshLogs,
    setLogs,
    currentUserID,
  };
}
