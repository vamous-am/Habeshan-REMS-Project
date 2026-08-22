import { zodResolver } from "@hookform/resolvers/zod";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "../../components/Button";
import { Card } from "../../components/Card";
import { ErrorState } from "../../components/ErrorState";
import { Input } from "../../components/Input";
import { Select } from "../../components/Select";
import { Table, type TableColumn } from "../../components/Table";
import {
  addTeamMember,
  createTeam,
  fetchTeams,
  fetchUsers,
  removeTeamMember,
  type TeamDTO,
} from "../../lib/api/adminClient";
import type { UserDTO } from "../../lib/api/authClient";
import { extractAuthErrorMessage } from "../../lib/api/authClient";
import { useToast } from "../../components/Toast";

const createTeamSchema = z.object({
  name: z.string().min(2, "Team name must be at least 2 characters"),
  manager_id: z.string().min(1, "Select a manager"),
});

type CreateTeamForm = z.infer<typeof createTeamSchema>;

export default function TeamManagement() {
  const { showSuccess } = useToast();
  const [teams, setTeams] = useState<TeamDTO[]>([]);
  const [users, setUsers] = useState<UserDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [apiError, setApiError] = useState<string | null>(null);
  const [actionTeamId, setActionTeamId] = useState<string | null>(null);
  const [memberSelections, setMemberSelections] = useState<
    Record<string, string>
  >({});

  const form = useForm<CreateTeamForm>({
    resolver: zodResolver(createTeamSchema),
    defaultValues: { name: "", manager_id: "" },
  });

  const userNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const user of users) {
      map.set(user.id, user.full_name);
    }
    return map;
  }, [users]);

  const managerOptions = useMemo(
    () =>
      users.map((user) => ({
        value: user.id,
        label: `${user.full_name} (${user.role})`,
      })),
    [users]
  );

  const memberOptions = useMemo(
    () =>
      [{ value: "", label: "Select user…" }].concat(
        users.map((user) => ({
          value: user.id,
          label: user.full_name,
        }))
      ),
    [users]
  );

  const loadData = useCallback(async () => {
    setLoading(true);
    setApiError(null);
    try {
      const [teamList, userList] = await Promise.all([fetchTeams(), fetchUsers()]);
      setTeams(teamList);
      setUsers(userList);
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  async function handleCreateTeam(values: CreateTeamForm) {
    setApiError(null);
    try {
      const team = await createTeam({
        name: values.name.trim(),
        manager_id: values.manager_id,
      });
      setTeams((current) => [...current, team]);
      form.reset({ name: "", manager_id: "" });
      showSuccess(`Team "${team.name}" created.`);
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    }
  }

  async function handleAddMember(teamId: string) {
    const userId = memberSelections[teamId];
    if (!userId) {
      setApiError("Select a user to add to the team.");
      return;
    }

    setActionTeamId(teamId);
    setApiError(null);
    try {
      await addTeamMember(teamId, userId);
      setMemberSelections((current) => ({ ...current, [teamId]: "" }));
      showSuccess("Member added to team.");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setActionTeamId(null);
    }
  }

  async function handleRemoveMember(teamId: string) {
    const userId = memberSelections[teamId];
    if (!userId) {
      setApiError("Select a user to remove from the team.");
      return;
    }

    setActionTeamId(teamId);
    setApiError(null);
    try {
      await removeTeamMember(teamId, userId);
      setMemberSelections((current) => ({ ...current, [teamId]: "" }));
      showSuccess("Member removed from team.");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setActionTeamId(null);
    }
  }

  const columns = useMemo<TableColumn<TeamDTO>[]>(
    () => [
      {
        id: "name",
        header: "Team",
        sortValue: (team) => team.name,
        filterValue: (team) => team.name,
        cell: (team) => (
          <span className="font-medium text-ink900">{team.name}</span>
        ),
      },
      {
        id: "manager",
        header: "Manager",
        sortValue: (team) => userNameById.get(team.manager_id) ?? team.manager_id,
        filterValue: (team) => userNameById.get(team.manager_id) ?? "",
        cell: (team) =>
          userNameById.get(team.manager_id) ?? team.manager_id,
      },
      {
        id: "members",
        header: "Members",
        cell: (team) => (
          <div className="flex min-w-[280px] flex-col gap-2">
            <Select
              aria-label={`Add or remove member for ${team.name}`}
              value={memberSelections[team.id] ?? ""}
              options={memberOptions}
              disabled={actionTeamId === team.id}
              onChange={(event) =>
                setMemberSelections((current) => ({
                  ...current,
                  [team.id]: event.target.value,
                }))
              }
            />
            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                disabled={actionTeamId === team.id}
                onClick={() => void handleAddMember(team.id)}
              >
                Add
              </Button>
              <Button
                type="button"
                variant="secondary"
                disabled={actionTeamId === team.id}
                onClick={() => void handleRemoveMember(team.id)}
              >
                Remove
              </Button>
            </div>
          </div>
        ),
      },
    ],
    [actionTeamId, memberOptions, memberSelections, userNameById]
  );

  return (
    <section className="flex flex-col gap-8">
      <div>
        <h2 className="font-display text-xl font-semibold text-ink">
          Team management
        </h2>
        <p className="mt-1 text-sm text-ink500">
          Create teams, assign managers, and manage membership.
        </p>
      </div>

      {apiError && <ErrorState message={apiError} />}

      <Card title="Create team" className="max-w-lg">
        <form
          className="flex flex-col gap-4"
          onSubmit={form.handleSubmit(handleCreateTeam)}
          noValidate
        >
          <Input
            label="Team name"
            error={form.formState.errors.name?.message}
            {...form.register("name")}
          />
          <Select
            label="Manager"
            options={[{ value: "", label: "Select manager…" }, ...managerOptions]}
            error={form.formState.errors.manager_id?.message}
            {...form.register("manager_id")}
          />
          <Button type="submit">Create team</Button>
        </form>
      </Card>

      {loading ? (
        <p className="text-sm text-ink500">Loading teams…</p>
      ) : (
        <Table
          columns={columns}
          data={teams}
          getRowKey={(team) => team.id}
          filterPlaceholder="Search teams"
          emptyMessage="No teams yet. Create one above."
        />
      )}
    </section>
  );
}
