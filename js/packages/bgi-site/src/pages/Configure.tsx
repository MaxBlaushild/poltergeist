import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { bgiApi, ApiError } from '../api/client';
import type { BoxProfile, Configuration, GameDetail, ParameterSchema, PreviewResponse, SleeveProfile } from '../api/types';
import SchemaForm from '../components/SchemaForm';
import StackedStlViewer from '../components/StackedStlViewer';
import { useCart } from '../hooks/useCart';
import { getSessionId } from '../lib/session';
import { paramsToSearch, searchToParams } from '../lib/paramsUrl';
import { derivedBoundFormulas } from '../lib/derivedBounds';

function defaultValues(schema: ParameterSchema, boxProfiles: BoxProfile[]): Record<string, unknown> {
  const values: Record<string, unknown> = {};
  for (const [name, prop] of Object.entries(schema.properties)) {
    if (prop.default !== undefined) values[name] = prop.default;
    else if (prop.minimum !== undefined) values[name] = prop.minimum;
    else if (prop.enum && prop.enum.length > 0) values[name] = prop.enum[0];
  }
  // boxProfileId has no static default in the schema (the set of boxes is
  // per-game data, not known at schema-authoring time) — default to the
  // original box if one exists, same reasoning reef's tank-select leaves
  // to the customer but this product needs a starting box to preview at all.
  if (values.boxProfileId === undefined) {
    const original = boxProfiles.find((b) => b.source === 'original') ?? boxProfiles[0];
    if (original) values.boxProfileId = original.id;
  }
  return values;
}

