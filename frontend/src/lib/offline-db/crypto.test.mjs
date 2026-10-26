import test from 'node:test';
import assert from 'node:assert/strict';
import { webcrypto } from 'node:crypto';
import { atob, btoa } from 'node:buffer';
import { TextDecoder, TextEncoder } from 'node:util';

const crypto = webcrypto;

// Canonical AAD builder matching crypto.ts
function buildAttendanceAAD(recordUuid, orgId, userId) {
  return new TextEncoder().encode(`attendance-log:v1|${recordUuid}|${orgId}|${userId}`);
}

function bufferToBase64(buffer) {
  const bytes = buffer instanceof Uint8Array ? buffer : new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function base64ToBuffer(base64) {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

async function generateTestKey() {
  return await crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    true,
    ['encrypt', 'decrypt']
  );
}

async function encryptAttendancePayload(recordUuid, orgId, userId, payload, key) {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const aad = buildAttendanceAAD(recordUuid, orgId, userId);
  const plainTextBytes = new TextEncoder().encode(JSON.stringify(payload));

  const cipherBuffer = await crypto.subtle.encrypt(
    {
      name: 'AES-GCM',
      iv,
      additionalData: aad,
    },
    key,
    plainTextBytes
  );

  return {
    iv: bufferToBase64(iv),
    ciphertext: bufferToBase64(cipherBuffer),
    crypto_version: 1,
  };
}

async function decryptAttendancePayload(recordUuid, orgId, userId, ivBase64, ciphertextBase64, key) {
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
        iv,
        additionalData: aad,
      },
      key,
      ciphertext
    );

    const decryptedJson = new TextDecoder().decode(plainBuffer);
    const parsed = JSON.parse(decryptedJson);

    if (!parsed.action_type || !parsed.timestamp || !parsed.device_hash) {
      throw new Error('Decrypted payload missing required attendance fields.');
    }

    return parsed;
  } catch (err) {
    throw new Error(
      `Decryption failed: ${err instanceof Error ? err.message : String(err)}`,
      { cause: err }
    );
  }
}

async function generateDeviceHash(recordUuid, userId, actionType, timestamp) {
  const rawPayload = `${recordUuid}|${userId}|${actionType}|${timestamp}`;
  const data = new TextEncoder().encode(rawPayload);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}

test('AES-GCM: Full encrypt / decrypt round-trip with canonical AAD', async () => {
  const key = await generateTestKey();
  const recordUuid = '11111111-1111-1111-1111-111111111111';
  const orgId = 'org-addis-tech';
  const userId = 'user-aman-01';

  const nowISO = new Date().toISOString();
  const hash = await generateDeviceHash(recordUuid, userId, 'CLOCK_IN', nowISO);

  const payload = {
    action_type: 'CLOCK_IN',
    timestamp: nowISO,
    device_hash: hash,
  };

  const encrypted = await encryptAttendancePayload(recordUuid, orgId, userId, payload, key);
  assert.equal(encrypted.crypto_version, 1);
  assert.ok(encrypted.iv);
  assert.ok(encrypted.ciphertext);

  const decrypted = await decryptAttendancePayload(
    recordUuid,
    orgId,
    userId,
    encrypted.iv,
    encrypted.ciphertext,
    key
  );

  assert.equal(decrypted.action_type, 'CLOCK_IN');
  assert.equal(decrypted.timestamp, nowISO);
  assert.equal(decrypted.device_hash, hash);
});

test('AES-GCM: Fresh random IV is generated for repeated encryptions', async () => {
  const key = await generateTestKey();
  const recordUuid = '22222222-2222-2222-2222-222222222222';
  const orgId = 'org-addis-tech';
  const userId = 'user-aman-01';

  const payload = {
    action_type: 'CLOCK_OUT',
    timestamp: '2026-10-25T17:00:00Z',
    device_hash: 'test-hash',
  };

  const enc1 = await encryptAttendancePayload(recordUuid, orgId, userId, payload, key);
  const enc2 = await encryptAttendancePayload(recordUuid, orgId, userId, payload, key);

  assert.notEqual(enc1.iv, enc2.iv, 'IVs must be unique per encryption');
  assert.notEqual(enc1.ciphertext, enc2.ciphertext, 'Ciphertext must differ due to unique IV');
});

