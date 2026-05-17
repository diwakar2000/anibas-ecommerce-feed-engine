# Anibas E-commerce Feed Engine

A small full-stack starter with:

- Svelte + Vite frontend
- Go backend using Gin
- Postgres persistence through Bun ORM
- Docker Compose orchestration on lightweight Alpine-based images

## Project Layout

```text
.
├── backend/        # Go API, repositories, service wiring, database setup
├── frontend/       # Svelte/Vite UI
├── docker-compose.yml
└── .env.example
```

## Quick Start

```bash
cp .env.example .env
docker compose up --build
```

Open:

- Frontend: http://localhost:5173
- Backend health: http://localhost:8080/healthz
- Products API: http://localhost:8080/api/v1/products

## Local Development

Backend:

```bash
cd backend
go mod tidy
go run ./cmd/api
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

When running outside Docker, point the backend at your local Postgres with:

```bash
POSTGRES_HOST=localhost go run ./cmd/api
```

## API

`GET /healthz`

Returns service and database status.

`GET /api/v1/products`

Lists products ordered by creation time.

`POST /api/v1/products`

Creates a product.

```json
{
  "sku": "ANI-001",
  "title": "Organic Cotton T-shirt",
  "description": "Soft everyday tee",
  "price_cents": 2499,
  "currency": "USD",
  "inventory": 42
}
```

