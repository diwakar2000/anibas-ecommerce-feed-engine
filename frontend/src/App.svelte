<script lang="ts">
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

  type Product = {
    id: number;
    sku: string;
    title: string;
    description: string;
    price_cents: number;
    currency: string;
    inventory: number;
    created_at: string;
  };

  type ProductForm = {
    sku: string;
    title: string;
    description: string;
    price_cents: number;
    currency: string;
    inventory: number;
  };

  let products = $state<Product[]>([]);
  let loading = $state(true);
  let saving = $state(false);
  let error = $state("");
  let form = $state<ProductForm>({
    sku: "ANI-TEE-001",
    title: "Organic Cotton T-shirt",
    description: "Soft product data ready for the ecommerce feed.",
    price_cents: 2499,
    currency: "USD",
    inventory: 32
  });

  const formatter = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD"
  });

  async function loadProducts() {
    loading = true;
    error = "";

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products`);
      if (!response.ok) {
        throw new Error("Could not load products");
      }

      products = await response.json();
    } catch (err) {
      error = err instanceof Error ? err.message : "Something went wrong";
    } finally {
      loading = false;
    }
  }

  async function createProduct() {
    saving = true;
    error = "";

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(form)
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => ({}));
        throw new Error(payload.error ?? "Could not create product");
      }

      const product = await response.json();
      products = [product, ...products];
      form.sku = `ANI-${Date.now().toString().slice(-6)}`;
      form.title = "";
      form.description = "";
      form.price_cents = 1999;
      form.currency = "USD";
      form.inventory = 10;
    } catch (err) {
      error = err instanceof Error ? err.message : "Something went wrong";
    } finally {
      saving = false;
    }
  }

  function formatPrice(product: Product) {
    if (product.currency === "USD") {
      return formatter.format(product.price_cents / 100);
    }

    return `${product.currency} ${(product.price_cents / 100).toFixed(2)}`;
  }

  $effect(() => {
    loadProducts();
  });
</script>

<main class="shell">
  <section class="overview">
    <div>
      <p class="eyebrow">Anibas Feed Engine</p>
      <h1>Product catalog feed control</h1>
      <p class="lede">
        Review, create, and stage product records before they move into downstream marketplace feeds.
      </p>
    </div>

    <div class="status" aria-label="API status">
      <span class:error={Boolean(error)}></span>
      {error ? "API needs attention" : "API connected"}
    </div>
  </section>

  <section class="workspace" aria-label="Product feed workspace">
    <form class="panel form-panel" onsubmit={(event) => {
      event.preventDefault();
      createProduct();
    }}>
      <div class="panel-heading">
        <p>New product</p>
        <span>Live catalog</span>
      </div>

      <label>
        SKU
        <input bind:value={form.sku} required />
      </label>

      <label>
        Title
        <input bind:value={form.title} required />
      </label>

      <label>
        Description
        <textarea bind:value={form.description} rows="4"></textarea>
      </label>

      <div class="field-grid">
        <label>
          Price cents
          <input type="number" min="1" bind:value={form.price_cents} required />
        </label>

        <label>
          Currency
          <input maxlength="3" bind:value={form.currency} required />
        </label>
      </div>

      <label>
        Inventory
        <input type="number" min="0" bind:value={form.inventory} />
      </label>

      <button type="submit" disabled={saving}>
        {saving ? "Saving..." : "Create product"}
      </button>
    </form>

    <section class="panel products-panel" aria-label="Products">
      <div class="panel-heading">
        <p>Catalog</p>
        <button type="button" class="ghost" onclick={loadProducts}>Refresh</button>
      </div>

      {#if loading}
        <p class="muted">Loading products...</p>
      {:else if error}
        <p class="alert">{error}</p>
      {:else if products.length === 0}
        <p class="muted">No products yet.</p>
      {:else}
        <div class="product-list">
          {#each products as product}
            <article>
              <div>
                <p class="sku">{product.sku}</p>
                <h2>{product.title}</h2>
                <p>{product.description}</p>
              </div>

              <dl>
                <div>
                  <dt>Price</dt>
                  <dd>{formatPrice(product)}</dd>
                </div>
                <div>
                  <dt>Inventory</dt>
                  <dd>{product.inventory}</dd>
                </div>
              </dl>
            </article>
          {/each}
        </div>
      {/if}
    </section>
  </section>
</main>

<style>
  .shell {
    width: min(1120px, calc(100% - 32px));
    margin: 0 auto;
    padding: 48px 0;
  }

  .overview {
    display: flex;
    align-items: end;
    justify-content: space-between;
    gap: 24px;
    margin-bottom: 28px;
  }

  .eyebrow,
  .panel-heading span,
  .sku,
  dt {
    color: #647269;
    font-size: 0.78rem;
    font-weight: 700;
    letter-spacing: 0;
    text-transform: uppercase;
  }

  h1,
  h2,
  p {
    margin: 0;
  }

  h1 {
    max-width: 680px;
    color: #101814;
    font-size: clamp(2.15rem, 5vw, 4.7rem);
    line-height: 0.98;
    margin-top: 10px;
  }

  .lede {
    max-width: 620px;
    color: #526056;
    font-size: 1.08rem;
    line-height: 1.65;
    margin-top: 18px;
  }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    min-height: 40px;
    padding: 0 14px;
    border: 1px solid #d8ded4;
    border-radius: 8px;
    background: #fffaf1;
    color: #233229;
    font-weight: 700;
    white-space: nowrap;
  }

  .status span {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: #2a9d68;
  }

  .status span.error {
    background: #c24135;
  }

  .workspace {
    display: grid;
    grid-template-columns: minmax(300px, 380px) minmax(0, 1fr);
    gap: 20px;
    align-items: start;
  }

  .panel {
    border: 1px solid #dde2d9;
    border-radius: 8px;
    background: #fffdf8;
    box-shadow: 0 16px 40px rgb(39 47 42 / 0.08);
  }

  .form-panel {
    display: grid;
    gap: 16px;
    padding: 22px;
  }

  .panel-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 4px;
  }

  .panel-heading p {
    color: #111b16;
    font-size: 1.1rem;
    font-weight: 800;
  }

  label {
    display: grid;
    gap: 7px;
    color: #39483f;
    font-size: 0.92rem;
    font-weight: 700;
  }

  input,
  textarea {
    width: 100%;
    border: 1px solid #cfd7cc;
    border-radius: 8px;
    background: #fff;
    color: #162019;
    outline: none;
    padding: 11px 12px;
  }

  textarea {
    resize: vertical;
  }

  input:focus,
  textarea:focus {
    border-color: #3d7f5b;
    box-shadow: 0 0 0 3px rgb(61 127 91 / 0.16);
  }

  .field-grid {
    display: grid;
    grid-template-columns: 1fr 110px;
    gap: 12px;
  }

  button {
    min-height: 44px;
    border: 0;
    border-radius: 8px;
    background: #173f2d;
    color: #fff;
    font-weight: 800;
  }

  button:disabled {
    cursor: wait;
    opacity: 0.7;
  }

  .ghost {
    min-height: 34px;
    border: 1px solid #d4dbd1;
    background: #fff;
    color: #173f2d;
    padding: 0 12px;
  }

  .products-panel {
    min-height: 420px;
    padding: 22px;
  }

  .muted,
  .alert {
    color: #647269;
    line-height: 1.6;
    padding-top: 16px;
  }

  .alert {
    color: #a23b31;
    font-weight: 700;
  }

  .product-list {
    display: grid;
    gap: 12px;
    margin-top: 16px;
  }

  article {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 18px;
    padding: 18px;
    border: 1px solid #e2e6df;
    border-radius: 8px;
    background: #fff;
  }

  article h2 {
    color: #101814;
    font-size: 1.08rem;
    line-height: 1.25;
    margin-top: 5px;
  }

  article p:last-child {
    color: #5f6c64;
    line-height: 1.55;
    margin-top: 8px;
  }

  dl {
    display: grid;
    grid-template-columns: repeat(2, minmax(82px, auto));
    gap: 12px;
    margin: 0;
  }

  dd {
    margin: 4px 0 0;
    color: #101814;
    font-weight: 800;
    white-space: nowrap;
  }

  @media (max-width: 820px) {
    .shell {
      width: min(100% - 24px, 640px);
      padding: 32px 0;
    }

    .overview {
      display: grid;
      align-items: start;
    }

    .workspace,
    article {
      grid-template-columns: 1fr;
    }

    .status {
      width: fit-content;
    }
  }

  @media (max-width: 480px) {
    .field-grid,
    dl {
      grid-template-columns: 1fr;
    }

    h1 {
      font-size: 2.25rem;
    }
  }
</style>
