import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ToastProvider } from "./components/Toast";
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
import ProtectedRoute from "./routes/ProtectedRoute";

function ForbiddenPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-paper p-6">
      <div className="max-w-md text-center">
        <h1 className="font-display text-2xl font-semibold text-ink">
          Access denied
        </h1>
        <p className="mt-2 text-sm text-ink500">
          You do not have permission to view this page.
        </p>
        <a
          href="/tasks"
          className="mt-4 inline-block text-sm text-ink underline underline-offset-2"
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
          <Route path="/" element={<Navigate to="/login" replace />} />

          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/forgot-password" element={<PasswordReset />} />
          <Route path="/reset-password" element={<PasswordReset />} />
          <Route path="/forbidden" element={<ForbiddenPage />} />

          <Route
            path="/attendance"
            element={
              <ProtectedRoute>
                <div className="flex min-h-screen items-center justify-center bg-paper p-4">
                  <ClockWidget />
                </div>
              </ProtectedRoute>
            }
          />

          <Route
            path="/tasks"
            element={
              <ProtectedRoute>
                <TaskListPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/tasks/:id"
            element={
              <ProtectedRoute>
                <TaskDetailPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin"
            element={
              <ProtectedRoute allowedRoles={["admin"]}>
                <AdminLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/admin/users" replace />} />
            <Route path="users" element={<UserManagement />} />
            <Route path="teams" element={<TeamManagement />} />
            <Route path="settings" element={<OrganizationSettings />} />
          </Route>

          <Route
            path="*"
            element={
              <div className="p-6 font-body text-ink">
                <h2 className="font-display text-xl">Page not found</h2>
                <a href="/login" className="mt-2 inline-block text-ink underline">
                  Go to sign in
                </a>
              </div>
            }
          />
        </Routes>
      </BrowserRouter>
    </ToastProvider>
  );
}
