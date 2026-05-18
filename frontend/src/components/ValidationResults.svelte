<script lang="ts">
  import {
    isValidCurrency,
    mappingHasValue,
    mappingSourceLabel,
    resolveMappedValue
  } from "../mapping";
  import type { FieldMappingState, SchemaField, TargetRequirement, UploadPreview } from "../types";
  import type { AnibasWasm } from "../wasm";

  type Finding = {
    level: "error" | "warning";
    title: string;
    detail: string;
    rowNumber?: number;
  };

  type Props = {
    preview: UploadPreview | null;
    fields: SchemaField[];
    mapping: FieldMappingState;
    targetRequirements: TargetRequirement[];
    selectedTarget: string;
    wasmEngine: AnibasWasm | null;
    onStatusChange?: (errorCount: number) => void;
    onBack: () => void;
    onContinue: () => void;
  };

  let {
    preview,
    fields,
    mapping,
    targetRequirements,
    selectedTarget,
    wasmEngine,
    onStatusChange,
    onBack,
    onContinue
  }: Props = $props();
  let selectedRequirement = $derived(
    targetRequirements.find((requirement) => requirement.id === selectedTarget) ?? targetRequirements[0]
  );
  let requiredCount = $derived(
    requiredFields(selectedRequirement).length +
      requirementGroups(selectedRequirement).filter((group) => group.level === "required").length
  );
  let fieldCount = $derived(fields.length);
  let findings = $derived(validateWithEngine(preview, selectedRequirement, mapping, wasmEngine));
  let errors = $derived(findings.filter((finding) => finding.level === "error"));
  let warnings = $derived(findings.filter((finding) => finding.level === "warning"));

  $effect(() => {
    onStatusChange?.(errors.length);
  });

  function requiredFields(requirement: TargetRequirement | undefined) {
    return requirement?.required_fields ?? [];
  }

  function requirementGroups(requirement: TargetRequirement | undefined) {
    return requirement?.requirement_groups ?? [];
  }

  function validatePreview(
    currentPreview: UploadPreview | null,
    requirement: TargetRequirement | undefined,
    currentMapping: FieldMappingState
  ): Finding[] {
    const result: Finding[] = [];

    if (!currentPreview) {
      return [{
        level: "error",
        title: "Upload a catalog first",
        detail: "Validation needs parsed rows and detected columns before it can check the feed."
      }];
    }

    if (!requirement) {
      return [{
        level: "error",
        title: "Choose a target format",
        detail: "A target format is required so the app knows which fields matter."
      }];
    }

    for (const item of requiredFields(requirement)) {
      if (!mappingHasValue(currentMapping, item.field) && !canDeriveField(currentPreview, currentMapping, item.field)) {
        result.push({
          level: "error",
          title: `Map ${labelForField(item.field)}`,
          detail: item.note ?? `${labelForField(item.field)} is required for ${requirement.label}.`
        });
      }
    }

    for (const group of requirementGroups(requirement).filter((item) => item.level === "required")) {
      const mappedFields = group.fields.filter((field) => mappingHasValue(currentMapping, field));
      if (mappedFields.length < group.min) {
        result.push({
          level: "error",
          title: `Map at least ${group.min} of ${group.fields.map(labelForField).join(", ")}`,
          detail: group.note
        });
      }
    }

    for (const row of currentPreview.preview_rows) {
      for (const item of requiredFields(requirement)) {
        if (!mappingHasValue(currentMapping, item.field) && !canDeriveField(currentPreview, currentMapping, item.field)) {
          continue;
        }

        if (!resolveMappedValue(currentMapping, item.field, row.values)) {
          result.push({
            level: "error",
            title: `${labelForField(item.field)} is empty`,
            detail: `${mappingSourceLabel(currentMapping, item.field) || "Derived value"} has no value in preview row ${row.row_number}.`,
            rowNumber: row.row_number
          });
        }
      }

      for (const group of requirementGroups(requirement).filter((item) => item.level === "required")) {
        const mappedFields = group.fields.filter((field) => mappingHasValue(currentMapping, field));
        if (mappedFields.length < group.min) {
          continue;
        }

        const valuesPresent = mappedFields.filter((field) => resolveMappedValue(currentMapping, field, row.values)).length;
        if (valuesPresent < group.min) {
          result.push({
            level: "error",
            title: `Identifier value is missing`,
            detail: `Preview row ${row.row_number} needs at least ${group.min} of ${group.fields.map(labelForField).join(", ")}.`,
            rowNumber: row.row_number
          });
        }
      }
    }

    result.push(...validateDuplicates(currentPreview, currentMapping));
    result.push(...validateFieldFormats(currentPreview, currentMapping));

    return result;
  }

  function validateWithEngine(
    currentPreview: UploadPreview | null,
    requirement: TargetRequirement | undefined,
    currentMapping: FieldMappingState,
    engine: AnibasWasm | null
  ): Finding[] {
    if (!currentPreview || !engine) {
      return validatePreview(currentPreview, requirement, currentMapping);
    }

    try {
      return engine.validateCatalog({
        preview_rows: currentPreview.preview_rows,
        mapping: currentMapping,
        target_platform: selectedTarget,
        source_platform: currentPreview.catalog_import.source_platform
      }).findings;
    } catch {
      return validatePreview(currentPreview, requirement, currentMapping);
    }
  }

  function canDeriveField(
    currentPreview: UploadPreview,
    currentMapping: FieldMappingState,
    fieldName: string
  ) {
    if (fieldName !== "currency" || !mappingHasValue(currentMapping, "price")) {
      return false;
    }

    return currentPreview.preview_rows.some((row) => resolveMappedValue(currentMapping, "currency", row.values));
  }

  function validateDuplicates(currentPreview: UploadPreview, currentMapping: FieldMappingState) {
    if (!mappingHasValue(currentMapping, "sku")) {
      return [];
    }

    const seen = new Map<string, number>();
    const duplicates: Finding[] = [];
    for (const row of currentPreview.preview_rows) {
      const sku = resolveMappedValue(currentMapping, "sku", row.values).toLowerCase();
      if (!sku) {
        continue;
      }

      const firstRow = seen.get(sku);
      if (firstRow) {
        duplicates.push({
          level: "error",
          title: "Duplicate SKU in preview",
          detail: `SKU "${resolveMappedValue(currentMapping, "sku", row.values)}" appears in rows ${firstRow} and ${row.row_number}.`,
          rowNumber: row.row_number
        });
        continue;
      }

      seen.set(sku, row.row_number);
    }

    return duplicates;
  }

  function validateFieldFormats(currentPreview: UploadPreview, currentMapping: FieldMappingState) {
    const result: Finding[] = [];
    for (const row of currentPreview.preview_rows) {
      const price = mappedValue(row.values, currentMapping, "price");
      if (price && !isValidPrice(price)) {
        result.push(formatFinding("error", "Invalid price", row.row_number, price, "Use a positive number such as 19.99."));
      }

      const quantity = mappedValue(row.values, currentMapping, "quantity");
      if (quantity && !/^\d+$/.test(quantity)) {
        result.push(formatFinding("warning", "Invalid quantity", row.row_number, quantity, "Use a whole number greater than or equal to 0."));
      }

      const currency = mappedValue(row.values, currentMapping, "currency");
      if (currency && !isValidCurrency(currency)) {
        result.push(formatFinding("warning", "Invalid currency", row.row_number, currency, "Use a three-letter ISO code like USD or a currency symbol like $."));
      }

      for (const field of ["image_url", "product_url"]) {
        const value = mappedValue(row.values, currentMapping, field);
        if (value && !isHttpURL(value)) {
          result.push(formatFinding("error", `Invalid ${labelForField(field)}`, row.row_number, value, "Use a full http:// or https:// URL."));
        }
      }

      const availability = mappedValue(row.values, currentMapping, "availability");
      if (availability && !["in stock", "out of stock", "preorder", "available for order", "discontinued"].includes(availability.toLowerCase())) {
        result.push(formatFinding("warning", "Unusual availability", row.row_number, availability, "Use a marketplace-supported availability value."));
      }

      const condition = mappedValue(row.values, currentMapping, "condition");
      if (condition && !["new", "used", "refurbished"].includes(condition.toLowerCase())) {
        result.push(formatFinding("warning", "Unusual condition", row.row_number, condition, "Use new, used, or refurbished."));
      }
    }

    return result;
  }

  function mappedValue(values: Record<string, string>, currentMapping: FieldMappingState, field: string) {
    return resolveMappedValue(currentMapping, field, values);
  }

  function isValidPrice(value: string) {
    const normalized = value.replace(/[$,\s]/g, "");
    const amount = Number(normalized);
    return Number.isFinite(amount) && amount > 0;
  }

  function isHttpURL(value: string) {
    try {
      const url = new URL(value);
      return url.protocol === "http:" || url.protocol === "https:";
    } catch {
      return false;
    }
  }

  function formatFinding(level: Finding["level"], title: string, rowNumber: number, value: string, detail: string): Finding {
    return {
      level,
      title,
      detail: `Preview row ${rowNumber}: "${value}". ${detail}`,
      rowNumber
    };
  }

  function labelForField(fieldName: string) {
    return fields.find((field) => field.name === fieldName)?.label ?? fieldName;
  }

  function downloadReport() {
    const rows = [
      ["level", "row_number", "title", "detail"],
      ...findings.map((finding) => [
        finding.level,
        finding.rowNumber ? String(finding.rowNumber) : "",
        finding.title,
        finding.detail
      ])
    ];
    const csv = rows
      .map((row) => row.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(","))
      .join("\n");
    const url = URL.createObjectURL(new Blob([csv], { type: "text/csv" }));
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "validation-report.csv";
    anchor.click();
    URL.revokeObjectURL(url);
  }
