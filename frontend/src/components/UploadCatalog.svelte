<script lang="ts">
  import type { UploadPreview } from "../types";

  type Props = {
    preview: UploadPreview | null;
    disabled?: boolean;
    onUploadFile: (file: File, sourcePlatform: string) => Promise<void>;
  };

  let { preview, disabled = false, onUploadFile }: Props = $props();
  let selectedFile = $state<File | null>(null);
  let sourcePlatform = $state("generic_csv");
  let uploading = $state(false);
  let error = $state("");

  function selectFile(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    selectedFile = input.files?.[0] ?? null;
    error = "";
  }

  async function uploadCatalog() {
    if (!selectedFile) {
      error = "Choose a CSV file";
      return;
    }
    if (disabled) {
      error = "Catalog engine is still loading";
      return;
    }

    uploading = true;
    error = "";

    try {
      await onUploadFile(selectedFile, sourcePlatform);
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not parse catalog";
    } finally {
      uploading = false;
    }
  }
</script>

<section class="panel">
  <div class="panel-heading">
    <div>
      <h2>Upload Catalog</h2>
      <p>{preview ? preview.catalog_import.filename : "CSV import"}</p>
    </div>
    {#if preview}
      <span>{preview.row_count} rows</span>
    {/if}
  </div>

  <form
    onsubmit={(event) => {
      event.preventDefault();
      uploadCatalog();
    }}
  >
    <label>
      Source platform
      <select bind:value={sourcePlatform}>
        <option value="generic_csv">Generic CSV</option>
        <option value="tiktok_shop">TikTok Shop</option>
        <option value="facebook_commerce">Facebook Commerce</option>
        <option value="shopify">Shopify</option>
        <option value="woocommerce">WooCommerce</option>
        <option value="google_merchant_center">Google Merchant Center</option>
      </select>
    </label>

    <label>
      Catalog file
      <input type="file" accept=".csv,text/csv" disabled={disabled || uploading} onchange={selectFile} />
    </label>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    <button type="submit" disabled={disabled || uploading}>
      {uploading ? "Parsing..." : disabled ? "Loading Engine" : "Parse Catalog"}
    </button>
  </form>
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
    flex: 1 1 220px;
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
  .panel-heading span {
    color: #62716a;
    line-height: 1.45;
    overflow-wrap: anywhere;
  }

  .panel-heading span {
    font-size: 0.88rem;
    font-weight: 700;
  }

  form {
    display: grid;
    gap: 14px;
  }

  label {
    display: grid;
    gap: 7px;
    color: #33443c;
    font-weight: 700;
    min-width: 0;
  }

  input,
  select {
    width: 100%;
    min-width: 0;
    min-height: 42px;
    border: 1px solid rgb(194 210 204 / 0.95);
    border-radius: 12px;
    background: rgb(255 255 255 / 0.82);
    color: #13251c;
    padding: 0 12px;
  }

  input[type="file"] {
    padding: 9px 12px;
  }

  .error {
    border: 1px solid #f1b7a8;
    border-radius: 8px;
    background: #fff6f2;
    color: #9b3324;
    padding: 10px 12px;
  }

  @media (max-width: 640px) {
    .panel {
      border-radius: 18px;
    }
  }
</style>
