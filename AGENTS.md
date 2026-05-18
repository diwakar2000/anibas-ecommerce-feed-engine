# AGENTS.md

## Project Name

anibas-ecommerce-feed-engine

## Internal Product Name

Anibas Feed Engine

## Product Category

Commerce catalog transformation platform

---

# Product Vision

Anibas Feed Engine is a commerce catalog transformation platform.

The purpose of the application is to help ecommerce and social-commerce sellers move product catalogs between different marketplaces and commerce platforms with minimal manual work.

The application is NOT:
- an ecommerce storefront builder
- a checkout/payment platform
- a Shopify replacement
- an AI wrapper
- a CRM (yet)
- a warehouse management system
- an ERP

The application IS:
- a catalog conversion engine
- a marketplace feed transformation tool
- a product data normalization system
- a multi-channel catalog publishing platform

The main value of the product is:
- reducing repetitive manual listing work
- converting incompatible marketplace formats
- validating product feeds before upload
- preparing sellers for multi-platform selling

---

# Long-Term Direction

The project should evolve in phases.

## Phase 1 — File-Based Catalog Conversion

Features:
- CSV/XLSX upload
- column detection
- field mapping
- validation
- export to target marketplace formats

No live integrations yet.

This phase is the highest priority.

---

## Phase 2 — Saved Mapping Profiles

Features:
- reusable mappings
- saved import/export templates
- category mapping presets
- pricing transformation rules

---

## Phase 3 — Direct Marketplace Integrations

Potential platforms:
- Facebook Commerce
- Instagram Shops
- TikTok Shop
- Shopify
- WooCommerce
- Google Merchant Center

Features:
- OAuth connections
- direct product import
- direct product publishing

Still NOT real-time sync.

---

## Phase 4 — Inventory Synchronization

Features:
- inventory updates
- stock propagation
- webhook handling
- reconciliation jobs
- retry queues
- audit logs

This phase introduces distributed state complexity and should not be started prematurely.

---

## Phase 5 — Operational CRM

Potential future features:
- customer timeline
- order tracking
- messaging aggregation
- shipping events
- call logging
- support workflows

This is intentionally out of scope for the MVP.

---

# Core Product Principles

## 1. Universal Product Schema First

All marketplace formats must convert through an internal normalized schema.

Never create direct platform-to-platform conversion pipelines like:
- TikTok → Facebook
- Shopify → WooCommerce

Instead always use:

Platform Importer
→ Universal Product Schema
→ Platform Exporter

The internal schema is the source of truth.

---

## 2. Marketplace Logic Must Be Modular

Every platform integration should be isolated.

Each marketplace should contain:
- importer
- exporter
- validators
- category mapping
- transformation rules

No marketplace-specific logic should leak into unrelated components.

---

## 3. Avoid Premature Complexity

Do NOT implement:
- real-time sync
- websockets
- distributed workers
- microservices
- event sourcing
- advanced queue orchestration
- AI features
- recommendation systems
- analytics dashboards

until the MVP is stable and validated.

Simple architecture is preferred early.

---

## 4. Validation Is a Core Feature

Validation is not secondary.

The application should help users avoid marketplace upload failures.

Examples:
- duplicate SKUs
- invalid image URLs
- missing required fields
- invalid price formats
- unsupported categories
- oversized titles
- missing variants

Validation should produce:
- errors
- warnings
- actionable feedback

---

## 5. Variants Are Important

Marketplace product variants are handled differently across platforms.

Variant handling must be designed carefully from the beginning.

Examples:
- parent-child products
- flat variation rows
- grouped options
- matrix-style structures

The system architecture should allow flexible variant representation.

---

# Backend Architecture Rules

Preferred backend structure:

/internal
    /domain
    /services
    /repositories
    /handlers
    /validators
    /importers
    /exporters
    /mappers

Guidelines:
- keep domain logic separated from transport logic
- keep handlers thin
- business logic belongs in services
- validation should be reusable
- importers/exporters should implement interfaces

Avoid:
- huge service files
- tightly coupled packages
- circular dependencies
- hidden global state

---

# Frontend Principles

Frontend should prioritize:
- clarity
- usability
- operational workflows
- fast interactions

Avoid:
- flashy animations
- excessive UI frameworks
- overly decorative dashboards

The UI should feel like:
- business software
- operational tooling
- workflow software

not a social app.

---

# Database Principles

Preferred database:
- PostgreSQL

Guidelines:
- explicit schemas
- indexed lookup fields
- audit-friendly timestamps
- avoid premature optimization

Do not build:
- distributed database systems
- multi-region architecture
- complex tenancy systems

early in the project.

---

# Naming Conventions

Preferred terminology:

Use:
- catalog
- feed
- transformation
- mapping
- publishing
- channel
- marketplace
- product schema

Avoid:
- AI-powered
- sync (until true sync exists)
- omnichannel ERP
- all-in-one commerce platform

---

# Performance Expectations

The application should comfortably handle:
- thousands of products
- large CSV imports
- large exports

CSV parsing and transformation should be memory-conscious.

Avoid loading unnecessarily large datasets entirely into memory if streaming approaches are feasible.

---

# Security Principles

Never trust uploaded files.

Uploaded files must:
- be validated
- size-limited
- sanitized where applicable

Avoid:
- executing uploaded content
- unsafe parsing logic
- insecure file handling

---

# Testing Expectations

Critical logic should be tested:
- parsers
- validators
- mappers
- exporters

Transformation correctness is extremely important.

Marketplace export regressions are unacceptable.

---

# MVP Priorities

Highest priority:
1. Upload
2. Parse
3. Map
4. Validate
5. Export

Everything else is secondary.

---

# Out of Scope For MVP

The following should NOT be added unless explicitly requested:

- AI features
- chatbots
- payment systems
- storefront builders
- checkout flows
- order fulfillment
- warehouse management
- shipping integrations
- CRM systems
- customer messaging systems
- social media scraping
- analytics platforms
- affiliate systems
- ad systems
- recommendation engines

Focus on catalog transformation only.

---

# Definition of Success

A seller should be able to:
1. export products from Platform A
2. upload file into Anibas Feed Engine
3. map fields once
4. validate issues
5. export a compatible file for Platform B
6. upload successfully without manual spreadsheet work

That is the core product outcome.