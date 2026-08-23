/**
 * components/TaskAssignmentPanel.tsx
 *
 * UI for assigning/unassigning users on a task detail page (FR-TASK-02).
 * Supports assigning by selecting team members (Name / Email / Role) or typing an email/name.
 */

import { useState, useEffect } from "react";
import {
  UserPlus,
  UserX,
  Mail,
  Search,
  Check,
  AlertCircle,
  Loader2,
} from "lucide-react";
import type { Task, AssignableUser, AssignedUser } from "../types";
import {
  extractApiErrorMessage,
  assignTask,
  unassignTask,
  fetchAssignableUsers,
} from "../api";
import { getUserRole } from "../../../lib/api/authClient";

interface Props {
  task: Task;
  onChanged: () => void | Promise<void>;
}

export default function TaskAssignmentPanel({ task, onChanged }: Props) {
  const role = getUserRole();
  const canManageAssignments = role === "admin" || role === "manager";

  const [availableUsers, setAvailableUsers] = useState<AssignableUser[]>([]);
  const [loadingUsers, setLoadingUsers] = useState(false);
  const [selectedUserEmail, setSelectedUserEmail] = useState("");
  const [customInput, setCustomInput] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<Record<string, string>>({});
  const [assigning, setAssigning] = useState(false);
  const [confirmUnassign, setConfirmUnassign] = useState<string | null>(null);

  // Load assignable team members from org
  useEffect(() => {
    let active = true;
    if (canManageAssignments) {
      setLoadingUsers(true);
      fetchAssignableUsers()
        .then((users) => {
          if (active) setAvailableUsers(users);
        })
        .catch(() => {
          if (active) setAvailableUsers([]);
        })
        .finally(() => {
          if (active) setLoadingUsers(false);
        });
    }
    return () => {
      active = false;
    };
  }, [canManageAssignments]);

  // Combine assigned_users or match assigned_to IDs with availableUsers
  const currentAssigned: AssignedUser[] = (task.assigned_users && task.assigned_users.length > 0)
    ? task.assigned_users
    : (task.assigned_to ?? []).map((uid) => {
        const found = availableUsers.find((u) => u.id === uid);
        return {
          id: uid,
          full_name: found ? found.full_name : "Assigned Member",
          email: found ? found.email : "",
          role: found ? found.role : "employee",
        };
      });

  const assignedIdSet = new Set(task.assigned_to ?? []);
  const unassignedMembers = availableUsers.filter((u) => !assignedIdSet.has(u.id));

  async function handleAssignUser(userEmailOrIdent: string) {
    const target = userEmailOrIdent.trim();
    if (!target) return;

    setAssigning(true);
    setError(null);
    try {
      await assignTask(task.id, {
        identifiers: [target],
      });
      setSelectedUserEmail("");
      setCustomInput("");
      await onChanged();
    } catch (err: unknown) {
      setError(extractApiErrorMessage(err));
    } finally {
      setAssigning(false);
    }
  }

  async function handleUnassign(identifier: string) {
    setActionError((prev) => ({ ...prev, [identifier]: "" }));
    try {
      await unassignTask(task.id, identifier);
      setConfirmUnassign(null);
      await onChanged();
    } catch (err: unknown) {
      setActionError((prev) => ({
        ...prev,
        [identifier]: extractApiErrorMessage(err),
      }));
    }
  }

  function getInitials(name: string, email: string): string {
    if (name && name !== "Assigned Member") {
      const parts = name.trim().split(" ");
      if (parts.length >= 2) {
        return (parts[0][0] + parts[1][0]).toUpperCase();
      }
      return name.slice(0, 2).toUpperCase();
    }
    if (email) {
      return email.slice(0, 2).toUpperCase();
    }
    return "US";
  }

  return (
    <div className="rounded-lg border border-ink/10 bg-paper p-5 shadow-xs transition-all">
      {/* Header */}
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="font-display text-base font-bold text-ink">
            Assigned Team Members
          </h3>
          <span className="rounded-full bg-ink/10 px-2 py-0.5 font-mono text-xs font-semibold text-ink">
            {currentAssigned.length}
          </span>
        </div>
        {canManageAssignments && (
          <span className="text-xs text-ink500">
            Select or search members by name/email
          </span>
        )}
      </div>

      {/* Current Assignees List */}
      {currentAssigned.length === 0 ? (
        <div className="mb-5 rounded-md border border-dashed border-ink/20 bg-paper-dim/60 p-4 text-center">
          <p className="text-xs text-ink500">
            No team members currently assigned to this task.
          </p>
        </div>
      ) : (
        <ul className="mb-5 space-y-2">
          {currentAssigned.map((u) => {
            const initials = getInitials(u.full_name, u.email);
            return (
              <li
                key={u.id}
                className="flex items-center justify-between rounded-md border border-ink/10 bg-paper-dim/40 p-3 transition-colors hover:bg-paper-dim"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-ink text-xs font-bold text-paper shadow-xs">
                    {initials}
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="truncate text-sm font-semibold text-ink">
                        {u.full_name || "Team Member"}
                      </p>
                      <span className="shrink-0 rounded bg-ink/5 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-ink500">
                        {u.role}
                      </span>
                    </div>
                    {u.email && (
                      <div className="flex items-center gap-1 text-xs text-ink500">
                        <Mail className="h-3 w-3 shrink-0" />
                        <span className="truncate">{u.email}</span>
                      </div>
                    )}
                    {actionError[u.id] && (
                      <p className="text-xs font-medium text-status-rejected">
                        {actionError[u.id]}
                      </p>
                    )}
                  </div>
                </div>

                {canManageAssignments && (
                  <div>
                    {confirmUnassign === u.id ? (
                      <div className="flex items-center gap-1.5">
                        <button
                          type="button"
                          onClick={() => handleUnassign(u.id)}
                          className="rounded bg-status-rejected px-2.5 py-1 text-xs font-medium text-paper shadow-xs hover:bg-status-rejected/90"
                        >
                          Remove
                        </button>
                        <button
                          type="button"
                          onClick={() => setConfirmUnassign(null)}
                          className="rounded border border-ink/10 bg-paper px-2 py-1 text-xs font-medium text-ink hover:bg-ink/5"
                        >
                          Cancel
                        </button>
                      </div>
                    ) : (
                      <button
                        type="button"
                        onClick={() => setConfirmUnassign(u.id)}
                        className="rounded p-1.5 text-ink500 transition-colors hover:bg-status-rejected/10 hover:text-status-rejected"
                        title="Remove member from task"
                      >
                        <UserX className="h-4 w-4" />
                      </button>
                    )}
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      )}

      {/* Assignment Control for Managers and Admins */}
      {canManageAssignments && (
        <div className="border-t border-ink/10 pt-4">
          <label className="mb-2 block text-xs font-medium uppercase tracking-wider text-ink500">
            Assign Team Member
          </label>

          {/* Quick Select from Organization Directory */}
          {unassignedMembers.length > 0 && (
            <div className="mb-3">
              <div className="flex gap-2">
                <select
                  value={selectedUserEmail}
                  onChange={(e) => setSelectedUserEmail(e.target.value)}
                  disabled={loadingUsers}
                  className="flex-1 rounded border border-ink/15 bg-paper px-3 py-2 text-sm text-ink outline-none transition focus:border-ink"
                >
                  <option value="">
                    {loadingUsers
                      ? "Loading organization directory…"
                      : "-- Choose team member by name or email --"}
                  </option>
                  {unassignedMembers.map((m) => (
                    <option key={m.id} value={m.email}>
                      {m.full_name} ({m.email}) — {m.role.toUpperCase()}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  disabled={assigning || !selectedUserEmail}
                  onClick={() => handleAssignUser(selectedUserEmail)}
                  className="flex items-center gap-1.5 rounded bg-ink px-4 py-2 text-xs font-medium text-paper shadow-xs transition hover:bg-ink-light disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {assigning ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <UserPlus className="h-3.5 w-3.5" />
                  )}
                  <span>Assign</span>
                </button>
              </div>
            </div>
          )}

          {/* Or Search/Type by Name or Email directly */}
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleAssignUser(customInput);
            }}
            className="flex gap-2"
          >
            <div className="relative flex-1">
              <Search className="absolute top-2.5 left-3 h-4 w-4 text-ink/40" />
              <input
                type="text"
                value={customInput}
                onChange={(e) => {
                  setCustomInput(e.target.value);
                  if (error) setError(null);
                }}
                placeholder="Or type user email (e.g. alice@habeshan.local) or name…"
                className="w-full rounded border border-ink/15 bg-paper py-2 pr-3 pl-9 text-sm text-ink outline-none transition placeholder:text-ink/40 focus:border-ink"
              />
            </div>
            <button
              type="submit"
              disabled={assigning || !customInput.trim()}
              className="flex items-center gap-1.5 rounded border border-ink/20 bg-paper-dim px-4 py-2 text-xs font-semibold text-ink transition hover:bg-ink/5 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {assigning ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Check className="h-3.5 w-3.5" />
              )}
              <span>Add</span>
            </button>
          </form>

          {/* Error notice */}
          {error && (
            <div className="mt-2.5 flex items-center gap-1.5 text-xs text-status-rejected">
              <AlertCircle className="h-3.5 w-3.5 shrink-0" />
              <span>{error}</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
