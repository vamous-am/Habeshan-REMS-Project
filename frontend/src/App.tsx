import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ToastProvider } from "./components/Toast";
import { ClockWidget } from "./features/attendance/ClockWidget";
import Login from "./features/auth/Login";
import PasswordReset from "./features/auth/PasswordReset";
import Register from "./features/auth/Register";
import TaskListPage from "./features/tasks/pages/TaskListPage";
import TaskDetailPage from "./features/tasks/pages/TaskDetailPage";

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

          <Route
            path="/attendance"
            element={
              <div className="flex min-h-screen items-center justify-center bg-paper p-4">
                <ClockWidget />
              </div>
            }
          />

          <Route path="/tasks" element={<TaskListPage />} />
          <Route path="/tasks/:id" element={<TaskDetailPage />} />

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