test('AES-GCM Tamper Rejection: Corrupted ciphertext fails decryption', async () => {
  const key = await generateTestKey();
  const recordUuid = '33333333-3333-3333-3333-333333333333';
  const orgId = 'org-addis-tech';
  const userId = 'user-aman-01';

  const encrypted = await encryptAttendancePayload(
    recordUuid,
    orgId,
    userId,
    { action_type: 'CLOCK_IN', timestamp: '2026-10-25T08:00:00Z', device_hash: 'hash-abc' },
    key
  );

  // Flip a byte in the ciphertext
  const bytes = base64ToBuffer(encrypted.ciphertext);
  bytes[bytes.length - 1] ^= 0xff; // flip last byte (auth tag)
  const tamperedCiphertext = bufferToBase64(bytes);

  await assert.rejects(
    async () => {
      await decryptAttendancePayload(recordUuid, orgId, userId, encrypted.iv, tamperedCiphertext, key);
    },
    /Decryption failed/,
    'Tampered ciphertext must be rejected by AES-GCM authentication tag'
  );
});

test('AES-GCM Tamper Rejection: Corrupted IV fails decryption', async () => {
  const key = await generateTestKey();
  const recordUuid = '44444444-4444-4444-4444-444444444444';
  const orgId = 'org-addis-tech';
  const userId = 'user-aman-01';

  const encrypted = await encryptAttendancePayload(
    recordUuid,
    orgId,
    userId,
    { action_type: 'CLOCK_IN', timestamp: '2026-10-25T08:00:00Z', device_hash: 'hash-abc' },
    key
  );

  // Flip a byte in the IV
  const ivBytes = base64ToBuffer(encrypted.iv);
  ivBytes[0] ^= 0x01;
  const tamperedIv = bufferToBase64(ivBytes);

  await assert.rejects(
    async () => {
      await decryptAttendancePayload(recordUuid, orgId, userId, tamperedIv, encrypted.ciphertext, key);
    },
    /Decryption failed/,
    'Tampered IV must be rejected by AES-GCM'
  );
});

test('AES-GCM AAD Tamper Rejection: Mismatched userId fails authentication', async () => {
  const key = await generateTestKey();
  const recordUuid = '55555555-5555-5555-5555-555555555555';
  const orgId = 'org-addis-tech';
  const originalUser = 'user-aman-01';
  const attackerUser = 'user-chala-02';

  const encrypted = await encryptAttendancePayload(
    recordUuid,
    orgId,
    originalUser,
    { action_type: 'CLOCK_IN', timestamp: '2026-10-25T08:00:00Z', device_hash: 'hash-abc' },
    key
  );

  await assert.rejects(
    async () => {
      // Trying to decrypt with a different user ID AAD
      await decryptAttendancePayload(recordUuid, orgId, attackerUser, encrypted.iv, encrypted.ciphertext, key);
    },
    /Decryption failed/,
    'Decryption with different userId in AAD must fail'
  );
});

test('AES-GCM AAD Tamper Rejection: Mismatched recordUuid fails authentication', async () => {
  const key = await generateTestKey();
  const recordUuid = '66666666-6666-6666-6666-666666666666';
  const swappedUuid = '77777777-7777-7777-7777-777777777777';
  const orgId = 'org-addis-tech';
  const userId = 'user-aman-01';

  const encrypted = await encryptAttendancePayload(
    recordUuid,
    orgId,
    userId,
    { action_type: 'CLOCK_IN', timestamp: '2026-10-25T08:00:00Z', device_hash: 'hash-abc' },
    key
  );

  await assert.rejects(
    async () => {
      // Trying to associate ciphertext with a different record_uuid
      await decryptAttendancePayload(swappedUuid, orgId, userId, encrypted.iv, encrypted.ciphertext, key);
    },
    /Decryption failed/,
    'Decryption with substituted record_uuid in AAD must fail'
  );
});

test('AES-GCM Invalid IV length rejection: Non-12-byte IV is rejected immediately', async () => {
  const key = await generateTestKey();
  const shortIv = bufferToBase64(new Uint8Array(8));

  await assert.rejects(
    async () => {
      await decryptAttendancePayload('uuid', 'org', 'user', shortIv, 'dummy', key);
    },
    /Invalid IV length/,
    'Non-12-byte IV must be rejected'
  );
});
