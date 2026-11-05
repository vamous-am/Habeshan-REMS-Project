/**
 * CreateTaskForm.tsx
 *
 * Minimal form for creating a task (FR-TASK-01).
 * Only rendered for admin/manager callers — role enforcement is on the backend;
 * the frontend simply shows the form and lets the backend reject it if needed.
 *
 * Presentation uses the shared light grouped form components, the same
 * layout rules as the Add User and Create Team forms. The form logic is
 * unchanged.
 */

import { useState } from "react";
import type { Priority } from "../types";
import { createTask, extractApiErrorMessage } from "../api";
import {
  FormFieldGroup,
  GroupedFormButton,
  GroupedFormPanel,
  GroupedInputField,
  GroupedSelectField,
  GroupedTextareaField,
} from "../../../components/GroupedForm";

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
      <GroupedFormButton
        type="button"
        onClick={() => setOpen(true)}
        className="mb-4"
      >
        + New Task
      </GroupedFormButton>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="mb-4">
      <GroupedFormPanel
        title="New task"
        subtitle="Assign work with a clear deadline."
      >
        {error && (
          <div
            role="alert"
            className="mb-5 rounded-2xl border border-status-rejected/30 bg-status-rejected/5 px-4 py-3 text-sm text-status-rejected"
          >
            {error}
          </div>
        )}

        <FormFieldGroup>
          <GroupedInputField
            label="Title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            maxLength={150}
          />
          <GroupedTextareaField
            label="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
          />
          <GroupedSelectField
            label="Priority"
            options={[
              { value: "high", label: "High" },
              { value: "medium", label: "Medium" },
              { value: "low", label: "Low" },
            ]}
            value={priority}
            onChange={(e) => setPriority(e.target.value as Priority)}
          />
          <GroupedInputField
            label="Due date"
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
          />
        </FormFieldGroup>

        <div className="mt-5 flex gap-3">
          <GroupedFormButton
            type="submit"
            loading={loading}
            className="flex-1"
          >
            Create Task
          </GroupedFormButton>
          <GroupedFormButton
            type="button"
            variant="secondary"
            onClick={() => setOpen(false)}
            disabled={loading}
            className="px-6"
          >
            Cancel
          </GroupedFormButton>
        </div>
      </GroupedFormPanel>
    </form>
  );
}
