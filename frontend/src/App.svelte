<script lang="ts">
  import Dashboard from "./components/Dashboard.svelte";
  import ExportPanel from "./components/ExportPanel.svelte";
  import FieldMapping from "./components/FieldMapping.svelte";
  import PreviewTable from "./components/PreviewTable.svelte";
  import UploadCatalog from "./components/UploadCatalog.svelte";
  import ValidationResults from "./components/ValidationResults.svelte";
  import { dashboardData, findImport, removeImport, saveExport, saveImport } from "./storage";
  import {
    fallbackSchemaFields,
    fallbackTargetRequirements,
    type CatalogExport,
    type DashboardData,
    type FieldMappingState,
    type SchemaField,
    type TargetRequirement,
    type UploadPreview
  } from "./types";
  import { loadAnibasWasm, type AnibasWasm, type WasmExportResult } from "./wasm";

  type Screen = "dashboard" | "upload" | "mapping" | "validation" | "export";

  const workflowScreens: { id: Exclude<Screen, "dashboard">; label: string }[] = [
    { id: "upload", label: "Upload" },
    { id: "mapping", label: "Field Mapping" },
    { id: "validation", label: "Validation" },
    { id: "export", label: "Export" }
  ];
  const logoUrl = `${import.meta.env.BASE_URL}anibas-logo.svg`;

  let activeScreen = $state<Screen>("dashboard");
  let apiError = $state("");
  let storageError = $state("");
  let dashboard = $state<DashboardData>({
    recent_imports: [],
    recent_exports: []
  });
  let schemaFields = $state<SchemaField[]>(fallbackSchemaFields);
  let targetRequirements = $state<TargetRequirement[]>(fallbackTargetRequirements);
  let selectedTarget = $state("facebook_catalog_csv");
  let uploadPreview = $state<UploadPreview | null>(null);
  let deletingImportId = $state<number | null>(null);
  let openingImportId = $state<number | null>(null);
  let fieldMapping = $state<FieldMappingState>({});
  let wasmEngine = $state<AnibasWasm | null>(null);
  let loadingEngine = $state(true);
  let validationErrorCount = $state<number | null>(null);
  let initialized = false;

  async function loadDashboard() {
    try {
      dashboard = await dashboardData();
      storageError = "";
    } catch (err) {
      storageError = err instanceof Error ? err.message : "Browser storage is unavailable";
      dashboard = {
        recent_imports: [],
        recent_exports: []
      };
    }
  }

  async function initializeApp() {
    await loadDashboard();
    loadingEngine = true;
    try {
      const engine = await loadAnibasWasm();
      wasmEngine = engine;
      const payload = engine.schema();
      schemaFields = payload.fields ?? fallbackSchemaFields;
      targetRequirements = normalizeTargetRequirements(payload.target_requirements);
      apiError = "";
    } catch (err) {
      schemaFields = fallbackSchemaFields;
      targetRequirements = normalizeTargetRequirements(fallbackTargetRequirements);
      apiError = err instanceof Error ? err.message : "WASM engine unavailable";
    } finally {
      loadingEngine = false;
    }
  }

  async function handleUploaded(result: UploadPreview) {
    uploadPreview = result;
    fieldMapping = {};
    validationErrorCount = null;
    activeScreen = "mapping";
    await loadDashboard();
  }

  async function parseCatalogFile(file: File, sourcePlatform: string) {
    if (!wasmEngine) {
      throw new Error("Catalog engine is still loading");
    }

    const text = await file.text();
    const result = wasmEngine.parseCatalog(text, file.name, sourcePlatform);
    await saveImport(result);
    await handleUploaded(result);
  }

  async function openImport(id: number) {
    openingImportId = id;
    try {
      const payload = await findImport(id);
      if (!payload?.columns?.length) {
        throw new Error("This import only has metadata. Upload the file again to make it reusable.");
      }

      uploadPreview = payload;
      fieldMapping = {};
      validationErrorCount = null;
      activeScreen = "mapping";
      apiError = "";
    } catch (err) {
      apiError = err instanceof Error ? err.message : "Could not open catalog import";
    } finally {
      openingImportId = null;
    }
  }

  async function deleteImport(id: number) {
    if (!window.confirm("Delete this catalog import?")) {
      return;
    }

    deletingImportId = id;
    try {
      await removeImport(id);
      await loadDashboard();

      if (uploadPreview?.catalog_import.id === id) {
        uploadPreview = null;
        fieldMapping = {};
        activeScreen = "dashboard";
      }

      apiError = "";
    } catch (err) {
      apiError = err instanceof Error ? err.message : "Could not delete catalog import";
    } finally {
      deletingImportId = null;
    }
  }

  async function handleExportGenerated(result: WasmExportResult) {
    if (!uploadPreview) {
      return;
    }

    const exportRecord: CatalogExport = {
      id: Date.now(),
      import_id: uploadPreview.catalog_import.id,
      target_platform: selectedTarget,
      filename: result.filename,
      row_count: result.row_count,
      status: "generated",
      created_at: new Date().toISOString()
    };
    await saveExport(exportRecord);
    await loadDashboard();
  }

  function normalizeTargetRequirements(value: TargetRequirement[] | null | undefined): TargetRequirement[] {
    const requirements = value && value.length > 0 ? value : fallbackTargetRequirements;

    return requirements.map((requirement) => ({
      ...requirement,
      required_fields: requirement.required_fields ?? [],
      conditional_fields: requirement.conditional_fields ?? [],
      recommended_fields: requirement.recommended_fields ?? [],
      requirement_groups: requirement.requirement_groups ?? []
    }));
  }

  function canOpenScreen(screen: Screen) {
    if (screen === "dashboard" || screen === "upload") {
      return true;
    }

    return Boolean(uploadPreview);
  }

  function isStepReady(screen: Exclude<Screen, "dashboard">) {
    if (screen === "upload") {
      return Boolean(uploadPreview);
    }

    return false;
  }

  function workflowIndex(screen: Screen) {
    return workflowScreens.findIndex((item) => item.id === screen);
  }

  function activeWorkflowLabel() {
    return workflowScreens.find((item) => item.id === activeScreen)?.label ?? "Catalog Workflow";
  }

  function activeWorkflowHelp() {
    const help: Record<Exclude<Screen, "dashboard">, string> = {
      upload: "Import a catalog file and inspect detected columns before mapping.",
      mapping: "Map source columns or static values into the Universal Product Schema.",
      validation: "Review channel-specific errors and warnings before export.",
      export: "Generate a target-ready Facebook Catalog CSV from the current mapping."
    };

    return activeScreen === "dashboard" ? "" : help[activeScreen];
  }

  function previousScreen() {
    if (activeScreen === "upload") {
      return "dashboard";
    }

    const index = workflowIndex(activeScreen);
    return index > 0 ? workflowScreens[index - 1].id : "dashboard";
  }

  function nextScreen() {
    const index = workflowIndex(activeScreen);
    if (index < 0 || index >= workflowScreens.length - 1) {
      return null;
    }

    return workflowScreens[index + 1].id;
  }

  function canAdvance() {
    if (activeScreen === "upload") {
      return Boolean(uploadPreview);
    }
    if (activeScreen === "mapping") {
      return Boolean(uploadPreview?.columns.length);
    }
    if (activeScreen === "validation") {
      return Boolean(uploadPreview) && validationErrorCount === 0;
    }

    return false;
  }

  function nextBlockedReason() {
    if (activeScreen === "upload") {
      return "Parse a catalog file to continue.";
    }
    if (activeScreen === "mapping") {
      return "Open or upload a catalog before validation.";
    }
    if (activeScreen === "validation") {
      return validationErrorCount === null
        ? "Validation is calculating."
        : "Resolve validation errors to continue.";
    }

    return "";
  }

  function goBack() {
    activeScreen = previousScreen();
  }

  function goNext() {
    const next = nextScreen();
    if (next && canAdvance()) {
      activeScreen = next;
    }
  }

  $effect(() => {
    if (initialized) {
      return;
    }

    initialized = true;
    initializeApp().catch((err) => {
      apiError = err instanceof Error ? err.message : "Could not initialize Anibas Feed Engine";
      loadingEngine = false;
    });
  });
