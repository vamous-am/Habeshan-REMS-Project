/**
 * api/client.ts
 *
 * Thin axios wrapper.  All tasks API calls go through this instance.
 * Base URL is read from VITE_API_BASE_URL (.env.local / .env).
 *
 * Auth headers (X-User-ID, X-Org-ID) are injected by the request
 * interceptor below from localStorage.  These are the same headers the
 * backend handler reads via callerFromCtx() until JWT middleware is wired in.
 */

import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080",
  headers: { "Content-Type": "application/json" },
});

// Inject stub identity headers so manual testing works without a real auth flow.
api.interceptors.request.use((config) => {
  const userID = localStorage.getItem("x-user-id");
  const orgID = localStorage.getItem("x-org-id");
  if (userID) config.headers["X-User-ID"] = userID;
  if (orgID) config.headers["X-Org-ID"] = orgID;
  return config;
});

export default api;
