/**
 * CreateTaskForm.tsx
 *
 * Minimal form for creating a task (FR-TASK-01).
 * Only rendered for admin/manager callers — role enforcement is on the backend;
 * the frontend simply shows the form and lets the backend reject it if needed.
 */

import { useState } from "react";
import type { Priority } from "../types";
import { createTask, extractApiErrorMessage } from "../api";

interface Props {
  onCreated: () => void; // called after successful creation to trigger list refresh
}

export default function CreateTaskForm({ onCreated }: Props) {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState<Priority>("medium");
  const [dueDate, setDueDate] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) {
      setError("Title is required.");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      await createTask({
        title: title.trim(),
        description: description.trim() || undefined,
        priority,
        due_date: dueDate || undefined,
      });
      setTitle("");
      setDescription("");
      setPriority("medium");
      setDueDate("");
      setOpen(false);
      onCreated();
    } catch (e: unknown) {
      setError(extractApiErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }

  if (!open) {
    return (
      <button
        type="button"
        onClick={() => setOpen(true)}
        style={{
          marginBottom: 16,
          padding: "8px 14px",
          fontSize: 13,
          fontWeight: 500,
          background: "#2563eb",
          color: "#fff",
          border: "none",
          borderRadius: 6,
          cursor: "pointer",
        }}
      >
        + New Task
      </button>
    );
  }

  return (
    <form
      onSubmit={handleSubmit}
      style={{
        border: "1px solid #d1d5db",
        borderRadius: 8,
        padding: 16,
        marginBottom: 16,
        maxWidth: 480,
      }}
    >
      <h3 style={{ margin: "0 0 12px" }}>New Task</h3>

      {error && (
        <p style={{ color: "red", fontSize: 13, marginBottom: 8 }}>{error}</p>
      )}

      <label style={labelStyle}>
        Title *
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
          maxLength={150}
          style={inputStyle}
        />
      </label>

      <label style={labelStyle}>
        Description
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          style={{ ...inputStyle, resize: "vertical" }}
        />
      </label>

      <label style={labelStyle}>
        Priority
        <select
          value={priority}
          onChange={(e) => setPriority(e.target.value as Priority)}
          style={inputStyle}
        >
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
      </label>

      <label style={labelStyle}>
        Due Date
        <input
          type="date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
          style={inputStyle}
        />
      </label>

      <div style={{ display: "flex", gap: 8, marginTop: 4 }}>
        <button type="submit" disabled={loading}>
          {loading ? "Creating…" : "Create Task"}
        </button>
        <button
          type="button"
          onClick={() => setOpen(false)}
          disabled={loading}
        >
          Cancel
        </button>
      </div>
    </form>
  );
}

const labelStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: 4,
  marginBottom: 12,
  fontSize: 13,
  fontWeight: 500,
};

const inputStyle: React.CSSProperties = {
  padding: "6px 8px",
  border: "1px solid #d1d5db",
  borderRadius: 4,
  fontSize: 14,
  width: "100%",
  boxSizing: "border-box",
};
