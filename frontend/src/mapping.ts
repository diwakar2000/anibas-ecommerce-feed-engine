import type { FieldMappingState, FieldMappingValue } from "./types";

const currencySymbols = new Set(["$", "€", "£", "¥", "₹", "₦", "₱", "₩", "₺", "₫", "₪", "₽"]);

export function normalizeColumn(value: string) {
  return value.toLowerCase().trim().replace(/[\s_-]+/g, " ");
}

export function canonicalColumn(value: string, columns: string[]) {
  const normalizedValue = normalizeColumn(value);
  if (!normalizedValue) {
    return "";
  }

  return columns.find((column) => normalizeColumn(column) === normalizedValue) ?? "";
}

export function mappingFromInput(value: string, columns: string[]): FieldMappingValue | null {
  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }

  const column = canonicalColumn(trimmed, columns);
  if (column) {
    return { mode: "column", value: column };
  }

  return { mode: "static", value: trimmed };
}

export function normalizeMapping(rawMapping: unknown, columns: string[]): FieldMappingState {
  if (!rawMapping || typeof rawMapping !== "object") {
    return {};
  }

  const result: FieldMappingState = {};
  for (const [field, rawValue] of Object.entries(rawMapping as Record<string, unknown>)) {
    const value = normalizeMappingValue(rawValue, columns);
    if (value) {
      result[field] = value;
    }
  }

  return result;
}

export function normalizeMappingValue(rawValue: unknown, columns: string[]): FieldMappingValue | null {
  if (typeof rawValue === "string") {
    return mappingFromInput(rawValue, columns);
  }

  if (!rawValue || typeof rawValue !== "object") {
    return null;
  }

  const candidate = rawValue as Partial<FieldMappingValue>;
  const value = String(candidate.value ?? "").trim();
  if (!value) {
    return null;
  }

  return {
    mode: candidate.mode === "static" ? "static" : "column",
    value
  };
}

export function cleanMapping(mapping: FieldMappingState): FieldMappingState {
  return Object.fromEntries(
    Object.entries(mapping).filter(([, item]) => item.value.trim().length > 0)
  );
}

export function setMappingValue(
  mapping: FieldMappingState,
  fieldName: string,
  value: string,
  columns: string[]
): FieldMappingState {
  const nextMapping = { ...mapping };
  const nextValue = mappingFromInput(value, columns);
  if (!nextValue) {
    delete nextMapping[fieldName];
    return nextMapping;
  }

  nextMapping[fieldName] = nextValue;
  return nextMapping;
}

export function mappingInputValue(mapping: FieldMappingState, fieldName: string) {
  return mapping[fieldName]?.value ?? "";
}

export function mappingHasValue(mapping: FieldMappingState, fieldName: string) {
  return Boolean(mapping[fieldName]?.value.trim());
}

export function mappingModeLabel(mapping: FieldMappingState, fieldName: string) {
  const item = mapping[fieldName];
  if (!item?.value.trim()) {
    return "";
  }

  return item.mode === "static" ? "Static value" : "CSV column";
}

export function mappingSourceLabel(mapping: FieldMappingState, fieldName: string) {
  const item = mapping[fieldName];
  if (!item?.value.trim()) {
    return "";
  }

  return item.mode === "static" ? `static value "${item.value}"` : `source column "${item.value}"`;
}

export function resolveMappedValue(
  mapping: FieldMappingState,
  fieldName: string,
  values: Record<string, string>
) {
  const item = mapping[fieldName];
  if (fieldName === "currency" && !item) {
    return extractCurrencyToken(rawMappedValue(mapping, "price", values));
  }

  const rawValue = rawMappedValue(mapping, fieldName, values);
  if (fieldName === "price") {
    return extractPrice(rawValue) || rawValue;
  }
  if (fieldName === "currency") {
    return extractCurrencyToken(rawValue) || rawValue;
  }

  return rawValue;
}

export function rawMappedValue(
  mapping: FieldMappingState,
  fieldName: string,
  values: Record<string, string>
) {
  const item = mapping[fieldName];
  if (!item) {
    return "";
  }
  if (item.mode === "static") {
    return item.value.trim();
  }

  return (values[item.value] ?? "").trim();
}

export function extractPrice(value: string) {
  const match = value.match(/[-+]?\d[\d\s,]*(?:\.\d+)?|[-+]?\d+(?:,\d{2})/);
  if (!match) {
    return "";
  }

  return normalizePriceNumber(match[0]);
}

export function extractCurrencyToken(value: string) {
  const code = value.match(/\b[A-Z]{3}\b/i)?.[0]?.toUpperCase() ?? "";
  if (code) {
    return code;
  }

  return [...currencySymbols].find((symbol) => value.includes(symbol)) ?? "";
}

export function isValidCurrency(value: string) {
  return /^[A-Za-z]{3}$/.test(value) || currencySymbols.has(value);
}

function normalizePriceNumber(value: string) {
  let normalized = value.replace(/\s/g, "");
  if (/^\d+,\d{2}$/.test(normalized)) {
    normalized = normalized.replace(",", ".");
  } else {
    normalized = normalized.replace(/,/g, "");
  }

  return normalized;
}
