/**
 * lib/api/authClient.ts
 *
 * HTTP client for auth endpoints. All calls unwrap the standard
 * { status, data } envelope from the backend.
 */

import axios, { type AxiosResponse } from "axios";

const AUTH_BASE =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api/v1";

const authHttp = axios.create({
  baseURL: AUTH_BASE,
  headers: { "Content-Type": "application/json" },
});

interface Envelope<T> {
  status: "success" | "fail";
  data?: T;
  message?: string;
}

export class AuthApiError extends Error {
  readonly status?: number;

  constructor(message: string, status?: number) {
    super(message);
    this.name = "AuthApiError";
    this.status = status;
  }
}

function unwrap<T>(response: AxiosResponse<Envelope<T>>): T {
  const env = response.data;
  if (env?.status === "success" && env.data !== undefined) {
    return env.data;
  }
  if (env?.status === "fail" && env.message) {
    throw new AuthApiError(env.message, response.status);
  }
  throw new AuthApiError("Unexpected response from server", response.status);
}

export interface OrgSummary {
  org_id: string;
  org_name: string;
}

export interface UserDTO {
  id: string;
  org_id: string;
  email: string;
  full_name: string;
  phone?: string;
  role: string;
  status: string;
}

export interface AuthResponse {
  token: string;
  user: UserDTO;
}

export interface LookupResponse {
  orgs: OrgSummary[];
}

export interface ForgotPasswordResponse {
  reset_token: string;
}

export interface RegisterPayload {
  org_name: string;
  full_name: string;
  email: string;
  password: string;
  phone?: string;
}

export interface LoginPayload {
  email: string;
  password: string;
  org_id: string;
}

export interface ForgotPasswordPayload {
  email: string;
  org_id: string;
}

export interface ResetPasswordPayload {
  reset_token: string;
  new_password: string;
}

const AUTH_TOKEN_KEY = "auth_token";
const USER_ID_KEY = "x-user-id";
const ORG_ID_KEY = "x-org-id";

export function persistAuthSession({ token, user }: AuthResponse): void {
  localStorage.setItem(AUTH_TOKEN_KEY, token);
  localStorage.setItem(USER_ID_KEY, user.id);
  localStorage.setItem(ORG_ID_KEY, user.org_id);
}

export function clearAuthSession(): void {
  localStorage.removeItem(AUTH_TOKEN_KEY);
  localStorage.removeItem(USER_ID_KEY);
  localStorage.removeItem(ORG_ID_KEY);
}

export function getAuthToken(): string | null {
  return localStorage.getItem(AUTH_TOKEN_KEY);
}

export async function lookupOrgs(email: string): Promise<LookupResponse> {
  return unwrap(await authHttp.post("/auth/lookup", { email }));
}

export async function register(payload: RegisterPayload): Promise<AuthResponse> {
  return unwrap(await authHttp.post("/auth/register", payload));
}

export async function login(payload: LoginPayload): Promise<AuthResponse> {
  return unwrap(await authHttp.post("/auth/login", payload));
}

export async function forgotPassword(
  payload: ForgotPasswordPayload
): Promise<ForgotPasswordResponse> {
  return unwrap(await authHttp.post("/auth/forgot-password", payload));
}

export async function resetPassword(
  payload: ResetPasswordPayload
): Promise<{ message: string }> {
  return unwrap(await authHttp.post("/auth/reset-password", payload));
}

export async function logout(): Promise<void> {
  const token = getAuthToken();
  if (token) {
    try {
      await authHttp.post(
        "/auth/logout",
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
    } catch {
      // Logout is client-driven; ignore server errors.
    }
  }
  clearAuthSession();
}

export function extractAuthErrorMessage(err: unknown): string {
  if (err instanceof AuthApiError) return err.message;
  if (
    err &&
    typeof err === "object" &&
    "response" in (err as Record<string, unknown>)
  ) {
    const resp = (err as { response?: { data?: { message?: string } } })
      .response;
    const msg = resp?.data?.message;
    if (msg && typeof msg === "string") return msg;
  }
  if (err instanceof Error) return err.message;
  return "An unexpected error occurred";
}
