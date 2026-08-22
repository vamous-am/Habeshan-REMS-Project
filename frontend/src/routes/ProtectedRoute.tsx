import { Navigate, useLocation } from "react-router-dom";
import type { ReactNode } from "react";
import {
  getUserRole,
  isAuthenticated,
  type UserRole,
} from "../lib/api/authClient";

interface ProtectedRouteProps {
  children: ReactNode;
  allowedRoles?: UserRole[];
}

export default function ProtectedRoute({
  children,
  allowedRoles,
}: ProtectedRouteProps) {
  const location = useLocation();

  if (!isAuthenticated()) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  if (allowedRoles && allowedRoles.length > 0) {
    const role = getUserRole();
    if (!role || !allowedRoles.includes(role)) {
      return <Navigate to="/forbidden" replace />;
    }
  }

  return children;
}
