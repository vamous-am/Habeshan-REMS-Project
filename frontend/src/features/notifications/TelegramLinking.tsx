import { useEffect, useState } from "react";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL as string;

// TODO: replace this with real userID pulled from JWT/auth context
// once Dev 1's auth is wired into the frontend
const TEMP_USER_ID = "00000000-0000-0000-0000-000000000000";

interface TelegramStatus {
  linked: boolean;
  chat_id?: string;
  is_active?: boolean;
  linked_at?: string;
}

export default function TelegramLinking() {
  const [status, setStatus] = useState<TelegramStatus | null>(null);
  const [chatID, setChatID] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Fetch current link status on mount
  useEffect(() => {
    async function fetchStatus() {
      try {
        const res = await fetch(
          `${API_BASE_URL}/notifications/telegram/status?user_id=${TEMP_USER_ID}`
        );
        const json = await res.json();
        setStatus(json.data);
      } catch {
        setError("Could not fetch Telegram status.");
      } finally {
        setLoading(false);
      }
    }
    fetchStatus();
  }, []);

  async function handleLink() {
    if (!chatID.trim()) {
      setError("Please enter your Telegram Chat ID.");
      return;
    }
    setSubmitting(true);
    setError(null);
    setMessage(null);
    try {
      const res = await fetch(
        `${API_BASE_URL}/notifications/telegram/link?user_id=${TEMP_USER_ID}`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ chat_id: chatID }),
        }
      );
      const json = await res.json();
      if (!res.ok) {
        setError(json.message ?? "Something went wrong.");
      } else {
        setMessage("Telegram linked successfully!");
        setStatus({ linked: true, chat_id: chatID, is_active: true });
      }
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) {
    return <div className="p-6 text-sm text-gray-500">Loading…</div>;
  }

  return (
    <div className="p-6 max-w-md space-y-6">
      <h1 className="text-xl font-semibold">Telegram Notifications</h1>

      {status?.linked ? (
        <div className="border rounded-lg p-4 space-y-1">
          <p className="text-sm text-green-600 font-medium">✓ Telegram linked</p>
          <p className="text-sm text-gray-500">Chat ID: {status.chat_id}</p>
          <p className="text-sm text-gray-500">
            Status: {status.is_active ? "Active" : "Paused"}
          </p>
        </div>
      ) : (
        <div className="border rounded-lg p-4 space-y-1">
          <p className="text-sm text-gray-500">No Telegram account linked yet.</p>
        </div>
      )}

      <div className="space-y-3">
        <p className="text-sm text-gray-600">
          To link your Telegram, start a chat with our bot and paste your Chat ID below.
        </p>
        <input
          type="text"
          placeholder="Enter your Telegram Chat ID"
          value={chatID}
          onChange={(e) => setChatID(e.target.value)}
          className="w-full border rounded px-3 py-2 text-sm"
        />
        <button
          onClick={handleLink}
          disabled={submitting}
          className="w-full bg-blue-600 text-white rounded px-4 py-2 text-sm font-medium disabled:opacity-50"
        >
          {submitting ? "Linking…" : status?.linked ? "Re-link Telegram" : "Link Telegram"}
        </button>
      </div>

      {message && <p className="text-sm text-green-600">{message}</p>}
      {error && <p className="text-sm text-red-600">{error}</p>}
    </div>
  );
}
