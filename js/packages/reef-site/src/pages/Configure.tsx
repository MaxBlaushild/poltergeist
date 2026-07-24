import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { reefApi, ApiError } from '../api/client';
import type { Configuration, ParameterSchema, PreviewResponse, Product, TankProfile } from '../api/types';
import SchemaForm from '../components/SchemaForm';
import StlViewer from '../components/StlViewer';
import { useCart } from '../hooks/useCart';
import { getSessionId } from '../lib/session';
import { paramsToSearch, searchToParams } from '../lib/paramsUrl';

function defaultValues(schema: ParameterSchema): Record<string, unknown> {
  const values: Record<string, unknown> = {};
  for (const [name, prop] of Object.entries(schema.properties)) {
    if (prop.default !== undefined) values[name] = prop.default;
    else if (prop.minimum !== undefined) values[name] = prop.minimum;
    else if (prop.enum && prop.enum.length > 0) values[name] = prop.enum[0];
  }
  return values;
}

export default function Configure() {
  const { slug } = useParams<{ slug: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const { addItem } = useCart();

  const [product, setProduct] = useState<Product | null>(null);
  const [schema, setSchema] = useState<ParameterSchema | null>(null);
  const [tanks, setTanks] = useState<TankProfile[]>([]);
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

  // Load product + schema (+ tanks, if the schema has a tank-select field).
  useEffect(() => {
    if (!slug) return;
    Promise.all([reefApi.getProduct(slug), reefApi.getProductSchema(slug)]).then(([p, s]) => {
      setProduct(p);
      setSchema(s);
      const fromUrl = searchToParams(s, searchParams);
      setValues(Object.keys(fromUrl).length > 0 ? { ...defaultValues(s), ...fromUrl } : defaultValues(s));
      if (Object.values(s.properties).some((prop) => prop['x-control'] === 'tank-select')) {
        reefApi.listTanks().then(setTanks);
      }
      reefApi.recordEvent('configurator_opened', { sessionId, productSlug: slug });
      if (Object.keys(fromUrl).length > 0) {
        reefApi.recordEvent('share_link_opened', { sessionId, productSlug: slug });
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slug]);

  // Debounced live preview (R-2.6: 300ms), race-safe against out-of-order
  // responses so stale geometry is never shown alongside current params
  // (R-4.6).
  useEffect(() => {
    if (!slug || !schema || Object.keys(values).length === 0) return;
    if (debounceTimer.current) clearTimeout(debounceTimer.current);

    debounceTimer.current = setTimeout(() => {
      const myGeneration = ++requestGeneration.current;
      setPreviewPending(true);
      setPreviewError(null);
      reefApi
        .preview(slug, values, sessionId)
        .then((result) => {
          if (myGeneration !== requestGeneration.current) return; // superseded by a newer request
          setPreview(result);
          setPreviewPending(false);
          reefApi.recordEvent('preview_rendered', { sessionId, productSlug: slug });
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
  }, [slug, schema, values, sessionId]);

  // Poll the async validate job (R-5.1/R-2.10).
  useEffect(() => {
    if (!configuration || configuration.status !== 'pending') return;
    const interval = setInterval(async () => {
      const updated = await reefApi.getConfiguration(configuration.id);
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
      if (key === 'tankProfileId') {
        const tank = tanks.find((t) => t.id === value);
        const autofills = schema?.properties.tankProfileId?.['x-autofills'] ?? [];
        setValues((prev) => {
          const next: Record<string, unknown> = { ...prev, tankProfileId: value };
          if (tank) {
            for (const field of autofills) {
              if (field === 'glassThicknessMm') next.glassThicknessMm = tank.glassThicknessMm;
              if (field === 'rimThicknessMm') next.rimThicknessMm = tank.rimThicknessMm;
              if (field === 'rimWidthMm') next.rimWidthMm = tank.rimWidthMm;
              if (field === 'euroBrace') next.euroBrace = tank.euroBrace;
            }
          }
          return next;
        });
      } else {
        setValues((prev) => ({ ...prev, [key]: value }));
      }
      setConfiguration(null);
      reefApi.recordEvent('parameter_changed', { sessionId, productSlug: slug, metadata: { key } });
    },
    [schema, tanks, sessionId, slug],
  );

  const handleAddToCart = async () => {
    if (!slug) return;
    setValidating(true);
    const result = await reefApi.validate(slug, values, sessionId);
    const cfg = await reefApi.getConfiguration(result.configurationId);
    setConfiguration(cfg);
  };

  useEffect(() => {
    if (configuration?.status === 'valid' && slug) {
      addItem({ productSlug: slug, configurationId: configuration.id, quantity: 1 });
      reefApi.recordEvent('add_to_cart', { sessionId, productSlug: slug, configurationId: configuration.id });
      navigate('/cart');
    }
  }, [configuration, slug, sessionId, addItem, navigate]);

  const handleShare = () => {
    const search = paramsToSearch(values);
    const url = `${window.location.origin}/configure/${slug}?${search}`;
    setSearchParams(search);
    navigator.clipboard?.writeText(url).then(() => {
      setCopyStatus('copied');
      reefApi.recordEvent('share_link_created', { sessionId, productSlug: slug });
      setTimeout(() => setCopyStatus('idle'), 2000);
    });
  };

  if (!product || !schema) return <p>Loading…</p>;

  return (
    <div className="grid md:grid-cols-2 gap-8">
      <div>
        <h1 className="text-2xl font-bold mb-4">{product.name}</h1>
        <StlViewer url={preview?.previewUrl ?? null} pending={previewPending} plateFits={preview?.plateFits ?? true} />

        {preview && (
          <div className="mt-3 text-sm space-y-1">
            <p>
              Bounding box: {preview.bboxMm.xMm.toFixed(0)} × {preview.bboxMm.yMm.toFixed(0)} ×{' '}
              {preview.bboxMm.zMm.toFixed(0)} mm
            </p>
            <p className={preview.plateFits ? 'text-reef-teal' : 'text-red-600'}>
              {preview.plateFits ? '✓ Fits the print envelope' : '✗ Exceeds the print envelope — reduce size'}
            </p>
          </div>
        )}
        {previewError && <p className="mt-2 text-sm text-red-600">{previewError}</p>}

        <button onClick={handleShare} className="mt-4 text-sm text-reef-teal underline">
          {copyStatus === 'copied' ? 'Link copied!' : 'Copy shareable link'}
        </button>
      </div>

      <div>
        <SchemaForm schema={schema} values={values} onChange={handleChange} tanks={tanks} />

        <div className="mt-6">
          {configuration?.status === 'rejected' && (
            <p className="mb-3 rounded bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700">
              {configuration.rejectionReason}
            </p>
          )}
          <button
            onClick={handleAddToCart}
            disabled={validating}
            className="w-full rounded bg-reef-coral px-5 py-3 font-semibold text-reef-ink hover:opacity-90 disabled:opacity-50"
          >
            {validating ? 'Validating…' : 'Add to cart'}
          </button>
          <p className="mt-2 text-xs text-reef-ink/50">
            Adding to cart runs a full server-side slice — this can take up to a minute.
          </p>
        </div>
      </div>
    </div>
  );
}
