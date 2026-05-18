export type CatalogImport = {
  id: number;
  filename: string;
  source_platform: string;
  row_count: number;
  status: string;
  created_at: string;
};

export type CatalogExport = {
  id: number;
  import_id: number;
  target_platform: string;
  filename: string;
  row_count: number;
  status: string;
  created_at: string;
};

export type DashboardData = {
  recent_imports: CatalogImport[];
  recent_exports: CatalogExport[];
};

export type PreviewRow = {
  row_number: number;
  values: Record<string, string>;
};

export type UploadPreview = {
  catalog_import: CatalogImport;
  columns: string[];
  preview_rows: PreviewRow[];
  row_count: number;
  mapping_suggestions?: MappingSuggestion[];
};

export type MappingSuggestion = {
  field: string;
  source_column: string;
  confidence: number;
  reason: string;
};

export type FieldMappingValue = {
  mode: "column" | "static";
  value: string;
};

export type FieldMappingState = Record<string, FieldMappingValue>;

export type MappingProfile = {
  id: number;
  name: string;
  source_platform: string;
  target_platform: string;
  mapping_json: FieldMappingState | Record<string, string>;
  created_at: string;
  updated_at: string;
};

export type SchemaField = {
  name: string;
  label: string;
  required: boolean;
  description?: string;
};

export type FieldRequirement = {
  field: string;
  level: "required" | "conditional" | "recommended";
  note?: string;
};

export type FieldRequirementGroup = {
  fields: string[];
  min: number;
  level: "required" | "conditional" | "recommended";
  note: string;
};

export type TargetRequirement = {
  id: string;
  label: string;
  required_fields: FieldRequirement[];
  conditional_fields: FieldRequirement[];
  recommended_fields: FieldRequirement[];
  requirement_groups: FieldRequirementGroup[];
  unsupported_columns?: string[];
  notes?: string[];
};

export const fallbackSchemaFields: SchemaField[] = [
  { name: "id", label: "ID", required: false },
  { name: "sku", label: "SKU", required: true },
  { name: "title", label: "Title", required: true },
  { name: "description", label: "Description", required: true },
  { name: "price", label: "Price", required: true },
  { name: "currency", label: "Currency", required: true },
  { name: "quantity", label: "Quantity", required: true },
  { name: "condition", label: "Condition", required: true },
  { name: "brand", label: "Brand", required: false },
  { name: "gtin", label: "GTIN", required: false },
  { name: "mpn", label: "MPN", required: false },
  { name: "category", label: "Category", required: false },
  { name: "image_url", label: "Image URL", required: true },
  { name: "additional_image_urls", label: "Additional Image URLs", required: false },
  { name: "product_url", label: "Product URL", required: true },
  { name: "availability", label: "Availability", required: true },
  { name: "weight", label: "Weight", required: false },
  { name: "variant_group_id", label: "Variant Group ID", required: false },
  { name: "option_1_name", label: "Option 1 Name", required: false },
  { name: "option_1_value", label: "Option 1 Value", required: false },
  { name: "option_2_name", label: "Option 2 Name", required: false },
  { name: "option_2_value", label: "Option 2 Value", required: false },
  { name: "source_platform", label: "Source Platform", required: false },
  { name: "created_at", label: "Created At", required: false },
  { name: "updated_at", label: "Updated At", required: false }
];

export const fallbackTargetRequirements: TargetRequirement[] = [
  {
    id: "facebook_catalog_csv",
    label: "Facebook Catalog CSV",
    required_fields: [
      "title",
      "description",
      "availability",
      "condition",
      "price",
      "currency",
      "product_url",
      "image_url"
    ].map((field) => ({ field, level: "required" })),
    conditional_fields: [
      { field: "quantity", level: "conditional", note: "Required for some Meta commerce surfaces." },
      { field: "category", level: "conditional", note: "Required for some checkout and Marketplace use cases." }
    ],
    recommended_fields: [],
    requirement_groups: [
      {
        fields: ["id", "sku"],
        min: 1,
        level: "required",
        note: "Meta requires a stable product id; SKU can be exported as id."
      },
      {
        fields: ["brand", "gtin", "mpn"],
        min: 1,
        level: "required",
        note: "Meta requires at least one product identifier."
      }
    ]
  }
];
