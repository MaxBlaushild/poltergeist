import type { ParameterSchema } from '../api/types';

// R-5.4: shareable configurations — same encoding as reef-site's own
// lib/paramsUrl.ts, so a config link can be pasted into a BGG thread and
// reopened, no login required.
export function paramsToSearch(values: Record<string, unknown>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(values)) {
    if (value === null || value === undefined || value === '') continue;
    search.set(key, String(value));
  }
  return search.toString();
}

export function searchToParams(schema: ParameterSchema, search: URLSearchParams): Record<string, unknown> {
  const values: Record<string, unknown> = {};
  for (const [name, prop] of Object.entries(schema.properties)) {
    const raw = search.get(name);
    if (raw === null) continue;
    const typeName = typeof prop.type === 'string' ? prop.type : prop.type.find((t) => t && t !== 'null');
    if (typeName === 'number' || typeName === 'integer') {
      const n = Number(raw);
      if (!Number.isNaN(n)) values[name] = n;
    } else if (typeName === 'boolean') {
      values[name] = raw === 'true';
    } else {
      values[name] = raw;
    }
  }
  return values;
}
