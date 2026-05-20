import type { CatalogExport, DashboardData, MappingProfile, UploadPreview } from "./types";

const dbName = "anibas-feed-engine";
const dbVersion = 1;
const importLimit = 25;
const exportLimit = 25;
const profileLimit = 50;
const importsKey = "anibas.imports.v1";
const exportsKey = "anibas.exports.v1";
const profilesKey = "anibas.mapping-profiles.v1";
const storagePreferenceKey = "anibas.storage-mode.v1";

type StoreName = "imports" | "exports" | "profiles";
type StoredMappingProfile = MappingProfile & { source_target: string };

export type StorageMode = "indexeddb" | "localstorage";
export type StorageAccessStatus = {
  state: PermissionState | "unsupported";
  recommendedMode: StorageMode | null;
  requiresPrompt: boolean;
  message: string;
};

let databasePromise: Promise<IDBDatabase> | null = null;
let activeMode: StorageMode | null = null;

export async function storageAccessStatus(): Promise<StorageAccessStatus> {
  const savedPreference = readStoragePreference();
  if (savedPreference) {
    return {
      state: savedPreference === "indexeddb" ? "granted" : "denied",
      recommendedMode: savedPreference,
      requiresPrompt: false,
      message: ""
    };
  }

  if (!("indexedDB" in window)) {
    return {
      state: "unsupported",
      recommendedMode: "localstorage",
      requiresPrompt: false,
      message: ""
    };
  }

  const state = await persistentStoragePermission();
  if (state === "granted") {
    return {
      state,
      recommendedMode: "indexeddb",
      requiresPrompt: false,
      message: ""
    };
  }

  if (state === "denied") {
    return {
      state,
      recommendedMode: "localstorage",
      requiresPrompt: false,
      message: ""
    };
  }

  return {
    state,
    recommendedMode: null,
    requiresPrompt: true,
    message: "Choose how Anibas should save imports and mapping profiles in this browser."
  };
}

export async function requestPersistentStorage(): Promise<boolean> {
  if (!("indexedDB" in window) || !("storage" in navigator) || !navigator.storage.persist) {
    return false;
  }

  try {
    return await navigator.storage.persist();
  } catch {
    return false;
  }
}

export function configureStorage(mode: StorageMode) {
  activeMode = mode;
  writeStoragePreference(mode);
}

export function currentStorageMode() {
  return activeMode;
}

function readStoragePreference(): StorageMode | null {
  try {
    const value = window.localStorage.getItem(storagePreferenceKey);
    return value === "indexeddb" || value === "localstorage" ? value : null;
  } catch {
    return null;
  }
}

function writeStoragePreference(mode: StorageMode) {
  try {
    window.localStorage.setItem(storagePreferenceKey, mode);
  } catch {
    // If preference storage is unavailable, the active in-memory mode still lets this session proceed.
  }
}

export async function dashboardData(): Promise<DashboardData> {
  const [imports, exports] = await Promise.all([readImports(), readExports()]);

  return {
    recent_imports: imports.map((item) => item.catalog_import).slice(0, 10),
    recent_exports: exports.slice(0, 10)
  };
}

export async function readImports(): Promise<UploadPreview[]> {
  if (storageMode() === "localstorage") {
    return readLocalJSON<UploadPreview[]>(importsKey, []).sort(
      (a, b) => b.catalog_import.id - a.catalog_import.id
    );
  }

  const imports = await idbGetAll<UploadPreview>("imports");
  return imports.sort((a, b) => b.catalog_import.id - a.catalog_import.id);
}

export async function saveImport(preview: UploadPreview): Promise<void> {
  if (storageMode() === "localstorage") {
    const imports = (await readImports()).filter((item) => item.catalog_import.id !== preview.catalog_import.id);
    imports.unshift(preview);
    writeLocalJSON(importsKey, imports.slice(0, importLimit));
    return;
  }

  await idbPutRecord("imports", preview, preview.catalog_import.id);
  await trimImports();
}

export async function findImport(id: number): Promise<UploadPreview | null> {
  if (storageMode() === "localstorage") {
    return (await readImports()).find((item) => item.catalog_import.id === id) ?? null;
  }

  return (await idbGetRecord<UploadPreview>("imports", id)) ?? null;
}

export async function removeImport(id: number): Promise<void> {
  if (storageMode() === "localstorage") {
    writeLocalJSON(
      importsKey,
      (await readImports()).filter((item) => item.catalog_import.id !== id)
    );
    writeLocalJSON(
      exportsKey,
      (await readExports()).filter((item) => item.import_id !== id)
    );
    return;
  }

  await idbDeleteRecord("imports", id);

  const exports = await readExports();
  await Promise.all(
    exports
      .filter((item) => item.import_id === id)
      .map((item) => idbDeleteRecord("exports", item.id))
  );
}

export async function readExports(): Promise<CatalogExport[]> {
  if (storageMode() === "localstorage") {
    return readLocalJSON<CatalogExport[]>(exportsKey, []).sort((a, b) => b.id - a.id);
  }

  const exports = await idbGetAll<CatalogExport>("exports");
  return exports.sort((a, b) => b.id - a.id);
}

export async function saveExport(item: CatalogExport): Promise<void> {
  if (storageMode() === "localstorage") {
    const exports = (await readExports()).filter((existing) => existing.id !== item.id);
    exports.unshift(item);
    writeLocalJSON(exportsKey, exports.slice(0, exportLimit));
    return;
  }

  await idbPutRecord("exports", item);
  await trimExports();
}

