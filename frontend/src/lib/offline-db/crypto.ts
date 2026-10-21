/**
 * Generates a SHA-256 tamper-evident hash matching backend ComputeRecordHash
 * (record_uuid|user_id|action_type|timestamp).
 */
export async function generateDeviceHash(
  recordUuid: string,
  userId: string,
  actionType: string,
  timestamp: string
): Promise<string> {
  const rawPayload = `${recordUuid}|${userId}|${actionType}|${timestamp}`;

  const encoder = new TextEncoder();
  const data = encoder.encode(rawPayload);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);

  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}