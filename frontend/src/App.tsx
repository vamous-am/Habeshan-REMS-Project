/**
 * App.tsx
 *
 * Root component — sets up routing.
 * Only the Tasks feature is wired for now; other features will be added
 * as their backends are implemented.
 */

import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import TaskListPage from "./features/tasks/pages/TaskListPage";
import TaskDetailPage from "./features/tasks/pages/TaskDetailPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Default redirect to task list */}
        <Route path="/" element={<Navigate to="/tasks" replace />} />

        {/* Tasks */}
        <Route path="/tasks" element={<TaskListPage />} />
        <Route path="/tasks/:id" element={<TaskDetailPage />} />

        {/* Catch-all */}
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
