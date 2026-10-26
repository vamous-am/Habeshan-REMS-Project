/**
 * Offline Storage Cryptography & Key Management
 *
 * ARCHITECTURAL TRADEOFF & THREAT MODEL STATEMENT:
 * Offline records use a persistent device-held AES-256-GCM key stored in IndexedDB.
 * `user_id`, `org_id`, `record_uuid`, and `sync_status` remain plaintext indexed metadata;
 * sensitive attendance fields (action_type, timestamp, device_hash) remain encrypted at rest.
 *
 * AES-GCM Additional Authenticated Data (AAD) cryptographically binds ciphertext to its
 * metadata and detects record substitution or tampering, but does not by itself provide
 * cross-user confidentiality against other users sharing the same browser origin and device key.
 * Per-user routing and tamper binding are supported by plaintext metadata plus AAD; strong
 * per-user confidentiality is not guaranteed while multiple users share the same persistent device key.
 *
 * This intentionally deviates from the literal session-derived-key wording of NFR-SEC-02
 * to avoid making unsynced records unrecoverable after token expiry or logout in offline conditions.
 */

export interface AttendanceSensitivePayload {
  action_type: 'CLOCK_IN' | 'CLOCK_OUT';
  timestamp: string;
  device_hash: string;
}

export interface EncryptedPayloadBundle {
  iv: string; // 12-byte IV (Base64)
  ciphertext: string; // AES-GCM ciphertext + auth tag (Base64)
  crypto_version: number;
}

const KEY_DB_NAME = 'HabeshanREMS_KeyStore';
const KEY_STORE_NAME = 'device_keys';
const DEVICE_KEY_ID = 'offline_aes_gcm_v1';
export const CURRENT_CRYPTO_VERSION = 1;

let cachedDeviceKey: CryptoKey | null = null;

// ==========================================
// Canonical AAD Construction Helper
// ==========================================

/**
 * Builds the canonical Additional Authenticated Data (AAD) byte array.
 * Contract: attendance-log:v1|record_uuid|org_id|user_id
 * MUST be used identically across encryption, decryption, and validation.
 */
export function buildAttendanceAAD(
  recordUuid: string,
  orgId: string,
  userId: string
): Uint8Array {
  const canonicalString = `attendance-log:v1|${recordUuid}|${orgId}|${userId}`;
  return new TextEncoder().encode(canonicalString);
}

// ==========================================
// IndexedDB Device Key Storage (Structured Clone)
// ==========================================

function openKeyDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    if (typeof window === 'undefined' || !window.indexedDB) {
      return reject(new Error('IndexedDB is unavailable in this environment.'));
    }

    const request = window.indexedDB.open(KEY_DB_NAME, 1);
    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(KEY_STORE_NAME)) {
        db.createObjectStore(KEY_STORE_NAME);
      }
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error || new Error('Failed to open key database'));
  });
}

/**
 * Retrieves the existing device encryption key from IndexedDB without creating a new one.
 * Returns null if no key has been generated yet.
 */
export async function getExistingDeviceKey(): Promise<CryptoKey | null> {
  if (cachedDeviceKey) return cachedDeviceKey;

  try {
    const db = await openKeyDatabase();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(KEY_STORE_NAME, 'readonly');
      const store = tx.objectStore(KEY_STORE_NAME);
      const getReq = store.get(DEVICE_KEY_ID);

      getReq.onsuccess = () => {
        const key = (getReq.result as CryptoKey) || null;
        if (key) cachedDeviceKey = key;
        resolve(key);
      };
      getReq.onerror = () => reject(getReq.error || new Error('Failed to retrieve device key'));
    });
  } catch (err) {
    console.error('[Crypto] Error reading device key:', err);
    return null;
  }
}

/**
 * Retrieves the device key or generates a new AES-256-GCM key if none exists.
 * Does NOT replace an existing key if retrieval fails.
 */
