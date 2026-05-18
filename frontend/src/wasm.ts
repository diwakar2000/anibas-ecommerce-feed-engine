import type { FieldMappingState, SchemaField, TargetRequirement, UploadPreview } from "./types";

type WasmGlobal = {
  schema: () => string;
  parseCatalog: (csvText: string, filename: string, sourcePlatform: string) => string;
  validateCatalog: (payload: string) => string;
  exportCatalog: (payload: string) => string;
  exportFacebookCatalog: (payload: string) => string;
};

type WasmResponse<T> = {
  ok: boolean;
  data?: T;
  error?: string;
};

export type WasmValidationFinding = {
  level: "error" | "warning";
  title: string;
  detail: string;
  rowNumber?: number;
};

export type WasmExportResult = {
  filename: string;
  csv: string;
  row_count: number;
};

export type CatalogTransformPayload = {
  preview_rows: UploadPreview["preview_rows"];
  mapping: FieldMappingState;
  target_platform: string;
  source_platform: string;
};

export type AnibasWasm = {
  schema: () => { fields: SchemaField[]; target_requirements: TargetRequirement[] };
  parseCatalog: (csvText: string, filename: string, sourcePlatform: string) => UploadPreview;
  validateCatalog: (payload: CatalogTransformPayload) => { findings: WasmValidationFinding[] };
  exportCatalog: (payload: CatalogTransformPayload) => WasmExportResult;
  exportFacebookCatalog: (payload: CatalogTransformPayload) => WasmExportResult;
};

declare global {
  interface Window {
    Go?: new () => { importObject: WebAssembly.Imports; run: (instance: WebAssembly.Instance) => Promise<void> };
    anibasWasm?: WasmGlobal;
  }
}

let wasmPromise: Promise<AnibasWasm> | null = null;

export function loadAnibasWasm() {
  wasmPromise ??= startWasm();
  return wasmPromise;
}

async function startWasm(): Promise<AnibasWasm> {
  await loadScript(`${import.meta.env.BASE_URL}wasm_exec.js`);

  if (!window.Go) {
    throw new Error("Go WASM runtime did not load");
  }

  if (!window.anibasWasm) {
    const go = new window.Go();
    const wasmUrl = `${import.meta.env.BASE_URL}anibas.wasm`;
    const response = await fetch(wasmUrl);
    if (!response.ok) {
      throw new Error(`Could not load WASM engine from ${wasmUrl}`);
    }

    const bytes = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, go.importObject);
    void go.run(result.instance);
    await waitForEngine();
  }

  if (!window.anibasWasm) {
    throw new Error("Anibas WASM engine did not initialize");
  }

  const wasm = window.anibasWasm;
  return {
    schema: () => readResponse(wasm.schema()),
    parseCatalog: (csvText, filename, sourcePlatform) =>
      readResponse(wasm.parseCatalog(csvText, filename, sourcePlatform)),
    validateCatalog: (payload) => readResponse(wasm.validateCatalog(JSON.stringify(payload))),
    exportCatalog: (payload) => readResponse(wasm.exportCatalog(JSON.stringify(payload))),
    exportFacebookCatalog: (payload) => readResponse(wasm.exportFacebookCatalog(JSON.stringify(payload)))
  };
}

async function waitForEngine() {
  for (let attempt = 0; attempt < 50; attempt += 1) {
    if (window.anibasWasm) {
      return;
    }
    await new Promise((resolve) => window.setTimeout(resolve, 10));
  }
}

async function loadScript(src: string) {
  if (window.Go) {
    return;
  }

  await new Promise<void>((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>(`script[data-wasm-runtime="${src}"]`);
    if (existing) {
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener("error", () => reject(new Error("Could not load Go WASM runtime")), { once: true });
      return;
    }

    const script = document.createElement("script");
    script.src = src;
    script.async = true;
    script.dataset.wasmRuntime = src;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("Could not load Go WASM runtime"));
    document.head.append(script);
  });
}

function readResponse<T>(raw: string): T {
  const response = JSON.parse(raw) as WasmResponse<T>;
  if (!response.ok) {
    throw new Error(response.error ?? "WASM engine failed");
  }

  return response.data as T;
}
