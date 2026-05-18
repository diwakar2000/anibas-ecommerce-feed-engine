<script lang="ts">
  import { mappingHasValue, resolveMappedValue } from "../mapping";
  import type { FieldMappingState, TargetRequirement, UploadPreview } from "../types";
  import type { AnibasWasm, WasmExportResult } from "../wasm";

  type Props = {
    preview: UploadPreview | null;
    mapping: FieldMappingState;
    selectedTarget: string;
    targetRequirements: TargetRequirement[];
    wasmEngine: AnibasWasm | null;
    onTargetChange: (target: string) => void;
    onBack: () => void;
    onGenerated: (result: WasmExportResult) => void;
  };

  let {
    preview,
    mapping,
    selectedTarget,
    targetRequirements,
    wasmEngine,
    onTargetChange,
    onBack,
    onGenerated
  }: Props = $props();
  let exportMessage = $state("");
  let target = $derived(selectedTarget);
  let selectedRequirement = $derived(
    targetRequirements.find((requirement) => requirement.id === target)
  );
  let blockers = $derived(exportBlockers(preview, mapping, selectedRequirement, target, wasmEngine));
  let canGenerate = $derived(blockers.length === 0);

  function exportBlockers(
    currentPreview: UploadPreview | null,
    currentMapping: FieldMappingState,
    requirement: TargetRequirement | undefined,
    currentTarget: string,
    engine: AnibasWasm | null
  ) {
    const result: string[] = [];
    if (!currentPreview) {
      result.push("Upload or open a catalog import first.");
      return result;
    }
    if (!engine) {
      result.push("Catalog engine is still loading.");
    }
    if (!requirement) {
      result.push("Choose a target format.");
      return result;
    }

    for (const field of requirement.required_fields) {
      if (field.field === "currency" && canDeriveCurrency(currentPreview, currentMapping)) {
        continue;
      }
      if (!mappingHasValue(currentMapping, field.field)) {
        result.push(`Map ${fieldLabel(field.field)} or provide a static value.`);
      }
    }

    for (const group of requirement.requirement_groups.filter((item) => item.level === "required")) {
      const mappedCount = group.fields.filter((field) => mappingHasValue(currentMapping, field)).length;
      if (mappedCount < group.min) {
        result.push(`Map at least ${group.min} of ${group.fields.map(fieldLabel).join(", ")}.`);
      }
    }

    if (currentPreview.preview_rows.length === 0) {
      result.push("No preview rows are available to export.");
    }

    return result;
  }

  function canDeriveCurrency(currentPreview: UploadPreview, currentMapping: FieldMappingState) {
    if (!mappingHasValue(currentMapping, "price")) {
      return false;
    }

    return currentPreview.preview_rows.some((row) => resolveMappedValue(currentMapping, "currency", row.values));
  }

  function generateExport() {
    if (!preview || !canGenerate || !wasmEngine) {
      return;
    }

    const result = wasmEngine.exportCatalog({
      preview_rows: preview.preview_rows,
      mapping,
      target_platform: target,
      source_platform: preview.catalog_import.source_platform
    });
    const url = URL.createObjectURL(new Blob([result.csv], { type: "text/csv" }));
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = result.filename;
    anchor.click();
    URL.revokeObjectURL(url);
    exportMessage = `Generated ${result.row_count} ${targetLabel(target).toLowerCase()} preview rows. Full-catalog export will use the same mapping once full-row persistence is wired.`;
    onGenerated(result);
  }

  function targetLabel(targetId: string) {
    return targetRequirements.find((requirement) => requirement.id === targetId)?.label ?? targetId;
  }

  function targetFilename(targetId: string) {
    const filenames: Record<string, string> = {
      facebook_catalog_csv: "facebook_catalog_preview.csv",
      instagram_shops: "instagram_shops_preview.csv",
      google_merchant_center: "google_merchant_center_preview.csv",
      tiktok_catalog: "tiktok_catalog_preview.csv",
      shopify_csv: "shopify_products_preview.csv",
      woocommerce_csv: "woocommerce_products_preview.csv"
    };

    return filenames[targetId] ?? `${targetId}_preview.csv`;
  }

  function fieldLabel(field: string) {
    const labels: Record<string, string> = {
      id: "ID",
      sku: "SKU",
      title: "Title",
      description: "Description",
      availability: "Availability",
      condition: "Condition",
      price: "Price",
      currency: "Currency",
      product_url: "Product URL",
      image_url: "Image URL",
      brand: "Brand",
      gtin: "GTIN",
      mpn: "MPN",
      category: "Category",
      quantity: "Quantity"
    };

    return labels[field] ?? field;
  }
</script>

