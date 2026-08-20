/**
 * components/TaskAssignmentPanel.tsx
 *
 * UI for assigning/unassigning users on a task detail page (FR-TASK-02).
 *
 * Backend restrictions:
 *   - Only admin / manager can mutate assignments; backend returns 403 otherwise.
 *   - Assign accepts bulk user_ids (UUID strings).
 *   - Unassign removes one user at a time via DELETE /assignments/:userID.
 *
 * The backend does not (yet) provide a user directory endpoint, so users are
 * input by UUID string directly.  This is acceptable for manual testing and
 * matches the seed-data UUIDs documented in the seed migration.
 */

import { useState } from "react";
import type { Task } from "../types";
import { extractApiErrorMessage, assignTask, unassignTask } from "../api";

interface Props {
  task: Task;
  onChanged: () => void | Promise<void>;
}

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export default function TaskAssignmentPanel({ task, onChanged }: Props) {
  const assigned = task.assigned_to ?? [];
  const [newUserID, setNewUserID] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<Record<string, string>>({});
  const [assigning, setAssigning] = useState(false);
  const [confirmUnassign, setConfirmUnassign] = useState<string | null>(null);

  async function handleAssign(e: React.FormEvent) {
    e.preventDefault();
    const uid = newUserID.trim();
    if (!uid) return;
    if (!UUID_RE.test(uid)) {
      setError("User ID must be a valid UUID (e.g. 33333333-3333-3333-3333-333333333333).");
      return;
    }
    if (assigned.includes(uid)) {
      setError("User is already assigned to this task.");
      return;
    }
    setAssigning(true);
    setError(null);
    try {
      await assignTask(task.id, { user_ids: [uid] });
      setNewUserID("");
      await onChanged();
    } catch (err: unknown) {
      setError(extractApiErrorMessage(err));
    } finally {
      setAssigning(false);
    }
  }

  async function handleUnassign(userID: string) {
    setActionError((prev) => ({ ...prev, [userID]: "" }));
    try {
      await unassignTask(task.id, userID);
      setConfirmUnassign(null);
      await onChanged();
    } catch (err: unknown) {
      setActionError((prev) => ({
        ...prev,
        [userID]: extractApiErrorMessage(err),
      }));
    }
  }

  return (
    <div
      style={{
        border: "1px solid #e5e7eb",
        borderRadius: 8,
        padding: 16,
        marginBottom: 20,
        background: "#fff",
      }}
    >
      <h3 style={{ margin: "0 0 12px", fontSize: 14, fontWeight: 600 }}>
        Assignees
        {assigned.length > 0 && (
          <span
            style={{
              marginLeft: 8,
              padding: "1px 8px",
              borderRadius: 10,
              background: "#e5e7eb",
              color: "#374151",
              fontSize: 11,
              fontWeight: 500,
            }}
          >
            {assigned.length}
          </span>
        )}
      </h3>

      {/* Existing assignees list */}
      {assigned.length === 0 ? (
        <p style={{ fontSize: 13, color: "#6b7280", margin: "0 0 12px" }}>
          No users are assigned. Assign a user by UUID below.
        </p>
      ) : (
        <ul
          style={{
            listStyle: "none",
            padding: 0,
            margin: "0 0 12px",
            display: "flex",
            flexDirection: "column",
            gap: 6,
          }}
        >
          {assigned.map((uid) => (
            <li
              key={uid}
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                padding: "8px 12px",
                background: "#f9fafb",
                borderRadius: 6,
                border: "1px solid #f3f4f6",
              }}
            >
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  gap: 2,
                  fontFamily:
                    'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
                  fontSize: 12,
                  color: "#111827",
                }}
              >
                <span>{uid}</span>
                {actionError[uid] && (
                  <span style={{ color: "#dc2626", fontSize: 11 }}>
                    {actionError[uid]}
                  </span>
                )}
              </div>

              {confirmUnassign === uid ? (
                <div style={{ display: "flex", gap: 6 }}>
                  <button
                    onClick={() => handleUnassign(uid)}
                    style={{
                      padding: "4px 10px",
                      fontSize: 12,
                      background: "#dc2626",
                      color: "#fff",
                      border: "none",
                      borderRadius: 4,
                      cursor: "pointer",
                    }}
                  >
                    Confirm
                  </button>
                  <button
                    onClick={() => setConfirmUnassign(null)}
                    style={{
                      padding: "4px 10px",
                      fontSize: 12,
                      background: "#fff",
                      color: "#374151",
                      border: "1px solid #d1d5db",
                      borderRadius: 4,
                      cursor: "pointer",
                    }}
                  >
                    Cancel
                  </button>
                </div>
              ) : (
                <button
                  onClick={() => setConfirmUnassign(uid)}
                  style={{
                    padding: "4px 10px",
                    fontSize: 12,
                    background: "#fff",
                    color: "#991b1b",
                    border: "1px solid #fecaca",
                    borderRadius: 4,
                    cursor: "pointer",
                  }}
                >
                  Unassign
                </button>
              )}
            </li>
          ))}
        </ul>
      )}

      {/* Add new assignee */}
      <form onSubmit={handleAssign} style={{ display: "flex", gap: 8 }}>
        <input
          value={newUserID}
          onChange={(e) => {
            setNewUserID(e.target.value);
            if (error) setError(null);
          }}
          placeholder="Paste user UUID (e.g. 33333333-3333-3333-3333-333333333333)"
          style={{
            flex: 1,
            padding: "6px 10px",
            fontSize: 13,
            border: "1px solid #d1d5db",
            borderRadius: 4,
            fontFamily:
              'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
          }}
        />
        <button
          type="submit"
          disabled={assigning || !newUserID.trim()}
          style={{
            padding: "6px 14px",
            fontSize: 13,
            background: "#2563eb",
            color: "#fff",
            border: "none",
            borderRadius: 4,
            cursor: assigning ? "progress" : "pointer",
            opacity: assigning || !newUserID.trim() ? 0.6 : 1,
          }}
        >
          {assigning ? "Adding…" : "+ Assign"}
        </button>
      </form>

      {error && (
        <p style={{ color: "#dc2626", fontSize: 12, margin: "8px 0 0" }}>
          {error}
        </p>
      )}

      <p style={{ fontSize: 11, color: "#9ca3af", margin: "8px 0 0" }}>
        Tip: use the seed UUIDs — Alice: 3333…3333, Bob: 4444…4444, Manager: 2222…2222, Admin: 1111…1111.
      </p>
    </div>
  );
}
