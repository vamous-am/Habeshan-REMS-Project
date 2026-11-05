import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "../../components/Button";
import { ErrorState } from "../../components/ErrorState";
import {
  FormFieldGroup,
  FormNote,
  GroupedFormButton,
  GroupedFormPanel,
  GroupedInputField,
  GroupedSelectField,
} from "../../components/GroupedForm";
import { StatusBadge } from "../../components/StatusBadge";
import {
  userRoleVariant,
  userStatusVariant,
} from "../../components/statusBadgeUtils";
import { Select } from "../../components/Select";
import { Table, type TableColumn } from "../../components/Table";
import { useToast } from "../../components/useToast";
import {
  createUser,
  deleteUser,
  fetchUsers,
  updateUser,
  updateUserRole,
} from "../../lib/api/adminClient";
import type { UserDTO, UserRole } from "../../lib/api/authClient";
import { extractAuthErrorMessage } from "../../lib/api/authClient";
import {
  ETHIOPIAN_PHONE_MESSAGE,
  ETHIOPIAN_PHONE_PATTERN,
} from "../../lib/phone";

const ROLE_OPTIONS: { value: UserRole; label: string }[] = [
  { value: "admin", label: "Admin" },
  { value: "manager", label: "Manager" },
  { value: "employee", label: "Employee" },
];

const CREATE_ROLE_OPTIONS: {
  value: "manager" | "employee";
  label: string;
}[] = [
  { value: "employee", label: "Employee" },
  { value: "manager", label: "Manager" },
];

const createUserSchema = z.object({
  full_name: z.string().min(2, "Full name must be at least 2 characters"),
  email: z.string().email("Enter a valid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  role: z.enum(["manager", "employee"]),
  phone: z
    .string()
    .regex(ETHIOPIAN_PHONE_PATTERN, ETHIOPIAN_PHONE_MESSAGE)
    .optional()
    .or(z.literal("")),
});

type CreateUserForm = z.infer<typeof createUserSchema>;

export default function UserManagement() {
  const { showSuccess } = useToast();
  const [users, setUsers] = useState<UserDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [apiError, setApiError] = useState<string | null>(null);
  const [actionUserId, setActionUserId] = useState<string | null>(null);

  const form = useForm<CreateUserForm>({
    resolver: zodResolver(createUserSchema),
    defaultValues: {
      full_name: "",
      email: "",
      password: "",
      role: "employee",
      phone: "",
    },
  });

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

  async function handleCreateUser(values: CreateUserForm) {
    setApiError(null);
    try {
      const user = await createUser({
        full_name: values.full_name.trim(),
        email: values.email.trim(),
        password: values.password,
        role: values.role,
        phone: values.phone?.trim() || undefined,
      });
      setUsers((current) => [...current, user]);
      form.reset({
        full_name: "",
        email: "",
        password: "",
        role: "employee",
        phone: "",
      });
      showSuccess(`User "${user.full_name}" created.`);
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    }
  }

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
    <section className="flex flex-col gap-8">
      <div>
        <h2 className="font-display text-xl font-semibold text-ink">
          User management
        </h2>
        <p className="mt-1 text-sm text-ink500">
          Create users, assign roles, deactivate accounts, or soft-delete.
        </p>
      </div>

      {apiError && <ErrorState message={apiError} />}

      <GroupedFormPanel
        title="Add user"
        subtitle="Create a manager or employee account for this organization."
      >
        <form
          onSubmit={form.handleSubmit(handleCreateUser)}
          noValidate
        >
          <FormFieldGroup>
            <GroupedInputField
              label="Full name"
              autoComplete="name"
              error={form.formState.errors.full_name?.message}
              {...form.register("full_name")}
            />
            <GroupedInputField
              label="Email"
              type="email"
              autoComplete="email"
              error={form.formState.errors.email?.message}
              {...form.register("email")}
            />
            <GroupedInputField
              label="Password"
              type="password"
              autoComplete="new-password"
              hint="At least 8 characters"
              error={form.formState.errors.password?.message}
              {...form.register("password")}
            />
            <GroupedSelectField
              label="Role"
              options={CREATE_ROLE_OPTIONS}
              error={form.formState.errors.role?.message}
              {...form.register("role")}
            />
            <GroupedInputField
              label="Phone"
              type="tel"
              autoComplete="tel"
              placeholder="09XXXXXXXX (optional)"
              error={form.formState.errors.phone?.message}
              {...form.register("phone")}
            />
          </FormFieldGroup>
          <FormNote title="Admin is not offered here.">
            The first admin still comes from the CLI seed-admin command.
          </FormNote>
          <GroupedFormButton
            type="submit"
            loading={form.formState.isSubmitting}
            className="mt-5 w-full"
          >
            Add user
          </GroupedFormButton>
        </form>
      </GroupedFormPanel>

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
