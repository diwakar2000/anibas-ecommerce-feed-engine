<script lang="ts">
  import type { UploadPreview } from "../types";

  type Props = {
    preview: UploadPreview | null;
  };

  let { preview }: Props = $props();
  let expandedCell = $state("");

  function toggleCell(event: MouseEvent, key: string) {
    event.stopPropagation();
    expandedCell = expandedCell === key ? "" : key;
  }

  function collapseCell(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest(".cell-value")) {
      expandedCell = "";
    }
  }
</script>

<svelte:window onclick={collapseCell} />

<section class="panel">
  <div class="panel-heading">
    <div>
      <h2>Preview</h2>
      <p>
        {preview
          ? `${preview.preview_rows.length} preview rows from ${preview.row_count} total rows`
          : "No file uploaded"}
      </p>
    </div>
    {#if preview}
      <span>{preview.columns.length} columns</span>
    {/if}
  </div>

  {#if preview}
    <div class="columns">
      {#each preview.columns as column}
        <span>{column}</span>
      {/each}
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Row</th>
            {#each preview.columns as column}
              <th>{column}</th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each preview.preview_rows as row}
            <tr>
              <td>{row.row_number}</td>
              {#each preview.columns as column}
                {@const key = `${row.row_number}:${column}`}
                {@const value = row.values[column] ?? ""}
                <td>
                  <button
                    class="cell-value"
                    class:expanded={expandedCell === key}
                    type="button"
                    aria-expanded={expandedCell === key}
                    onclick={(event) => toggleCell(event, key)}
                  >
                    {value || "—"}
                  </button>
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else}
    <p class="empty">No preview available.</p>
  {/if}
</section>

<style>
  .panel {
    display: grid;
    gap: 16px;
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
  .empty {
    color: #62716a;
    line-height: 1.45;
    overflow-wrap: anywhere;
  }

  .panel-heading span {
    font-size: 0.88rem;
    font-weight: 700;
  }

  .columns {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .columns span {
    max-width: 100%;
    border: 1px solid rgb(215 228 223 / 0.92);
    border-radius: 999px;
    background: rgb(255 255 255 / 0.52);
    color: #255840;
    font-size: 0.8rem;
    font-weight: 700;
    overflow-wrap: anywhere;
    padding: 5px 9px;
  }

  .table-wrap {
    max-width: 100%;
    overflow-x: auto;
    border: 1px solid rgb(214 226 222 / 0.78);
    border-radius: 16px;
  }

  table {
    width: max-content;
    min-width: 100%;
    border-collapse: collapse;
  }

  th,
  td {
    border-bottom: 1px solid #edf1ef;
    padding: 10px 12px;
    text-align: left;
    vertical-align: top;
  }

  th:first-child,
  td:first-child {
    position: sticky;
    left: 0;
    z-index: 1;
    width: 72px;
    min-width: 72px;
    max-width: 72px;
    background: rgb(249 251 250 / 0.96);
  }

  th {
    background: rgb(244 247 246 / 0.86);
    color: #33443c;
    font-size: 0.82rem;
    max-width: 190px;
    min-width: 190px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  td {
    color: #1f3028;
    font-size: 0.9rem;
    max-width: 190px;
    min-width: 190px;
    overflow-wrap: anywhere;
  }

  .cell-value {
    display: -webkit-box;
    width: 100%;
    min-height: 34px;
    max-height: 44px;
    border: 1px solid transparent;
    border-radius: 10px;
    background: transparent;
    box-shadow: none;
    color: #1f3028;
    cursor: zoom-in;
    font-size: 0.84rem;
    font-weight: 600;
    line-height: 1.35;
    overflow: hidden;
    padding: 0;
    text-align: left;
    text-overflow: ellipsis;
    transform: none;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
  }

  .cell-value:hover {
    border-color: rgb(37 88 64 / 0.18);
    background: rgb(255 255 255 / 0.58);
    box-shadow: none;
    transform: none;
  }

  .cell-value.expanded {
    display: block;
    max-height: 220px;
    min-width: 320px;
    overflow: auto;
    border-color: rgb(37 88 64 / 0.35);
    background: rgb(255 255 255 / 0.96);
    box-shadow: 0 18px 34px rgb(31 42 55 / 0.16);
    cursor: zoom-out;
    padding: 10px;
    position: relative;
    z-index: 3;
    line-clamp: unset;
    -webkit-line-clamp: unset;
  }

  tbody tr:last-child td {
    border-bottom: 0;
  }

  @media (max-width: 640px) {
    .panel {
      border-radius: 18px;
    }
  }
</style>
