import type { CatalogExport, DashboardData, MappingProfile, UploadPreview } from "./types";

const dbName = "anibas-feed-engine";
const dbVersion = 1;
const importLimit = 25;
const exportLimit = 25;
const profileLimit = 50;

type StoreName = "imports" | "exports" | "profiles";
type StoredMappingProfile = MappingProfile & { source_target: string };

let databasePromise: Promise<IDBDatabase> | null = null;

export async function dashboardData(): Promise<DashboardData> {
  const [imports, exports] = await Promise.all([readImports(), readExports()]);

  return {
    recent_imports: imports.map((item) => item.catalog_import).slice(0, 10),
    recent_exports: exports.slice(0, 10)
  };
}

export async function readImports(): Promise<UploadPreview[]> {
  const imports = await getAll<UploadPreview>("imports");
  return imports.sort((a, b) => b.catalog_import.id - a.catalog_import.id);
}

export async function saveImport(preview: UploadPreview): Promise<void> {
  await putRecord("imports", preview, preview.catalog_import.id);
  await trimImports();
}

export async function findImport(id: number): Promise<UploadPreview | null> {
  return (await getRecord<UploadPreview>("imports", id)) ?? null;
}

export async function removeImport(id: number): Promise<void> {
  await deleteRecord("imports", id);

  const exports = await readExports();
  await Promise.all(
    exports
      .filter((item) => item.import_id === id)
      .map((item) => deleteRecord("exports", item.id))
  );
}

export async function readExports(): Promise<CatalogExport[]> {
  const exports = await getAll<CatalogExport>("exports");
  return exports.sort((a, b) => b.id - a.id);
}

export async function saveExport(item: CatalogExport): Promise<void> {
  await putRecord("exports", item);
  await trimExports();
}

export async function readProfiles(sourcePlatform: string, targetPlatform: string): Promise<MappingProfile[]> {
  const profiles = await getAll<StoredMappingProfile>("profiles");

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

  await putRecord("profiles", stored);
  await trimProfiles();

  return saved;
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

async function getRecord<T>(storeName: StoreName, key: IDBValidKey): Promise<T | undefined> {
  const database = await openDatabase();
  const store = database.transaction(storeName, "readonly").objectStore(storeName);
  return requestResult<T | undefined>(store.get(key));
}

async function getAll<T>(storeName: StoreName): Promise<T[]> {
  const database = await openDatabase();
  const store = database.transaction(storeName, "readonly").objectStore(storeName);
  return requestResult<T[]>(store.getAll());
}

async function putRecord(storeName: StoreName, value: unknown, key?: IDBValidKey): Promise<void> {
  const database = await openDatabase();
  const store = database.transaction(storeName, "readwrite").objectStore(storeName);
  await requestResult<IDBValidKey>(key === undefined ? store.put(value) : store.put(value, key));
}

async function deleteRecord(storeName: StoreName, key: IDBValidKey): Promise<void> {
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

async function trimImports() {
  const imports = await readImports();
  await Promise.all(
    imports.slice(importLimit).map((item) => deleteRecord("imports", item.catalog_import.id))
  );
}

async function trimExports() {
  const exports = await readExports();
  await Promise.all(exports.slice(exportLimit).map((item) => deleteRecord("exports", item.id)));
}

async function trimProfiles() {
  const profiles = (await getAll<StoredMappingProfile>("profiles")).sort((a, b) => b.id - a.id);
  await Promise.all(profiles.slice(profileLimit).map((profile) => deleteRecord("profiles", profile.id)));
}

function profileLookupKey(sourcePlatform: string, targetPlatform: string) {
  return `${sourcePlatform}:${targetPlatform}`;
}

function publicProfile(profile: StoredMappingProfile): MappingProfile {
  const { source_target: _sourceTarget, ...publicItem } = profile;
  return publicItem;
}
