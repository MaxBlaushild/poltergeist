import type { ParameterSchema } from '../api/types';

// Mirrors reef-site's lib/derivedBounds.ts shape exactly, so a future
// customer-facing parameter with a real packing constraint (e.g. exposing
// holesPerRow directly) can add an entry here with no changes to
// SchemaForm/Configure. Starts empty for v1: card/tray count is entirely
// assembler-computed (go/pkg/reef/set) from sleeve/box/manifest, not a
// slider the customer ever sees directly.
type BoundFormula = (values: Record<string, unknown>, schema: ParameterSchema) => number;

export const derivedBoundFormulas: Record<string, Record<string, { min?: BoundFormula; max?: BoundFormula }>> = {};
