import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ToastProvider } from "./components/Toast";
import { AppLayout } from "./components/AppLayout";
import { AdminLayout } from "./features/admin/AdminLayout";
import OrganizationSettings from "./features/admin/OrganizationSettings";
import TeamManagement from "./features/admin/TeamManagement";
import UserManagement from "./features/admin/UserManagement";
import { ClockWidget } from "./features/attendance/ClockWidget";
import Login from "./features/auth/Login";
import PasswordReset from "./features/auth/PasswordReset";
import Register from "./features/auth/Register";
import TaskListPage from "./features/tasks/pages/TaskListPage";
import TaskDetailPage from "./features/tasks/pages/TaskDetailPage";
import ApprovalQueue from "./features/timesheets/ApprovalQueue";
import TimesheetList from "./features/timesheets/TimesheetList";
import ManagerDashboardLive from "./features/dashboard/ManagerDashboardLive";
import ProtectedRoute from "./routes/ProtectedRoute";

function ForbiddenPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-paper p-6">
      <div className="max-w-md text-center">
        <h1 className="font-display text-2xl font-semibold text-ink">
          Access denied
        </h1>
        <p className="mt-2 text-sm text-ink500">
          You do not have permission to view this page with your current account role.
        </p>
        <a
          href="/tasks"
          className="mt-4 inline-block rounded bg-ink px-4 py-2 text-sm font-medium text-paper transition hover:bg-ink-light"
        >
          Go to tasks
        </a>
      </div>
    </div>
  );
}

export default function App() {
  return (
    <ToastProvider>
      <BrowserRouter>
        <Routes>
          {/* Public Authentication Routes */}
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/forgot-password" element={<PasswordReset />} />
          <Route path="/reset-password" element={<PasswordReset />} />
          <Route path="/forbidden" element={<ForbiddenPage />} />

          {/* Protected Application Routes wrapped with unified AppLayout */}
          <Route
            path="/attendance"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <div className="flex min-h-[calc(100vh-4.5rem)] items-center justify-center bg-paper p-6">
                    <ClockWidget />
                  </div>
                </AppLayout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/tasks"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <TaskListPage />
                </AppLayout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/tasks/:id"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <TaskDetailPage />
                </AppLayout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/timesheets"
            element={
              <ProtectedRoute>
                <AppLayout>
                  <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
                    <TimesheetList />
                  </div>
                </AppLayout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/approvals"
            element={
              <ProtectedRoute allowedRoles={["manager", "admin"]}>
                <AppLayout>
                  <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
                    <ApprovalQueue />
                  </div>
                </AppLayout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/dashboard"
            element={
              <ProtectedRoute allowedRoles={["manager", "admin"]}>
                <AppLayout>
                  <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
                    <ManagerDashboardLive />
                  </div>
                </AppLayout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin"
            element={
              <ProtectedRoute allowedRoles={["admin"]}>
                <AppLayout>
                  <AdminLayout />
                </AppLayout>
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/admin/users" replace />} />
            <Route path="users" element={<UserManagement />} />
            <Route path="teams" element={<TeamManagement />} />
            <Route path="settings" element={<OrganizationSettings />} />
          </Route>

          {/* 404 Fallback */}
          <Route
            path="*"
            element={
              <div className="flex min-h-screen flex-col items-center justify-center p-6 text-center font-body text-ink">
                <h1 className="font-display text-3xl font-bold text-ink">404</h1>
                <p className="mt-2 text-base text-ink500">Page not found</p>
                <a
                  href="/tasks"
                  className="mt-4 rounded bg-ink px-4 py-2 text-sm font-medium text-paper hover:bg-ink-light"
                >
                  Return to application
                </a>
              </div>
            }
          />
        </Routes>
      </BrowserRouter>
    </ToastProvider>
  );
}
