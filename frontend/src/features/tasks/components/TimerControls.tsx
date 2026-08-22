/**
 * components/TimerControls.tsx
 *
 * Start / Pause / Resume / Stop controls for a task timer.
 * Pause requires the user to select a reason (FR-TASK-06).
 * Shows:
 *   - Live-updating current elapsed time (HH:MM:SS) when running
 *   - Completed total logged time
 *   - Error messages
 *   - Load state
 */

import { useState } from "react";
import type { PauseReason, TimerState } from "../types";
import { PAUSE_REASONS } from "../types";
import { formatDuration, formatElapsedSeconds } from "../useTimer";

interface Props {
  taskID: string;
  timerState: TimerState;
  loading: boolean;
  error: string | null;
  elapsedSeconds: number;
  completedMinutes: number;
  onStart: () => void;
  onPause: (reason: PauseReason) => void;
  onResume: () => void;
  onStop: () => void;
}

export default function TimerControls(props: Props) {
  const {
    timerState,
    loading,
    error,
    elapsedSeconds,
    completedMinutes,
    onStart,
    onPause,
    onResume,
    onStop,
  } = props;
  const [showPauseModal, setShowPauseModal] = useState(false);
  const [selectedReason, setSelectedReason] = useState<PauseReason>(
    PAUSE_REASONS[0]
  );
  const [confirmStop, setConfirmStop] = useState(false);

  function handlePauseConfirm() {
    setShowPauseModal(false);
    onPause(selectedReason);
  }

  function handleStopConfirm() {
    setConfirmStop(false);
    onStop();
  }

  return (
    <div style={{ marginTop: 8 }}>
      {/* Time display */}
      <div
        style={{
          display: "flex",
          gap: 24,
          alignItems: "center",
          marginBottom: 12,
          padding: "12px 16px",
          background: "#f9fafb",
          border: "1px solid #e5e7eb",
          borderRadius: 8,
        }}
      >
        <div>
          <div style={{ fontSize: 11, color: "#6b7280", marginBottom: 2 }}>
            {timerState === "running"
              ? "Current session"
              : timerState === "paused"
                ? "Paused at"
                : "Current session"}
          </div>
          <div
            style={{
              fontSize: 28,
              fontWeight: 700,
              color: timerState === "running" ? "#047857" : "#374151",
              fontFamily:
                'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
              fontVariantNumeric: "tabular-nums",
            }}
          >
            {timerState === "idle"
              ? "—"
              : formatElapsedSeconds(
                  timerState === "paused" ? Math.max(0, completedMinutes * 60) : elapsedSeconds
                )}
          </div>
        </div>
        <div style={{ borderLeft: "1px solid #e5e7eb", height: 40 }} />
        <div>
          <div style={{ fontSize: 11, color: "#6b7280", marginBottom: 2 }}>
            Total logged
          </div>
          <div
            style={{
              fontSize: 20,
              fontWeight: 600,
              color: "#111827",
            }}
          >
            {formatDuration(
              completedMinutes +
                (timerState === "running"
                  ? Math.floor(elapsedSeconds / 60)
                  : 0)
            )}
          </div>
        </div>
        <div style={{ flex: 1 }} />
        <div>
          <TimerStateChip state={timerState} />
        </div>
      </div>

      {error && (
        <p
          style={{
            color: "#dc2626",
            fontSize: 13,
            marginBottom: 8,
            padding: "8px 12px",
            background: "#fef2f2",
            borderRadius: 4,
            border: "1px solid #fecaca",
          }}
        >
          {error}
        </p>
      )}

      {/* Buttons */}
      <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
        {timerState === "idle" && (
          <PrimaryButton onClick={onStart} disabled={loading} color="#047857">
            ▶ Start Timer
          </PrimaryButton>
        )}

        {timerState === "running" && (
          <>
            <PrimaryButton
              onClick={() => setShowPauseModal(true)}
              disabled={loading}
              color="#d97706"
            >
              ⏸ Pause
            </PrimaryButton>
            {confirmStop ? (
              <>
                <SecondaryButton onClick={handleStopConfirm} disabled={loading}>
                  ✓ Confirm Stop
                </SecondaryButton>
                <SecondaryButton
                  onClick={() => setConfirmStop(false)}
                  disabled={loading}
                >
                  Cancel
                </SecondaryButton>
              </>
            ) : (
              <SecondaryButton
                onClick={() => setConfirmStop(true)}
                disabled={loading}
              >
                ⏹ Stop
              </SecondaryButton>
            )}
          </>
        )}

        {timerState === "paused" && (
          <>
            <PrimaryButton onClick={onResume} disabled={loading} color="#047857">
              ▶ Resume
            </PrimaryButton>
            {confirmStop ? (
              <>
                <SecondaryButton onClick={handleStopConfirm} disabled={loading}>
                  ✓ Confirm Stop
                </SecondaryButton>
                <SecondaryButton
                  onClick={() => setConfirmStop(false)}
                  disabled={loading}
                >
                  Cancel
                </SecondaryButton>
              </>
            ) : (
              <SecondaryButton
                onClick={() => setConfirmStop(true)}
                disabled={loading}
              >
                ⏹ Stop
              </SecondaryButton>
            )}
          </>
        )}

        {loading && (
          <span style={{ fontSize: 13, color: "#6b7280", alignSelf: "center" }}>
            Working…
          </span>
        )}
      </div>

      {/* Pause reason modal */}
      {showPauseModal && (
        <div
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0,0,0,0.4)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 100,
          }}
          role="dialog"
          aria-modal="true"
        >
          <div
            style={{
              background: "#fff",
              padding: 24,
              borderRadius: 8,
              minWidth: 340,
              maxWidth: "90vw",
              boxShadow: "0 12px 40px rgba(0,0,0,0.15)",
            }}
          >
            <h3 style={{ margin: "0 0 8px", fontSize: 16 }}>
              Pause timer — reason required
            </h3>
            <p style={{ margin: "0 0 16px", fontSize: 13, color: "#6b7280" }}>
              Per FR-TASK-06, every pause must be attributed to a reason.
            </p>
            <label style={{ display: "flex", flexDirection: "column", gap: 4 }}>
              <span style={{ fontSize: 13, fontWeight: 500 }}>
                Why are you pausing?
              </span>
              <select
                value={selectedReason}
                onChange={(e) =>
                  setSelectedReason(e.target.value as PauseReason)
                }
                style={{
                  width: "100%",
                  marginBottom: 20,
                  padding: "8px 10px",
                  fontSize: 14,
                  border: "1px solid #d1d5db",
                  borderRadius: 4,
                  background: "#fff",
                }}
              >
                {PAUSE_REASONS.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </label>
            <div
              style={{
                display: "flex",
                gap: 8,
                justifyContent: "flex-end",
              }}
            >
              <SecondaryButton onClick={() => setShowPauseModal(false)}>
                Cancel
              </SecondaryButton>
              <PrimaryButton onClick={handlePauseConfirm} color="#d97706">
                Confirm Pause
              </PrimaryButton>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ─── small helper components ─────────────────────────────────────────────────

interface BtnProps {
  children: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
  color?: string;
}

function PrimaryButton({ children, onClick, disabled, color = "#2563eb" }: BtnProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      style={{
        padding: "8px 14px",
        fontSize: 13,
        fontWeight: 500,
        background: disabled ? "#9ca3af" : color,
        color: "#fff",
        border: "none",
        borderRadius: 6,
        cursor: disabled ? "not-allowed" : "pointer",
        opacity: disabled ? 0.7 : 1,
      }}
    >
      {children}
    </button>
  );
}

function SecondaryButton({ children, onClick, disabled }: BtnProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      style={{
        padding: "8px 14px",
        fontSize: 13,
        fontWeight: 500,
        background: disabled ? "#f3f4f6" : "#fff",
        color: "#374151",
        border: "1px solid #d1d5db",
        borderRadius: 6,
        cursor: disabled ? "not-allowed" : "pointer",
        opacity: disabled ? 0.7 : 1,
      }}
    >
      {children}
    </button>
  );
}

function TimerStateChip({ state }: { state: TimerState }) {
  const map: Record<TimerState, { label: string; bg: string; fg: string }> = {
    idle:    { label: "Idle",     bg: "#f3f4f6", fg: "#6b7280" },
    running: { label: "Running",  bg: "#dcfce7", fg: "#166534" },
    paused:  { label: "Paused",   bg: "#fef9c3", fg: "#854d0e" },
  };
  const m = map[state];
  return (
    <span
      style={{
        padding: "4px 10px",
        borderRadius: 12,
        fontSize: 12,
        fontWeight: 600,
        background: m.bg,
        color: m.fg,
      }}
    >
      {m.label}
    </span>
  );
}
