import type { ParameterSchema } from '../api/types';

// Schema-advertised "minimum"/"maximum" values are static per field, but
// some fields' real limit depends on another field's current value (marked
// in the schema via x-derivedBoundFrom) — holesPerTier's true ceiling
// depends on widthMm and plugHoleDiameterMm, since more/bigger holes need
// more room. Previously that only surfaced as a rejection message after a
// preview request ("holesPerTier 8 is too many for a 150mm-wide rack...");
// this recomputes the real bound client-side so the sliders themselves
// never allow the invalid combination.
//
// The relationship is genuinely bidirectional: shrinking width can make
// EVERY value of holesPerTier invalid, including the schema's own stated
// minimum of 4 (e.g. at the schema's widthMm floor of 60mm with 20mm
// holes, only 2 holes fit — below the 4 holesPerTier requires). So
// widthMm's own *minimum* is raised to whatever width is needed to fit the
// schema's minimum holesPerTier, keeping the two constraints from ever
// fighting each other (holesPerTier always adapts to width, never the
// reverse — matching x-derivedBoundFrom's own direction).
//
// Mirrors go/pkg/reef/generate/frag_rack.go's FragRackMaxHolesPerTier
// exactly — keep both in sync if that geometry constraint ever changes.
const FRAG_MIN_EDGE_WALL_MM = 3.0;

export function fragRackMaxHolesPerTier(widthMm: number, plugHoleDiameterMm: number): number {
  const edgeMargin = plugHoleDiameterMm / 2 + FRAG_MIN_EDGE_WALL_MM;
  const usable = widthMm - 2 * edgeMargin;
  if (usable <= 0) return 1;
  const minSpacing = plugHoleDiameterMm + FRAG_MIN_EDGE_WALL_MM;
  const max = Math.floor(usable / minSpacing) + 1;
  return max < 1 ? 1 : max;
}

// Inverse of the above: the minimum widthMm that fits a given holesPerTier.
function fragRackMinWidthForHoles(holesPerTier: number, plugHoleDiameterMm: number): number {
  const edgeMargin = plugHoleDiameterMm / 2 + FRAG_MIN_EDGE_WALL_MM;
  const minSpacing = plugHoleDiameterMm + FRAG_MIN_EDGE_WALL_MM;
  return 2 * edgeMargin + (holesPerTier - 1) * minSpacing;
}

// Mirrors go/pkg/reef/generate/shelf_rack.go's ShelfRackMaxHolesPerRow /
// ShelfRackMaxRows — same packing math as frag rack's, applied
// independently on shelf-rack's width and depth axes (its deck is a 2D grid
// rather than stacked tiers). Kept as its own functions rather than reusing
// fragRackMaxHolesPerTier so each stays a direct 1:1 mirror of its Go
// counterpart — keep all three in sync if the geometry constraint changes.
const SHELF_MIN_EDGE_WALL_MM = 3.0;

export function shelfRackMaxHolesPerRow(widthMm: number, plugHoleDiameterMm: number): number {
  return shelfMaxCountForSpan(widthMm, plugHoleDiameterMm);
}

export function shelfRackMaxRows(depthMm: number, plugHoleDiameterMm: number): number {
  return shelfMaxCountForSpan(depthMm, plugHoleDiameterMm);
}

function shelfMaxCountForSpan(spanMm: number, plugHoleDiameterMm: number): number {
  const edgeMargin = plugHoleDiameterMm / 2 + SHELF_MIN_EDGE_WALL_MM;
  const usable = spanMm - 2 * edgeMargin;
  if (usable <= 0) return 1;
  const minSpacing = plugHoleDiameterMm + SHELF_MIN_EDGE_WALL_MM;
  const max = Math.floor(usable / minSpacing) + 1;
  return max < 1 ? 1 : max;
}

function shelfMinSpanForCount(count: number, plugHoleDiameterMm: number): number {
  const edgeMargin = plugHoleDiameterMm / 2 + SHELF_MIN_EDGE_WALL_MM;
  const minSpacing = plugHoleDiameterMm + SHELF_MIN_EDGE_WALL_MM;
  return 2 * edgeMargin + (count - 1) * minSpacing;
}

type BoundFormula = (values: Record<string, unknown>, schema: ParameterSchema) => number;

// Keyed by product slug -> field name -> { min?, max? } over the current
// values (+ schema, for reading the other field's own static bounds).
// Generic enough that a second derived-bound field (on this product or a
// future one) just adds an entry here, no changes to SchemaForm/Configure.
export const derivedBoundFormulas: Record<string, Record<string, { min?: BoundFormula; max?: BoundFormula }>> = {
  'magnetic-frag-rack': {
    holesPerTier: {
      max: (values) =>
        fragRackMaxHolesPerTier(Number(values.widthMm ?? 0), Number(values.plugHoleDiameterMm ?? 0)),
    },
    widthMm: {
      min: (values, schema) => {
        const schemaMinHoles = schema.properties.holesPerTier?.minimum ?? 4;
        return fragRackMinWidthForHoles(schemaMinHoles, Number(values.plugHoleDiameterMm ?? 0));
      },
    },
  },
  'shelf-rack': {
    holesPerRow: {
      max: (values) =>
        shelfRackMaxHolesPerRow(Number(values.widthMm ?? 0), Number(values.plugHoleDiameterMm ?? 0)),
    },
    widthMm: {
      min: (values, schema) => {
        const schemaMinHoles = schema.properties.holesPerRow?.minimum ?? 4;
        return shelfMinSpanForCount(schemaMinHoles, Number(values.plugHoleDiameterMm ?? 0));
      },
    },
    rowCount: {
      max: (values) =>
        shelfRackMaxRows(Number(values.depthMm ?? 0), Number(values.plugHoleDiameterMm ?? 0)),
    },
    depthMm: {
      min: (values, schema) => {
        const schemaMinRows = schema.properties.rowCount?.minimum ?? 1;
        return shelfMinSpanForCount(schemaMinRows, Number(values.plugHoleDiameterMm ?? 0));
      },
    },
  },
};