</script>

<div class="brand-backdrop" aria-hidden="true">
  <span class="logo-cloud tiktok">TikTok Shop</span>
  <span class="logo-cloud meta">Meta</span>
  <span class="logo-cloud instagram">Instagram</span>
  <span class="logo-cloud shopify">Shopify</span>
  <span class="logo-cloud google">Google Merchant</span>
  <span class="logo-cloud woo">WooCommerce</span>
</div>

<main class="shell">
  <header class="app-header">
    <button class="brand-lockup" type="button" onclick={() => (activeScreen = "dashboard")}>
      <img src={logoUrl} alt="" />
      <span>
        <strong>Anibas Feed Engine</strong>
        <small>Catalog transformation</small>
      </span>
    </button>

    <div class="app-nav">
      <button
        type="button"
        class:active={activeScreen === "dashboard"}
        onclick={() => (activeScreen = "dashboard")}
      >
        Dashboard
      </button>
      <button
        type="button"
        class:active={activeScreen !== "dashboard"}
        onclick={() => (activeScreen = "upload")}
      >
        Catalog Workflow
      </button>
    </div>

    <div class="status" class:error={Boolean(apiError)}>
      <span></span>
      {apiError ? "WASM offline" : loadingEngine ? "WASM loading" : "WASM ready"}
    </div>
  </header>

  {#if apiError || storageError}
    <p class="alert">{apiError || storageError}</p>
  {/if}

  {#if activeScreen === "dashboard"}
    <Dashboard
      dashboard={dashboard}
      deletingImportId={deletingImportId}
      openingImportId={openingImportId}
      onUploadClick={() => (activeScreen = "upload")}
      onOpenImport={openImport}
      onDeleteImport={deleteImport}
    />
  {:else}
    <section class="workflow-shell">
      <nav class="workflow-nav" aria-label="Catalog workflow">
        {#each workflowScreens as screen}
          {@const locked = !canOpenScreen(screen.id)}
          <button
            type="button"
            class:active={activeScreen === screen.id}
            class:ready={isStepReady(screen.id)}
            disabled={locked}
            onclick={() => {
              if (!locked) {
                activeScreen = screen.id;
              }
            }}
          >
            <strong>{screen.label}</strong>
          </button>
        {/each}
      </nav>

      <div class="content-toolbar">
        <div>
          <p>Catalog workflow</p>
          <h1>{activeWorkflowLabel()}</h1>
          <span>{activeWorkflowHelp()}</span>
        </div>
        <div class="toolbar-actions">
          <button class="secondary" type="button" onclick={goBack}>
            {activeScreen === "upload" ? "Back to Dashboard" : "Back"}
          </button>
          {#if nextScreen()}
            <span>
              <button type="button" disabled={!canAdvance()} title={canAdvance() ? "" : nextBlockedReason()} onclick={goNext}>
                Next
              </button>
              {#if !canAdvance()}
                <small>{nextBlockedReason()}</small>
              {/if}
            </span>
          {/if}
        </div>
      </div>
    </section>
  {/if}

  {#if activeScreen === "upload"}
    <section class="upload-grid">
      <UploadCatalog preview={uploadPreview} disabled={!wasmEngine} onUploadFile={parseCatalogFile} />
      <PreviewTable preview={uploadPreview} />
    </section>
  {:else if activeScreen === "mapping"}
    <FieldMapping
      columns={uploadPreview?.columns ?? []}
      fields={schemaFields}
      mapping={fieldMapping}
      previewRows={uploadPreview?.preview_rows ?? []}
      sourcePlatform={uploadPreview?.catalog_import.source_platform ?? "generic_csv"}
      suggestions={uploadPreview?.mapping_suggestions ?? []}
      targetRequirements={targetRequirements}
      selectedTarget={selectedTarget}
      onTargetChange={(target) => (selectedTarget = target)}
      onMappingChange={(nextMapping) => {
        fieldMapping = nextMapping;
        validationErrorCount = null;
      }}
      onBack={() => (activeScreen = "upload")}
      onContinue={() => (activeScreen = "validation")}
    />
  {:else if activeScreen === "validation"}
    <ValidationResults
      preview={uploadPreview}
      fields={schemaFields}
      mapping={fieldMapping}
      targetRequirements={targetRequirements}
      selectedTarget={selectedTarget}
      wasmEngine={wasmEngine}
      onStatusChange={(errorCount) => (validationErrorCount = errorCount)}
      onBack={() => (activeScreen = "mapping")}
      onContinue={() => (activeScreen = "export")}
    />
  {:else if activeScreen === "export"}
    <ExportPanel
      preview={uploadPreview}
      mapping={fieldMapping}
      selectedTarget={selectedTarget}
      targetRequirements={targetRequirements}
      wasmEngine={wasmEngine}
      onTargetChange={(target) => {
        selectedTarget = target;
        validationErrorCount = null;
      }}
      onBack={() => (activeScreen = "validation")}
      onGenerated={handleExportGenerated}
    />
  {/if}
</main>

<style>
  :global(body) {
    overflow-x: hidden;
    background:
      radial-gradient(circle at top left, rgb(21 107 96 / 0.22), transparent 34rem),
      radial-gradient(circle at 88% 12%, rgb(231 88 127 / 0.18), transparent 28rem),
      linear-gradient(135deg, #f3fbf9 0%, #f6f4ff 44%, #fff8f2 100%);
  }

  .brand-backdrop {
    position: fixed;
    inset: 0;
    z-index: -1;
    overflow: hidden;
    pointer-events: none;
  }

  .brand-backdrop::before {
    position: absolute;
    inset: 8% 4% auto;
    height: 460px;
    border-radius: 999px;
    background:
      radial-gradient(circle at 22% 45%, rgb(20 184 166 / 0.28), transparent 14rem),
      radial-gradient(circle at 47% 42%, rgb(236 72 153 / 0.24), transparent 14rem),
      radial-gradient(circle at 70% 48%, rgb(59 130 246 / 0.22), transparent 15rem),
      radial-gradient(circle at 82% 36%, rgb(34 197 94 / 0.2), transparent 12rem);
    content: "";
    filter: blur(24px);
    opacity: 0.9;
  }

  .logo-cloud {
    position: absolute;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 126px;
    min-height: 126px;
    border: 1px solid rgb(255 255 255 / 0.46);
    border-radius: 999px;
    color: rgb(15 23 42 / 0.72);
    font-size: 0.8rem;
    font-weight: 900;
    text-align: center;
    box-shadow: 0 26px 70px rgb(25 39 52 / 0.13);
    filter: blur(0.25px);
    mix-blend-mode: multiply;
    padding: 18px;
  }

  .logo-cloud.tiktok {
    top: 88px;
    left: 5%;
    background: linear-gradient(135deg, rgb(14 165 233 / 0.26), rgb(244 63 94 / 0.24));
    transform: rotate(-12deg);
  }

  .logo-cloud.meta {
    top: 34px;
    left: 24%;
    background: linear-gradient(135deg, rgb(37 99 235 / 0.25), rgb(20 184 166 / 0.22));
    transform: rotate(8deg);
  }

  .logo-cloud.instagram {
    top: 132px;
    right: 22%;
    background: linear-gradient(135deg, rgb(236 72 153 / 0.25), rgb(251 146 60 / 0.24));
    transform: rotate(13deg);
  }

  .logo-cloud.shopify {
    top: 62px;
    right: 6%;
    background: linear-gradient(135deg, rgb(34 197 94 / 0.25), rgb(132 204 22 / 0.18));
    transform: rotate(-9deg);
  }

  .logo-cloud.google {
    top: 275px;
    left: 15%;
    background: linear-gradient(135deg, rgb(250 204 21 / 0.22), rgb(59 130 246 / 0.18));
    transform: rotate(10deg);
  }

  .logo-cloud.woo {
    top: 292px;
    right: 12%;
    background: linear-gradient(135deg, rgb(124 58 237 / 0.18), rgb(236 72 153 / 0.18));
    transform: rotate(-11deg);
  }

  .shell {
    width: min(1180px, calc(100% - 32px));
    margin: 0 auto;
    padding: 22px 0 48px;
  }

  .app-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 22px;
    margin-bottom: 22px;
  }

  .brand-lockup {
    display: inline-flex;
    align-items: center;
    gap: 12px;
    min-height: 0;
    border: 0;
    background: transparent;
    box-shadow: none;
    color: #13251c;
    padding: 0;
    text-align: left;
  }

  .brand-lockup:hover {
    box-shadow: none;
    transform: none;
  }

  .brand-lockup img {
    width: 46px;
    height: 46px;
    flex: 0 0 auto;
  }

  .brand-lockup span {
    display: grid;
    gap: 2px;
    min-width: 0;
  }

  .brand-lockup strong {
    color: #13251c;
    font-size: 1rem;
    line-height: 1.1;
  }

  .brand-lockup small,
  .content-toolbar p,
  .content-toolbar span {
    color: #52645f;
  }

  .app-nav {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 6px;
    justify-content: center;
  }

  .app-nav button {
    min-height: 36px;
    border-radius: 999px;
    background: transparent;
    box-shadow: none;
    color: #40564d;
    padding: 0 12px;
  }

  .app-nav button:hover {
    background: rgb(255 255 255 / 0.48);
    box-shadow: none;
  }

  .app-nav button.active {
    background: #183f31;
    color: #ffffff;
  }

  h1,
  .alert {
    margin: 0;
  }

  h1 {
    color: #13251f;
    font-size: clamp(1.45rem, 2.6vw, 2rem);
    line-height: 1.1;
  }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    min-height: 38px;
    border: 1px solid rgb(255 255 255 / 0.62);
    border-radius: 999px;
    background: rgb(255 255 255 / 0.7);
    color: #244f3b;
    font-weight: 800;
    padding: 0 12px;
    white-space: nowrap;
  }

  .status span {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: #2f8d5c;
  }

  .status.error {
    color: #9b3324;
  }

  .status.error span {
    background: #c94b37;
  }

  .workflow-shell {
    display: grid;
    gap: 16px;
    margin-bottom: 18px;
  }

  .workflow-nav {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    align-items: stretch;
    gap: 0;
  }

  .workflow-nav button {
    position: relative;
    display: grid;
    place-items: center;
    min-height: 44px;
    border: 0;
    border-radius: 0;
    background: rgb(255 255 255 / 0.58);
    color: #30423a;
    box-shadow: none;
    padding: 0 18px 0 26px;
    transition:
      background 160ms ease,
      color 160ms ease,
      opacity 160ms ease;
  }

  .workflow-nav button::after {
    position: absolute;
    top: 0;
    right: -22px;
    z-index: 2;
    width: 44px;
    height: 44px;
    background: inherit;
    clip-path: polygon(0 0, 52% 0, 100% 50%, 52% 100%, 0 100%, 48% 50%);
    content: "";
  }

  .workflow-nav button:first-child {
    border-radius: 999px 0 0 999px;
    padding-left: 18px;
  }

  .workflow-nav button:last-child {
    border-radius: 0 999px 999px 0;
  }

  .workflow-nav button:last-child::after {
    display: none;
  }

  .workflow-nav button:not(:disabled):hover {
    background: rgb(255 255 255 / 0.78);
    box-shadow: none;
    transform: none;
  }

  .workflow-nav button strong {
    font-size: 0.9rem;
    line-height: 1;
    white-space: nowrap;
  }

  .workflow-nav button.active {
    background: linear-gradient(135deg, #1d563b, #287357);
    color: #fff;
  }

  .workflow-nav button:disabled {
    color: #7a8983;
    opacity: 0.7;
  }

  .content-toolbar {
    display: flex;
    align-items: end;
    justify-content: space-between;
    gap: 18px;
  }

  .content-toolbar > div:first-child {
    display: grid;
    gap: 5px;
    min-width: 0;
  }

  .content-toolbar p {
    margin: 0;
    font-size: 0.78rem;
    font-weight: 900;
    letter-spacing: 0;
    text-transform: uppercase;
  }

  .content-toolbar span {
    max-width: 64ch;
    line-height: 1.45;
    overflow-wrap: anywhere;
  }

  .toolbar-actions {
    display: flex;
    flex: 0 0 auto;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: 10px;
    justify-content: flex-end;
  }

  .toolbar-actions span {
    display: grid;
    gap: 5px;
    justify-items: end;
  }

  .toolbar-actions button {
    min-width: 112px;
  }

  .toolbar-actions small {
    max-width: 220px;
    color: #62716a;
    font-size: 0.75rem;
    font-weight: 700;
    line-height: 1.3;
    text-align: right;
  }

  .secondary {
    border: 1px solid rgb(194 210 204 / 0.9);
    background: rgb(255 255 255 / 0.62);
    color: #244f3b;
  }

  .alert {
    border: 1px solid rgb(241 183 168 / 0.9);
    border-color: rgb(241 183 168 / 0.9);
    border-radius: 14px;
    background: rgb(255 246 242 / 0.7);
    color: #9b3324;
    font-weight: 700;
    margin-bottom: 16px;
    padding: 10px 12px;
  }

  .upload-grid {
    display: grid;
    grid-template-columns: minmax(280px, 360px) minmax(0, 1fr);
    gap: 18px;
    align-items: start;
  }

  @media (max-width: 860px) {
    .app-header,
    .content-toolbar,
    .upload-grid {
      display: grid;
      align-items: start;
    }

    .app-nav {
      justify-content: start;
    }

    .status {
      width: fit-content;
    }

    .logo-cloud {
      opacity: 0.55;
    }
  }

  @media (max-width: 540px) {
    .shell {
      width: min(100% - 22px, 1180px);
      padding-top: 18px;
    }

    .brand-lockup img {
      width: 40px;
      height: 40px;
    }

    .workflow-nav {
      grid-template-columns: 1fr;
      gap: 8px;
    }

    .workflow-nav button {
      border-radius: 999px;
      justify-content: center;
      padding: 0 14px;
    }

    .workflow-nav button::after {
      display: none;
    }

    .toolbar-actions,
    .toolbar-actions span,
    .toolbar-actions button {
      width: 100%;
    }

    .toolbar-actions span,
    .toolbar-actions small {
      justify-items: stretch;
      max-width: none;
      text-align: left;
    }
  }
</style>
