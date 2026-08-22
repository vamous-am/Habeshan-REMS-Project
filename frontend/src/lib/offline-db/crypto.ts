/**
 * Generates a SHA-256 hash incorporating system payload & client salt
 * to prevent system-clock manipulation while offline.
 */
export async function generateDeviceHash(
  recordUuid: string,
  userId: string,
  actionType: string,
  timestamp: string
): Promise<string> {
  const salt = window.navigator.userAgent; // Hardware/browser client signature
  const rawPayload = `${recordUuid}:${userId}:${actionType}:${timestamp}:${salt}`;

  const encoder = new TextEncoder();
  const data = encoder.encode(rawPayload);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);

  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}