export type StatusVariant = "verified" | "pending" | "rejected" | "offline";

export function userStatusVariant(status: string): StatusVariant {
  if (status === "active") return "verified";
  if (status === "inactive") return "offline";
  return "pending";
}

export function userRoleVariant(role: string): StatusVariant {
  if (role === "admin") return "verified";
  if (role === "manager") return "pending";
  return "offline";
}
