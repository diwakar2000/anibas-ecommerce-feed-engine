<script lang="ts">
  import type { DashboardData } from "../types";

  type Props = {
    dashboard: DashboardData;
    deletingImportId: number | null;
    openingImportId: number | null;
    onUploadClick: () => void;
    onOpenImport: (id: number) => void;
    onDeleteImport: (id: number) => void;
  };

  let { dashboard, deletingImportId, openingImportId, onUploadClick, onOpenImport, onDeleteImport }: Props =
    $props();
  const bannerUrl = `${import.meta.env.BASE_URL}anibas-dashboard-banner.svg`;

  function formatDate(value: string) {
    if (!value) {
      return "Not recorded";
    }

    return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit"
    }).format(new Date(value));
  }
</script>

<section class="dashboard">
  <section class="dashboard-banner">
    <img src={bannerUrl} alt="" />
    <div>
      <p>Anibas Feed Engine</p>
      <h1>Catalog conversion workspace</h1>
      <span>Turn source catalog files into channel-ready publishing feeds.</span>
    </div>
    <button type="button" onclick={onUploadClick}>Upload Catalog</button>
  </section>

  <div class="summary">
    <div>
      <span>Recent imports</span>
      <strong>{dashboard.recent_imports.length}</strong>
    </div>
    <div>
      <span>Recent exports</span>
      <strong>{dashboard.recent_exports.length}</strong>
    </div>
    <button type="button" onclick={onUploadClick}>Upload Catalog</button>
  </div>

  <div class="activity">
    <section class="panel">
      <div class="panel-heading">
        <h2>Imports</h2>
        <span>Latest files</span>
      </div>

      {#if dashboard.recent_imports.length === 0}
        <p class="empty">No imports yet.</p>
      {:else}
        <div class="list">
          {#each dashboard.recent_imports as item}
            <article>
              <div>
                <strong>{item.filename}</strong>
                <span>{item.source_platform} · {item.row_count} rows</span>
              </div>
              <div class="row-actions">
                <time>{formatDate(item.created_at)}</time>
                <button
                  class="open-button"
                  type="button"
                  disabled={openingImportId === item.id}
                  onclick={() => onOpenImport(item.id)}
                >
                  {openingImportId === item.id ? "Opening" : "Open"}
                </button>
                <button
                  class="delete-button"
                  type="button"
                  disabled={deletingImportId === item.id}
                  onclick={() => onDeleteImport(item.id)}
                >
                  {deletingImportId === item.id ? "Deleting" : "Delete"}
                </button>
              </div>
            </article>
          {/each}
        </div>
      {/if}
    </section>

    <section class="panel">
      <div class="panel-heading">
        <h2>Exports</h2>
        <span>Generated files</span>
      </div>

      {#if dashboard.recent_exports.length === 0}
        <p class="empty">No exports yet.</p>
      {:else}
        <div class="list">
          {#each dashboard.recent_exports as item}
            <article>
              <div>
                <strong>{item.filename}</strong>
                <span>{item.target_platform} · {item.row_count} rows</span>
              </div>
              <time>{formatDate(item.created_at)}</time>
            </article>
          {/each}
        </div>
      {/if}
    </section>
  </div>
</section>

<style>
  .dashboard {
    display: grid;
    gap: 18px;
  }

  .dashboard-banner {
    position: relative;
    display: flex;
    align-items: end;
    justify-content: space-between;
    min-height: 260px;
    overflow: hidden;
    border-radius: 24px;
    padding: 28px;
    isolation: isolate;
  }

  .dashboard-banner img {
    position: absolute;
    inset: 0;
    z-index: -2;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .dashboard-banner::after {
    position: absolute;
    inset: 0;
    z-index: -1;
    background: linear-gradient(90deg, rgb(5 27 22 / 0.62), rgb(5 27 22 / 0.16) 68%, rgb(5 27 22 / 0.32));
    content: "";
  }

  .dashboard-banner div {
    display: grid;
    gap: 8px;
    max-width: 640px;
    min-width: 0;
  }

  .dashboard-banner p,
  .dashboard-banner h1,
  .dashboard-banner span {
    margin: 0;
    color: #ffffff;
  }

  .dashboard-banner p {
    font-size: 0.78rem;
    font-weight: 900;
    letter-spacing: 0;
    text-transform: uppercase;
  }

  .dashboard-banner h1 {
    font-size: clamp(2rem, 4vw, 3.6rem);
    line-height: 1;
  }

  .dashboard-banner span {
    max-width: 44ch;
    color: rgb(255 255 255 / 0.84);
    line-height: 1.5;
  }

  .dashboard-banner button {
    flex: 0 0 auto;
    border: 1px solid rgb(255 255 255 / 0.44);
    background: rgb(255 255 255 / 0.94);
    color: #183f31;
  }

  .summary {
    display: grid;
    grid-template-columns: repeat(2, minmax(160px, 1fr)) auto;
    gap: 14px;
    align-items: stretch;
  }

  .summary div,
  .panel {
    border: 1px solid rgb(255 255 255 / 0.62);
    border-radius: 22px;
    background: rgb(255 255 255 / 0.66);
    box-shadow: 0 24px 70px rgb(31 42 55 / 0.1);
    backdrop-filter: blur(22px);
  }

  .summary div {
    display: grid;
    gap: 8px;
    padding: 18px;
  }

  .summary span,
  .panel-heading span,
  article span,
  time {
    color: #62716a;
    font-size: 0.84rem;
  }

  .summary strong {
    color: #13251c;
    font-size: 2rem;
  }

  button {
    min-width: 170px;
  }

  .activity {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 18px;
  }

  .panel {
    min-height: 240px;
    padding: 20px;
  }

  .panel-heading {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    gap: 14px;
    margin-bottom: 14px;
    min-width: 0;
  }

  h2 {
    margin: 0;
    color: #13251c;
    font-size: 1rem;
  }

  .empty {
    color: #62716a;
    margin: 0;
    padding-top: 10px;
  }

  .list {
    display: grid;
    gap: 10px;
  }

  article {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
    border-top: 1px solid rgb(218 228 224 / 0.78);
    padding-top: 12px;
    min-width: 0;
  }

  article div {
    display: grid;
    gap: 4px;
    min-width: 0;
  }

  article .row-actions {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 10px;
  }

  article strong {
    overflow: hidden;
    color: #13251c;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  time {
    white-space: nowrap;
  }

  .open-button,
  .delete-button {
    min-width: 0;
    min-height: 30px;
    font-size: 0.78rem;
    font-weight: 800;
    padding: 0 10px;
  }

  .open-button {
    border: 1px solid rgb(194 210 204 / 0.9);
    background: rgb(255 255 255 / 0.72);
    color: #244f3b;
  }

  .delete-button {
    border: 1px solid #efc2b7;
    background: #fff8f5;
    color: #9b3324;
  }

  .open-button:disabled,
  .delete-button:disabled {
    cursor: not-allowed;
    opacity: 0.65;
  }

  @media (max-width: 760px) {
    .dashboard-banner {
      display: grid;
      align-items: end;
      min-height: 320px;
      padding: 22px;
    }

    .summary,
    .activity {
      grid-template-columns: 1fr;
    }

    button {
      width: 100%;
    }

    article,
    article .row-actions {
      align-items: start;
    }

    article {
      flex-wrap: wrap;
    }

    article .row-actions {
      display: grid;
      justify-items: end;
    }
  }
</style>
