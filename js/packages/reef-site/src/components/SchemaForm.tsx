import type { ParameterSchema, TankProfile } from '../api/types';

interface Props {
  schema: ParameterSchema;
  values: Record<string, unknown>;
  onChange: (key: string, value: unknown) => void;
  tanks?: TankProfile[];
  errors?: Record<string, string>;
  derived?: Record<string, string>; // read-only computed values (R-4.5), keyed by property name
}

// R-4.4: "A parameter added to the schema must appear in the UI with no
// frontend change." This renders entirely from the schema document fetched
// at runtime — no per-product form is hand-written anywhere in this app.
export default function SchemaForm({ schema, values, onChange, tanks, errors, derived }: Props) {
  const propertyNames = Object.keys(schema.properties);

  return (
    <div className="space-y-6">
      {propertyNames.map((name) => {
        const prop = schema.properties[name];
        const label = prop['x-label'] ?? name;
        const helpText = prop['x-helpText'];
        const diagram = prop['x-diagramAsset'];
        const error = errors?.[name];
        const derivedValue = derived?.[name];

        if (derivedValue !== undefined) {
          return (
            <Field key={name} label={label} helpText={helpText} diagram={diagram} error={error}>
              <div className="rounded border border-reef-teal/20 bg-reef-sand px-3 py-2 text-reef-ink/70">
                {derivedValue} <span className="text-xs">(derived)</span>
              </div>
            </Field>
          );
        }

        if (prop['x-control'] === 'tank-select') {
          return (
            <Field key={name} label={label} helpText={helpText} diagram={diagram} error={error}>
              <select
                className="w-full border border-reef-teal/30 rounded px-3 py-2"
                value={(values[name] as string) ?? ''}
                onChange={(e) => onChange(name, e.target.value || null)}
              >
                <option value="">Other (measure by hand)</option>
                {(tanks ?? []).map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.manufacturer} {t.model}
                  </option>
                ))}
              </select>
            </Field>
          );
        }

        const typeName = normalizeType(prop.type);

        if (typeName === 'boolean') {
          return (
            <Field key={name} label={label} helpText={helpText} diagram={diagram} error={error}>
              <input
                type="checkbox"
                checked={Boolean(values[name])}
                onChange={(e) => onChange(name, e.target.checked)}
                className="h-5 w-5"
              />
            </Field>
          );
        }

        if (prop.enum && prop.enum.length > 0) {
          return (
            <Field key={name} label={label} helpText={helpText} diagram={diagram} error={error}>
              <select
                className="w-full border border-reef-teal/30 rounded px-3 py-2"
                value={String(values[name] ?? prop.default ?? '')}
                onChange={(e) => onChange(name, coerceEnumValue(prop.enum!, e.target.value))}
              >
                {prop.enum.map((option) => (
                  <option key={String(option)} value={String(option)}>
                    {String(option)}
                    {prop['x-unit'] ? prop['x-unit'] : ''}
                  </option>
                ))}
              </select>
            </Field>
          );
        }

        if (typeName === 'number' || typeName === 'integer') {
          return (
            <Field key={name} label={label} helpText={helpText} diagram={diagram} error={error}>
              <div className="flex items-center gap-2">
                <input
                  type="range"
                  min={prop.minimum}
                  max={prop.maximum}
                  step={typeName === 'integer' ? 1 : 0.5}
                  value={Number(values[name] ?? prop.minimum ?? 0)}
                  onChange={(e) => onChange(name, Number(e.target.value))}
                  className="flex-1"
                />
                <span className="w-20 text-right text-sm tabular-nums">
                  {Number(values[name] ?? prop.minimum ?? 0)}
                  {prop['x-unit'] ?? ''}
                </span>
              </div>
              {(prop.minimum !== undefined || prop.maximum !== undefined) && (
                <p className="text-xs text-reef-ink/50 mt-1">
                  Range: {prop.minimum ?? '–'} to {prop.maximum ?? '–'}
                  {prop['x-unit'] ?? ''}
                </p>
              )}
            </Field>
          );
        }

        return (
          <Field key={name} label={label} helpText={helpText} diagram={diagram} error={error}>
            <input
              type="text"
              className="w-full border border-reef-teal/30 rounded px-3 py-2"
              value={String(values[name] ?? '')}
              onChange={(e) => onChange(name, e.target.value)}
            />
          </Field>
        );
      })}
    </div>
  );
}

function Field({
  label,
  helpText,
  diagram,
  error,
  children,
}: {
  label: string;
  helpText?: string;
  diagram?: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <div className="flex items-center justify-between mb-1">
        <label className="text-sm font-medium">{label}</label>
        {diagram && (
          <a href={diagram} target="_blank" rel="noreferrer" className="text-xs text-reef-teal underline">
            where to measure
          </a>
        )}
      </div>
      {children}
      {helpText && <p className="text-xs text-reef-ink/50 mt-1">{helpText}</p>}
      {error && (
        <p className="text-xs text-red-600 mt-1">
          {label}: {error}
        </p>
      )}
    </div>
  );
}

function normalizeType(t: string | (string | null)[]): string {
  if (typeof t === 'string') return t;
  return t.find((v) => v && v !== 'null') ?? '';
}

function coerceEnumValue(enumValues: (string | number)[], raw: string): string | number {
  const numeric = enumValues.find((v) => typeof v === 'number' && String(v) === raw);
  return numeric !== undefined ? numeric : raw;
}