export async function getOrCreateDeviceKey(): Promise<CryptoKey> {
  if (cachedDeviceKey) return cachedDeviceKey;

  const existing = await getExistingDeviceKey();
  if (existing) {
    cachedDeviceKey = existing;
    return existing;
  }

  // Generate a non-extractable 256-bit AES-GCM key
  const newKey = await crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    false, // non-extractable
    ['encrypt', 'decrypt']
  );

  const db = await openKeyDatabase();
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(KEY_STORE_NAME, 'readwrite');
    const store = tx.objectStore(KEY_STORE_NAME);
    const putReq = store.put(newKey, DEVICE_KEY_ID);

    putReq.onsuccess = () => resolve();
    putReq.onerror = () => reject(putReq.error || new Error('Failed to store generated device key'));
  });

  cachedDeviceKey = newKey;
  return newKey;
}

// ==========================================
// Base64 Encoding / Decoding Helpers
// ==========================================

export function bufferToBase64(buffer: ArrayBuffer | Uint8Array): string {
  const bytes = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

export function base64ToBuffer(base64: string): Uint8Array {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

// ==========================================
// AES-256-GCM Encrypt & Decrypt with AAD
// ==========================================

/**
 * Encrypts sensitive attendance payload fields using AES-256-GCM with a fresh 12-byte IV
 * and Additional Authenticated Data (AAD) binding record_uuid, org_id, and user_id.
 */
export async function encryptAttendancePayload(
  recordUuid: string,
  orgId: string,
  userId: string,
  payload: AttendanceSensitivePayload,
  key: CryptoKey
): Promise<EncryptedPayloadBundle> {
  // 12-byte cryptographically secure random IV — NEVER reuse an IV with AES-GCM
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const aad = buildAttendanceAAD(recordUuid, orgId, userId);
  const plainTextBytes = new TextEncoder().encode(JSON.stringify(payload));

  const cipherBuffer = await crypto.subtle.encrypt(
    {
      name: 'AES-GCM',
      iv: iv as BufferSource,
      additionalData: aad as BufferSource,
    },
    key,
    plainTextBytes as BufferSource
  );

  return {
    iv: bufferToBase64(iv),
    ciphertext: bufferToBase64(cipherBuffer),
    crypto_version: CURRENT_CRYPTO_VERSION,
  };
}

/**
 * Decrypts sensitive attendance payload fields using AES-256-GCM.
 * Validates the authentication tag against the provided AAD (record_uuid, org_id, user_id).
 * Throws an explicit Error if tampered, corrupted, or AAD does not match.
 */
export async function decryptAttendancePayload(
  recordUuid: string,
  orgId: string,
  userId: string,
  ivBase64: string,
  ciphertextBase64: string,
  key: CryptoKey
): Promise<AttendanceSensitivePayload> {
  const iv = base64ToBuffer(ivBase64);
  if (iv.byteLength !== 12) {
    throw new Error(`Invalid IV length: expected 12 bytes, got ${iv.byteLength}`);
  }

  const ciphertext = base64ToBuffer(ciphertextBase64);
  const aad = buildAttendanceAAD(recordUuid, orgId, userId);

  try {
    const plainBuffer = await crypto.subtle.decrypt(
      {
        name: 'AES-GCM',
        iv: iv as BufferSource,
        additionalData: aad as BufferSource,
      },
      key,
      ciphertext as BufferSource
    );

    const decryptedJson = new TextDecoder().decode(plainBuffer);
    const parsed = JSON.parse(decryptedJson) as AttendanceSensitivePayload;

    if (!parsed.action_type || !parsed.timestamp || !parsed.device_hash) {
      throw new Error('Decrypted payload missing required attendance fields.');
    }

    return parsed;
  } catch (err) {
    // Web Crypto decrypt rejects with an OperationError on tag or AAD mismatch
    throw new Error(
      `Decryption failed (integrity or AAD authentication mismatch): ${
        err instanceof Error ? err.message : String(err)
      }`,
      { cause: err }
    );
  }
}

// ==========================================
// Existing SHA-256 Tamper-Evident Hash
// ==========================================

/**
 * Generates a SHA-256 tamper-evident hash matching backend ComputeRecordHash
 * (record_uuid|user_id|action_type|timestamp).
 * Computed strictly over plaintext values before encryption.
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
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}