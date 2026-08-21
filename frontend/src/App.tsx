import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ClockWidget } from "./features/attendance/ClockWidget";
import TaskListPage from "./features/tasks/pages/TaskListPage";
import TaskDetailPage from "./features/tasks/pages/TaskDetailPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Default redirect */}
        <Route path="/" element={<Navigate to="/attendance" replace />} />

        {/* Attendance Dashboard */}
        <Route
          path="/attendance"
          element={
            <div className="min-h-screen bg-slate-100 flex items-center justify-center p-4">
              <ClockWidget />
            </div>
          }
        />

        {/* Tasks */}
        <Route path="/tasks" element={<TaskListPage />} />
        <Route path="/tasks/:id" element={<TaskDetailPage />} />

        {/* Catch-all route */}
        <Route
          path="*"
          element={
            <div style={{ padding: 24, fontFamily: "sans-serif" }}>
              <h2>Page not found</h2>
              <a href="/tasks">Go to Tasks</a>
            </div>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}