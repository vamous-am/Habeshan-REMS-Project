/**
 * Requests persistent storage from the browser using the StorageManager API.
 * This prevents automatic browser eviction of IndexedDB offline logs during storage pressure.
 */
export async function initStoragePersistence(): Promise<boolean> {
  if (typeof window === 'undefined' || !navigator?.storage?.persist) {
    console.warn('[Storage] Storage persistence API is not supported in this browser environment.');
    return false;
  }

  try {
    const isPersisted = await navigator.storage.persisted();
    if (isPersisted) {
      console.log('[Storage] Storage is already persistent.');
      return true;
    }

    const granted = await navigator.storage.persist();
    if (granted) {
      console.log('[Storage] Storage persistence granted successfully.');
    } else {
      console.warn('[Storage] Storage persistence request was denied by browser/user settings.');
    }
    return granted;
  } catch (err) {
    console.error('[Storage] Error while requesting persistent storage:', err);
    return false;
  }
}