<section class="panel">
  <div class="panel-heading">
    <div>
      <h2>Export</h2>
      <p>{preview ? `${preview.row_count} rows staged` : "No catalog staged"}</p>
    </div>
    <span>CSV</span>
  </div>

  <label>
    Target format
    <select value={target} onchange={(event) => onTargetChange(event.currentTarget.value)}>
      {#each targetRequirements as targetRequirement}
        <option value={targetRequirement.id}>{targetRequirement.label}</option>
      {/each}
    </select>
  </label>

  {#if blockers.length > 0}
    <div class="blockers">
      <strong>Export is waiting for:</strong>
      <ul>
        {#each blockers as blocker}
          <li>{blocker}</li>
        {/each}
      </ul>
    </div>
  {:else}
    <div class="ready-state">
      <strong>Ready to generate {targetLabel(target)}</strong>
      <span>{preview?.preview_rows.length ?? 0} staged preview rows will be included.</span>
    </div>
  {/if}

  {#if exportMessage}
    <p class="success">{exportMessage}</p>
  {/if}

  <div class="export-row">
    <div>
      <strong>{targetFilename(target)}</strong>
      <span>
        {targetLabel(target)} · {canGenerate ? "ready" : "needs mapping"}
      </span>
    </div>
    <div class="button-row">
      <button class="secondary" type="button" onclick={onBack}>Back to Validation</button>
      <button type="button" disabled={!canGenerate} onclick={generateExport}>Generate Export</button>
    </div>
  </div>
</section>

<style>
  .panel {
    display: grid;
    gap: 18px;
    border: 1px solid rgb(255 255 255 / 0.62);
    border-radius: 24px;
    background: rgb(255 255 255 / 0.66);
    box-shadow: 0 24px 70px rgb(31 42 55 / 0.12);
    backdrop-filter: blur(22px);
    padding: 20px;
  }

  .panel-heading,
  .export-row {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    gap: 18px;
    min-width: 0;
  }

  .panel-heading > div,
  .export-row > div:first-child {
    flex: 1 1 260px;
    min-width: 0;
  }

  h2,
  p {
    margin: 0;
  }

  h2 {
    color: #13251c;
    font-size: 1.05rem;
  }

  p,
  span {
    color: #62716a;
    font-size: 0.88rem;
    line-height: 1.45;
    overflow-wrap: anywhere;
  }

  .panel-heading > span {
    border: 1px solid #d7e4df;
    border-radius: 999px;
    color: #255840;
    font-weight: 700;
    padding: 5px 10px;
  }

  label {
    display: grid;
    gap: 8px;
    color: #33443c;
    font-weight: 700;
  }

  .blockers,
  .ready-state,
  .success {
    border: 1px solid rgb(214 226 222 / 0.78);
    border-radius: 16px;
    background: rgb(255 255 255 / 0.44);
    padding: 14px;
  }

  .blockers {
    border-color: rgb(217 119 6 / 0.42);
    background: rgb(255 251 235 / 0.72);
  }

  .blockers strong,
  .ready-state strong {
    color: #13251c;
  }

  .blockers ul {
    display: grid;
    gap: 6px;
    margin: 8px 0 0;
    padding-left: 20px;
  }

  .blockers li {
    color: #7c3f0f;
    line-height: 1.45;
  }

  .ready-state {
    border-color: #c9ded2;
    background: #f5fbf7;
  }

  .ready-state {
    display: grid;
    gap: 4px;
  }

  .success {
    margin: 0;
    border-color: rgb(37 88 64 / 0.26);
    color: #255840;
    font-weight: 700;
  }

  select {
    width: 100%;
    min-height: 42px;
    border: 1px solid rgb(194 210 204 / 0.95);
    border-radius: 12px;
    background: rgb(255 255 255 / 0.82);
    color: #13251c;
    padding: 0 12px;
  }

  .export-row {
    border-top: 1px solid rgb(218 228 224 / 0.78);
    padding-top: 16px;
  }

  .export-row div {
    display: grid;
    gap: 4px;
  }

  strong {
    color: #13251c;
  }

  .button-row {
    display: flex;
    flex: 0 1 auto;
    flex-wrap: wrap;
    gap: 10px;
    justify-content: flex-end;
  }

  .button-row button {
    min-width: 160px;
  }

  .secondary {
    border: 1px solid rgb(194 210 204 / 0.9);
    background: rgb(255 255 255 / 0.72);
    color: #244f3b;
  }

  @media (max-width: 560px) {
    .panel {
      border-radius: 18px;
    }

    .panel-heading,
    .export-row {
      align-items: stretch;
      flex-direction: column;
    }

    .button-row button,
    .button-row {
      width: 100%;
    }
  }
</style>
