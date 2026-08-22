/**
 * lib/api/adminClient.ts
 *
 * HTTP client for /admin/* endpoints. Uses the shared axios instance
 * which injects the Bearer token from authClient.
 */

import api from "./client";
import type { UserDTO, UserRole } from "./authClient";
import { AuthApiError } from "./authClient";

interface Envelope<T> {
  status: "success" | "fail";
  data?: T;
  message?: string;
}

function unwrap<T>(response: { data: Envelope<T>; status: number }): T {
  const env = response.data;
  if (env?.status === "success" && env.data !== undefined) {
    return env.data;
  }
  if (env?.status === "fail" && env.message) {
    throw new AuthApiError(env.message, response.status);
  }
  throw new AuthApiError("Unexpected response from server", response.status);
}

export interface TeamDTO {
  id: string;
  org_id: string;
  name: string;
  manager_id: string;
}

export interface OrgDTO {
  id: string;
  name: string;
  currency: string;
  timezone: string;
  seat_count: number;
  subscription_status: "trial" | "active" | "suspended";
}

export interface UpdateUserPayload {
  full_name?: string;
  phone?: string;
  status?: "active" | "inactive";
}

export interface UpdateOrgPayload {
  name?: string;
  currency?: string;
  timezone?: string;
}

export interface CreateTeamPayload {
  name: string;
  manager_id: string;
}

const BASE = "/admin";

export async function fetchUsers(): Promise<UserDTO[]> {
  return unwrap(await api.get(`${BASE}/users`));
}

export async function fetchUser(userId: string): Promise<UserDTO> {
  return unwrap(await api.get(`${BASE}/users/${userId}`));
}

export async function updateUser(
  userId: string,
  payload: UpdateUserPayload
): Promise<UserDTO> {
  return unwrap(await api.patch(`${BASE}/users/${userId}`, payload));
}

export async function updateUserRole(
  userId: string,
  role: UserRole
): Promise<UserDTO> {
  return unwrap(await api.patch(`${BASE}/users/${userId}/role`, { role }));
}

export async function deleteUser(userId: string): Promise<{ deleted: boolean }> {
  return unwrap(await api.delete(`${BASE}/users/${userId}`));
}

export async function fetchTeams(): Promise<TeamDTO[]> {
  return unwrap(await api.get(`${BASE}/teams`));
}

export async function createTeam(payload: CreateTeamPayload): Promise<TeamDTO> {
  return unwrap(await api.post(`${BASE}/teams`, payload));
}

export async function addTeamMember(
  teamId: string,
  userId: string
): Promise<{ added: boolean }> {
  return unwrap(
    await api.post(`${BASE}/teams/${teamId}/members`, { user_id: userId })
  );
}

export async function removeTeamMember(
  teamId: string,
  userId: string
): Promise<{ removed: boolean }> {
  return unwrap(await api.delete(`${BASE}/teams/${teamId}/members/${userId}`));
}

export async function fetchOrg(): Promise<OrgDTO> {
  return unwrap(await api.get(`${BASE}/org`));
}

export async function updateOrg(payload: UpdateOrgPayload): Promise<OrgDTO> {
  return unwrap(await api.patch(`${BASE}/org`, payload));
}