</script>

<section class="panel">
  <div class="panel-heading">
    <div>
      <h2>Validation Results</h2>
      <p>
        {preview ? `${preview.row_count} rows ready for validation` : "No catalog staged"} ·
        {selectedRequirement?.label ?? "Target"}
      </p>
    </div>
    <span>{errors.length} errors · {warnings.length} warnings · {requiredCount} required items</span>
  </div>

  {#if findings.length === 0}
    <div class="ready-state">
      <strong>No blocking issues found in the preview rows.</strong>
      <span>{fieldCount} schema fields checked against the current mapping.</span>
    </div>
  {:else}
    <div class="results">
      {#each findings as finding}
        <article class:error={finding.level === "error"} class:warning={finding.level === "warning"}>
          <div>
            <strong>{finding.title}</strong>
            <span>{finding.detail}</span>
          </div>
          <em>{finding.level}</em>
        </article>
      {/each}
    </div>
  {/if}

  <div class="actions">
    <button type="button" disabled={findings.length === 0} onclick={downloadReport}>Download Error Report</button>
    <button class="secondary" type="button" onclick={onBack}>Back to Mapping</button>
    <button type="button" disabled={!preview || errors.length > 0} onclick={onContinue}>Continue to Export</button>
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

  .panel-heading {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    gap: 18px;
    min-width: 0;
  }

  .panel-heading > div {
    flex: 1 1 280px;
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
  .panel-heading span,
  article span {
    color: #62716a;
    line-height: 1.45;
    overflow-wrap: anywhere;
  }

  .panel-heading span {
    flex: 0 1 auto;
    font-size: 0.88rem;
    font-weight: 700;
  }

  .results {
    display: grid;
    gap: 10px;
  }

  article,
  .ready-state {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
    border: 1px solid rgb(214 226 222 / 0.78);
    border-radius: 16px;
    background: rgb(255 255 255 / 0.44);
    padding: 14px;
  }

  .ready-state {
    border-color: #c9ded2;
    background: #f5fbf7;
  }

  article > div {
    display: grid;
    gap: 4px;
    min-width: 0;
  }

  article.error {
    border-color: rgb(220 38 38 / 0.42);
    background: rgb(254 242 242 / 0.72);
  }

  article.warning {
    border-color: rgb(217 119 6 / 0.42);
    background: rgb(255 251 235 / 0.72);
  }

  strong {
    color: #13251c;
  }

  em {
    border-radius: 999px;
    color: #33443c;
    font-size: 0.72rem;
    font-style: normal;
    font-weight: 900;
    text-transform: uppercase;
  }

  article.error em {
    color: #9b1c1c;
  }

  article.warning em {
    color: #92400e;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 12px;
    border-top: 1px solid rgb(218 228 224 / 0.78);
    padding-top: 16px;
  }

  .actions button:first-child {
    margin-right: auto;
  }

  .secondary {
    border: 1px solid rgb(194 210 204 / 0.9);
    background: rgb(255 255 255 / 0.72);
    color: #244f3b;
  }

  @media (max-width: 640px) {
    .panel {
      border-radius: 18px;
    }

    article {
      align-items: flex-start;
      flex-direction: column;
    }

    .actions button {
      width: 100%;
    }

    .actions button:first-child {
      margin-right: 0;
    }
  }
</style>
