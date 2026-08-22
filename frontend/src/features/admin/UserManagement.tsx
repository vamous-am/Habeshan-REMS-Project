import { useEffect, useMemo, useState } from "react";
import { Button } from "../../components/Button";
import { ErrorState } from "../../components/ErrorState";
import { StatusBadge } from "../../components/StatusBadge";
import {
  userRoleVariant,
  userStatusVariant,
} from "../../components/statusBadgeUtils";
import { Select } from "../../components/Select";
import { Table, type TableColumn } from "../../components/Table";
import {
  deleteUser,
  fetchUsers,
  updateUser,
  updateUserRole,
} from "../../lib/api/adminClient";
import type { UserDTO, UserRole } from "../../lib/api/authClient";
import { extractAuthErrorMessage } from "../../lib/api/authClient";

const ROLE_OPTIONS: { value: UserRole; label: string }[] = [
  { value: "admin", label: "Admin" },
  { value: "manager", label: "Manager" },
  { value: "employee", label: "Employee" },
];

export default function UserManagement() {
  const [users, setUsers] = useState<UserDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [apiError, setApiError] = useState<string | null>(null);
  const [actionUserId, setActionUserId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const data = await fetchUsers();
        if (!cancelled) {
          setUsers(data);
        }
      } catch (err) {
        if (!cancelled) {
          setApiError(extractAuthErrorMessage(err));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  async function handleRoleChange(userId: string, role: UserRole) {
    setActionUserId(userId);
    setApiError(null);
    try {
      const updated = await updateUserRole(userId, role);
      setUsers((current) =>
        current.map((user) => (user.id === userId ? updated : user))
      );
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setActionUserId(null);
    }
  }

  async function handleDeactivate(userId: string) {
    setActionUserId(userId);
    setApiError(null);
    try {
      const updated = await updateUser(userId, { status: "inactive" });
      setUsers((current) =>
        current.map((user) => (user.id === userId ? updated : user))
      );
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setActionUserId(null);
    }
  }

  async function handleReactivate(userId: string) {
    setActionUserId(userId);
    setApiError(null);
    try {
      const updated = await updateUser(userId, { status: "active" });
      setUsers((current) =>
        current.map((user) => (user.id === userId ? updated : user))
      );
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setActionUserId(null);
    }
  }

  async function handleDelete(userId: string, fullName: string) {
    const confirmed = window.confirm(
      `Permanently remove ${fullName}? This soft-deletes the account.`
    );
    if (!confirmed) return;

    setActionUserId(userId);
    setApiError(null);
    try {
      await deleteUser(userId);
      setUsers((current) => current.filter((user) => user.id !== userId));
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setActionUserId(null);
    }
  }

  const columns = useMemo<TableColumn<UserDTO>[]>(
    () => [
      {
        id: "name",
        header: "Name",
        sortValue: (user) => user.full_name,
        filterValue: (user) =>
          `${user.full_name} ${user.email} ${user.phone ?? ""}`,
        cell: (user) => (
          <div>
            <p className="font-medium text-ink900">{user.full_name}</p>
            {user.phone && (
              <p className="text-xs text-ink500">{user.phone}</p>
            )}
          </div>
        ),
      },
      {
        id: "email",
        header: "Email",
        sortValue: (user) => user.email,
        filterValue: (user) => user.email,
        cell: (user) => user.email,
      },
      {
        id: "role",
        header: "Role",
        sortValue: (user) => user.role,
        filterValue: (user) => user.role,
        cell: (user) => (
          <div className="flex flex-col gap-2">
            <StatusBadge
              label={user.role}
              variant={userRoleVariant(user.role)}
            />
            <Select
              aria-label={`Role for ${user.full_name}`}
              value={user.role}
              options={ROLE_OPTIONS}
              disabled={actionUserId === user.id}
              onChange={(event) =>
                void handleRoleChange(user.id, event.target.value as UserRole)
              }
              className="min-w-[140px]"
            />
          </div>
        ),
      },
      {
        id: "status",
        header: "Status",
        sortValue: (user) => user.status,
        filterValue: (user) => user.status,
        cell: (user) => (
          <StatusBadge
            label={user.status}
            variant={userStatusVariant(user.status)}
          />
        ),
      },
      {
        id: "actions",
        header: "Actions",
        cell: (user) => (
          <div className="flex flex-wrap gap-2">
            {user.status === "active" ? (
              <Button
                type="button"
                variant="secondary"
                disabled={actionUserId === user.id}
                onClick={() => void handleDeactivate(user.id)}
              >
                Deactivate
              </Button>
            ) : (
              <Button
                type="button"
                variant="secondary"
                disabled={actionUserId === user.id}
                onClick={() => void handleReactivate(user.id)}
              >
                Reactivate
              </Button>
            )}
            <Button
              type="button"
              variant="secondary"
              disabled={actionUserId === user.id}
              onClick={() => void handleDelete(user.id, user.full_name)}
              className="border-status-rejected/40 text-status-rejected hover:bg-status-rejected/5"
            >
              Remove
            </Button>
          </div>
        ),
      },
    ],
    [actionUserId]
  );

  return (
    <section>
      <div className="mb-6">
        <h2 className="font-display text-xl font-semibold text-ink">
          User management
        </h2>
        <p className="mt-1 text-sm text-ink500">
          List users, assign roles, deactivate accounts, or soft-delete.
        </p>
      </div>

      {apiError && (
        <div className="mb-4">
          <ErrorState message={apiError} />
        </div>
      )}

      {loading ? (
        <p className="text-sm text-ink500">Loading users…</p>
      ) : (
        <Table
          columns={columns}
          data={users}
          getRowKey={(user) => user.id}
          filterPlaceholder="Search by name, email, or phone"
          emptyMessage="No users in this organization."
        />
      )}
    </section>
  );
}