export async function readProfiles(sourcePlatform: string, targetPlatform: string): Promise<MappingProfile[]> {
  if (storageMode() === "localstorage") {
    return readLocalJSON<StoredMappingProfile[]>(profilesKey, [])
      .filter((profile) => profile.source_target === profileLookupKey(sourcePlatform, targetPlatform))
      .sort((a, b) => b.id - a.id)
      .map(publicProfile);
  }

  const profiles = await idbGetAll<StoredMappingProfile>("profiles");

  return profiles
    .filter((profile) => profile.source_target === profileLookupKey(sourcePlatform, targetPlatform))
    .sort((a, b) => b.id - a.id)
    .map(publicProfile);
}

export async function saveProfile(
  profile: Omit<MappingProfile, "id" | "created_at" | "updated_at">
): Promise<MappingProfile> {
  const now = new Date().toISOString();
  const saved: MappingProfile = {
    ...profile,
    id: Date.now(),
    created_at: now,
    updated_at: now
  };

  const stored: StoredMappingProfile = {
    ...saved,
    source_target: profileLookupKey(saved.source_platform, saved.target_platform)
  };

  if (storageMode() === "localstorage") {
    const profiles = readLocalJSON<StoredMappingProfile[]>(profilesKey, []).filter(
      (item) => item.id !== stored.id
    );
    writeLocalJSON(profilesKey, [stored, ...profiles].slice(0, profileLimit));
    return saved;
  }

  await idbPutRecord("profiles", stored);
  await trimProfiles();

  return saved;
}

async function persistentStoragePermission(): Promise<PermissionState | "unsupported"> {
  if (!navigator.permissions?.query) {
    return "unsupported";
  }

  try {
    const status = await navigator.permissions.query({
      name: "persistent-storage" as PermissionName
    });
    return status.state;
  } catch {
    return "unsupported";
  }
}

function storageMode(): StorageMode {
  if (!activeMode) {
    throw new Error("Choose browser storage before saving catalog data.");
  }

  return activeMode;
}

function openDatabase(): Promise<IDBDatabase> {
  if (databasePromise) {
    return databasePromise;
  }

  if (!("indexedDB" in window)) {
    return Promise.reject(new Error("IndexedDB is unavailable in this browser."));
  }

  databasePromise = new Promise((resolve, reject) => {
    const request = window.indexedDB.open(dbName, dbVersion);

    request.onupgradeneeded = () => {
      const database = request.result;

      if (!database.objectStoreNames.contains("imports")) {
        database.createObjectStore("imports");
      }

      if (!database.objectStoreNames.contains("exports")) {
        const exportsStore = database.createObjectStore("exports", { keyPath: "id" });
        exportsStore.createIndex("import_id", "import_id", { unique: false });
      }

      if (!database.objectStoreNames.contains("profiles")) {
        const profilesStore = database.createObjectStore("profiles", { keyPath: "id" });
        profilesStore.createIndex("source_target", "source_target", { unique: false });
      }
    };

    request.onerror = () => {
      databasePromise = null;
      reject(new Error(request.error?.message ?? "Could not open IndexedDB."));
    };

    request.onsuccess = () => {
      const database = request.result;
      database.onversionchange = () => database.close();
      resolve(database);
    };
  });

  return databasePromise;
}

async function idbGetRecord<T>(storeName: StoreName, key: IDBValidKey): Promise<T | undefined> {
  const database = await openDatabase();
  const store = database.transaction(storeName, "readonly").objectStore(storeName);
  return requestResult<T | undefined>(store.get(key));
}

async function idbGetAll<T>(storeName: StoreName): Promise<T[]> {
  const database = await openDatabase();
  const store = database.transaction(storeName, "readonly").objectStore(storeName);
  return requestResult<T[]>(store.getAll());
}

async function idbPutRecord(storeName: StoreName, value: unknown, key?: IDBValidKey): Promise<void> {
  const database = await openDatabase();
  const store = database.transaction(storeName, "readwrite").objectStore(storeName);
  await requestResult<IDBValidKey>(key === undefined ? store.put(value) : store.put(value, key));
}

async function idbDeleteRecord(storeName: StoreName, key: IDBValidKey): Promise<void> {
  const database = await openDatabase();
  const store = database.transaction(storeName, "readwrite").objectStore(storeName);
  await requestResult<undefined>(store.delete(key));
}

function requestResult<T>(request: IDBRequest): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onerror = () => reject(new Error(request.error?.message ?? "IndexedDB request failed."));
    request.onsuccess = () => resolve(request.result as T);
  });
}

function readLocalJSON<T>(key: string, fallback: T): T {
  try {
    const raw = window.localStorage.getItem(key);
    return raw ? (JSON.parse(raw) as T) : fallback;
  } catch {
    return fallback;
  }
}

function writeLocalJSON(key: string, value: unknown) {
  try {
    window.localStorage.setItem(key, JSON.stringify(value));
  } catch (err) {
    throw new Error(err instanceof Error ? err.message : "Could not save data to local browser storage.");
  }
}

async function trimImports() {
  const imports = await readImports();
  await Promise.all(
    imports.slice(importLimit).map((item) => idbDeleteRecord("imports", item.catalog_import.id))
  );
}

async function trimExports() {
  const exports = await readExports();
  await Promise.all(exports.slice(exportLimit).map((item) => idbDeleteRecord("exports", item.id)));
}

async function trimProfiles() {
  const profiles = (await idbGetAll<StoredMappingProfile>("profiles")).sort((a, b) => b.id - a.id);
  await Promise.all(profiles.slice(profileLimit).map((profile) => idbDeleteRecord("profiles", profile.id)));
}

function profileLookupKey(sourcePlatform: string, targetPlatform: string) {
  return `${sourcePlatform}:${targetPlatform}`;
}

function publicProfile(profile: StoredMappingProfile): MappingProfile {
  const { source_target: _sourceTarget, ...publicItem } = profile;
  return publicItem;
}