export default function Configure() {
  const { slug } = useParams<{ slug: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const { addItem } = useCart();

  const [game, setGame] = useState<GameDetail | null>(null);
  const [schema, setSchema] = useState<ParameterSchema | null>(null);
  const [sleeveProfiles, setSleeveProfiles] = useState<SleeveProfile[]>([]);
  const [values, setValues] = useState<Record<string, unknown>>({});
  const [preview, setPreview] = useState<PreviewResponse | null>(null);
  const [previewPending, setPreviewPending] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [validating, setValidating] = useState(false);
  const [configuration, setConfiguration] = useState<Configuration | null>(null);
  const [copyStatus, setCopyStatus] = useState<'idle' | 'copied'>('idle');

  const sessionId = useMemo(() => getSessionId(), []);
  const requestGeneration = useRef(0);
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Same generic, per-field derived-bound clamping as reef's Configure.tsx
  // — empty for bgi in v1 (see lib/derivedBounds.ts), kept so a future
  // customer-facing packing constraint needs no changes here.
  const derivedBounds = useMemo(() => {
    if (!slug || !schema) return undefined;
    const formulas = derivedBoundFormulas[slug];
    if (!formulas) return undefined;
    const max: Record<string, number> = {};
    const min: Record<string, number> = {};
    for (const [field, bound] of Object.entries(formulas)) {
      if (bound.max) max[field] = bound.max(values, schema);
      if (bound.min) min[field] = bound.min(values, schema);
    }
    return { min, max };
  }, [slug, schema, values]);

  useEffect(() => {
    if (!derivedBounds) return;
    setValues((prev) => {
      let changed = false;
      const next = { ...prev };
      for (const [field, max] of Object.entries(derivedBounds.max)) {
        const current = Number(next[field]);
        if (Number.isFinite(current) && current > max) {
          next[field] = max;
          changed = true;
        }
      }
      for (const [field, min] of Object.entries(derivedBounds.min)) {
        const current = Number(next[field]);
        if (Number.isFinite(current) && current < min) {
          next[field] = min;
          changed = true;
        }
      }
      return changed ? next : prev;
    });
  }, [derivedBounds]);

  // Load game (+ its box/sleeve profiles) + schema.
  useEffect(() => {
    if (!slug) return;
    Promise.all([bgiApi.getGame(slug), bgiApi.getGameSchema(slug), bgiApi.listSleeveProfiles()]).then(
      ([g, s, sleeves]) => {
        setGame(g);
        setSchema(s);
        setSleeveProfiles(sleeves);
        const fromUrl = searchToParams(s, searchParams);
        setValues(Object.keys(fromUrl).length > 0 ? { ...defaultValues(s, g.boxProfiles), ...fromUrl } : defaultValues(s, g.boxProfiles));
        bgiApi.recordEvent('configurator_opened', { sessionId, gameSlug: slug });
        if (Object.keys(fromUrl).length > 0) {
          bgiApi.recordEvent('share_link_opened', { sessionId, gameSlug: slug });
        }
      },
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slug]);

  // Debounced live preview (same 300ms/race-safe pattern as reef's own).
  useEffect(() => {
    if (!game?.productSlug || !schema || Object.keys(values).length === 0) return;
    if (!values.sleeveProfileId || !values.boxProfileId) return;
    if (debounceTimer.current) clearTimeout(debounceTimer.current);

    debounceTimer.current = setTimeout(() => {
      const myGeneration = ++requestGeneration.current;
      setPreviewPending(true);
      setPreviewError(null);
      bgiApi
        .preview(game.productSlug!, values, sessionId)
        .then((result) => {
          if (myGeneration !== requestGeneration.current) return;
          setPreview(result);
          setPreviewPending(false);
          bgiApi.recordEvent('preview_rendered', { sessionId, gameSlug: slug });
          bgiApi.recordEvent('fit_indicator_shown', {
            sessionId,
            gameSlug: slug,
            metadata: {
              assembledHeightMm: result.assembledHeightMm,
              boxInteriorDepthMm: result.boxInteriorDepthMm,
              fits: result.fitsBox,
              boxVerified: result.boxVerified,
              depthIsPlaceholder: result.depthIsPlaceholder,
            },
          });
        })
        .catch((e) => {
          if (myGeneration !== requestGeneration.current) return;
          setPreviewPending(false);
          setPreviewError(e instanceof ApiError ? e.message : 'Preview failed');
        });
    }, 300);

    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current);
    };
  }, [game, schema, values, sessionId, slug]);

  // Poll the async full-set generation job.
  useEffect(() => {
    if (!configuration || configuration.status !== 'pending') return;
    const interval = setInterval(async () => {
      const updated = await bgiApi.getConfiguration(configuration.id);
      setConfiguration(updated);
      if (updated.status !== 'pending') {
        setValidating(false);
        clearInterval(interval);
      }
    }, 1500);
    return () => clearInterval(interval);
  }, [configuration]);

  const handleChange = useCallback(
    (key: string, value: unknown) => {
      setValues((prev) => ({ ...prev, [key]: value }));
      setConfiguration(null);
      if (key === 'sleeveProfileId') bgiApi.recordEvent('sleeve_selected', { sessionId, gameSlug: slug });
      else if (key === 'boxProfileId') bgiApi.recordEvent('box_selected', { sessionId, gameSlug: slug });
      else bgiApi.recordEvent('parameter_changed', { sessionId, gameSlug: slug, metadata: { key } });
    },
    [sessionId, slug],
  );

  const handleAddToCart = async () => {
    if (!game?.productSlug) return;
    setValidating(true);
    const result = await bgiApi.validate(game.productSlug, values, sessionId);
    const cfg = await bgiApi.getConfiguration(result.configurationId);
    setConfiguration(cfg);
  };

  useEffect(() => {
    if (configuration?.status === 'valid' && game?.productSlug) {
      addItem({ productSlug: game.productSlug, configurationId: configuration.id, quantity: 1 });
      bgiApi.recordEvent('add_to_cart', { sessionId, gameSlug: slug, configurationId: configuration.id });
      navigate('/cart');
    }
    if (configuration?.status === 'rejected') {
      bgiApi.recordEvent('fit_check_failed', {
        sessionId,
        gameSlug: slug,
        configurationId: configuration.id,
        rule: 'set_validation',
        metadata: { reason: configuration.rejectionReason },
      });
    }
  }, [configuration, game, slug, sessionId, addItem, navigate]);

  const handleShare = () => {
    const search = paramsToSearch(values);
    const url = `${window.location.origin}/configure/${slug}?${search}`;
    setSearchParams(search);
    navigator.clipboard?.writeText(url).then(() => {
      setCopyStatus('copied');
      bgiApi.recordEvent('share_link_created', { sessionId, gameSlug: slug });
      setTimeout(() => setCopyStatus('idle'), 2000);
    });
  };

  if (!game || !schema) return <p className="text-bgi-ink/60">Loading…</p>;

  const previewTrays =
    configuration?.status === 'valid' && configuration.trays
      ? configuration.trays.map((t) => ({ url: t.stlUrl ?? '', heightMm: t.heightMm }))
      : preview?.firstTrayPreviewUrl
        ? [{ url: preview.firstTrayPreviewUrl, heightMm: preview.assembledHeightMm }]
        : [];

  const needsVerificationNotice =
    preview && (!preview.boxVerified || preview.depthIsPlaceholder);

  return (
    <div className="grid gap-8 md:grid-cols-2">
      <div>
        <h1 className="font-display mb-4 text-2xl font-bold text-bgi-ink">{game.name} Tray Set</h1>
        <StackedStlViewer trays={previewTrays} pending={previewPending} fitsBox={preview?.fitsBox ?? true} />

        {preview && (
          <div className="mt-3 space-y-1 text-sm">
            <p className="text-bgi-ink/70">
              Assembled height: {preview.assembledHeightMm.toFixed(0)}mm of {preview.boxInteriorDepthMm.toFixed(0)}mm
              box depth
              {preview.trayCount > 1 ? ` (${preview.trayCount} trays)` : ''}
            </p>
            <p className={preview.fitsBox ? 'font-medium text-bgi-teal' : 'font-medium text-red-600'}>
              {preview.fitsBox ? '✓ Fits the box, lid flush' : "✗ Won't fit — try a thinner sleeve class or a deeper box"}
            </p>
            {preview.unassembledComponents && preview.unassembledComponents.length > 0 && (
              <p className="text-bgi-coral">
                No tray design yet for: {preview.unassembledComponents.join(', ')} — these won't be included in this
                set.
              </p>
            )}
          </div>
        )}
        {previewError && <p className="mt-2 text-sm text-red-600">{previewError}</p>}

        {needsVerificationNotice && (
          <p className="banner-caution mt-3">
            Box and/or sleeve dimensions for this game are sourced from published references and not yet physically
            verified. See{' '}
            <Link to={`/games/${slug}/compatibility`} className="underline decoration-bgi-coral/50 underline-offset-2">
              compatibility notes
            </Link>
            .
          </p>
        )}

        <button
          onClick={handleShare}
          className="mt-4 text-sm font-medium text-bgi-teal underline decoration-bgi-teal/40 underline-offset-2 hover:text-bgi-coral"
        >
          {copyStatus === 'copied' ? 'Link copied!' : 'Copy shareable link'}
        </button>
      </div>

      <div className="card p-6">
        <SchemaForm
          schema={schema}
          values={values}
          onChange={handleChange}
          sleeveProfiles={sleeveProfiles}
          boxProfiles={game.boxProfiles}
          derivedMax={derivedBounds?.max}
          derivedMin={derivedBounds?.min}
        />

        <div className="mt-6">
          {configuration?.status === 'rejected' && (
            <p className="mb-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {configuration.rejectionReason}
            </p>
          )}
          <button onClick={handleAddToCart} disabled={validating} className="btn-primary w-full">
            {validating ? 'Validating…' : 'Add to cart'}
          </button>
          <p className="mt-2 text-xs text-bgi-ink/50">
            Adding to cart generates and slices every tray in the set — this can take a few minutes.
          </p>
        </div>
      </div>
    </div>
  );
}
