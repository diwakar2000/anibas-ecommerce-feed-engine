# Anibas E-commerce Feed Engine

Anibas Feed Engine is a catalog conversion tool for small social-commerce sellers. It prepares product listing files for marketplace and commerce-channel feeds such as Facebook Commerce, Instagram Shops, TikTok Shop, Shopify, WooCommerce, and Google Merchant Center.

This project is not an ecommerce storefront. It does not handle carts, checkout, themes, payments, order tracking, CRM, or real-time marketplace sync.

## Current Architecture

The app is now a fully static frontend:

- Svelte + Vite UI
- Go transformation engine compiled to WebAssembly
- Browser `localStorage` for saved imports, exports, and mapping profiles
- No Postgres
- No Gin API server
- No Docker runtime required
- GitHub Pages deployment through GitHub Actions

Go still owns the catalog transformation logic, but it runs in the browser as `anibas.wasm`.

## Project Layout

```text
.
├── backend/                 # Pure Go domain/importer/mapper/validator/exporter code plus WASM entrypoint
│   └── cmd/wasm             # Browser-facing Go WASM bridge
├── frontend/                # Svelte/Vite static app
└── .github/workflows/       # GitHub Pages CI/CD
```

## Local Development

Build the WASM engine, then run the frontend:

```bash
cd frontend
npm install
npm run wasm:build
npm run dev
```

Open:

- Frontend: http://localhost:5173

## Build

```bash
cd frontend
npm run build
```

The `prebuild` script compiles:

- `frontend/public/anibas.wasm`
- `frontend/public/wasm_exec.js`

Vite then copies them into the static `dist` output.

## Tests And Checks

```bash
cd backend
go test ./...

cd ../frontend
npm run check
npm run build
```

## GitHub Pages

The workflow at `.github/workflows/deploy-pages.yml`:

1. Installs Go and Node.
2. Runs Go tests.
3. Runs Svelte checks.
4. Builds the Go WASM engine.
5. Builds the Vite app with a GitHub Pages base path.
6. Deploys `frontend/dist` to GitHub Pages.

Enable GitHub Pages in repository settings and choose GitHub Actions as the source.

## Current MVP

- CSV catalog upload in the browser
- detected columns
- first-20-row preview
- auto field mapping suggestions
- case-insensitive mapping
- static mapping values
- price/currency extraction from combined price strings
- target-specific required, conditional, and recommended fields
- validation findings
- CSV exports for Facebook Catalog, Instagram Shops, Google Merchant Center, TikTok Catalog, Shopify, and WooCommerce
- saved mapping profiles in browser storage
- saved imports and export records in browser storage
