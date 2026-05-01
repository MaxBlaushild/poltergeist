import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import ContentDashboard from './ContentDashboard.tsx';
import { useZoneKinds, zoneKindLabel } from './zoneKindHelpers.ts';

type ShrineEffectKind =
  | 'strength'
  | 'dexterity'
  | 'constitution'
  | 'intelligence'
  | 'wisdom'
  | 'charisma'
  | 'health_regen'
  | 'mana_regen'
  | 'physical_damage'
  | 'arcane_damage'
  | 'holy_damage'
  | 'shadow_damage'
  | 'fire_resistance'
  | 'ice_resistance'
  | 'lightning_resistance'
  | 'poison_resistance'
  | 'physical_resistance'
  | 'warding';

type ShrineTemplateRecord = {
  id: string;
  zoneKind?: string;
  name: string;
  description: string;
  blessingName: string;
  effectDescription: string;
  effectKind: ShrineEffectKind;
  baseMagnitude: number;
  createdAt?: string;
  updatedAt?: string;
};

type ShrineTemplateGenerationJob = {
  id: string;
  zoneKind?: string;
  status: string;
  count: number;
  createdCount: number;
  errorMessage?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type ShrineTemplateFormState = {
  zoneKind: string;
  name: string;
  description: string;
  blessingName: string;
  effectDescription: string;
  effectKind: ShrineEffectKind;
  baseMagnitude: string;
};

type ShrineTemplateGenerationFormState = {
  count: string;
  zoneKind: string;
};

const effectKindOptions: Array<{
  value: ShrineEffectKind;
  label: string;
  hint: string;
}> = [
  { value: 'strength', label: 'Strength', hint: 'Power and force' },
  { value: 'dexterity', label: 'Dexterity', hint: 'Agility and finesse' },
  {
    value: 'constitution',
    label: 'Constitution',
    hint: 'Durability and grit',
  },
  {
    value: 'intelligence',
    label: 'Intelligence',
    hint: 'Arcane focus and insight',
  },
  { value: 'wisdom', label: 'Wisdom', hint: 'Clarity and perception' },
  { value: 'charisma', label: 'Charisma', hint: 'Presence and command' },
  {
    value: 'health_regen',
    label: 'Health Regen',
    hint: 'Steady vitality recovery',
  },
  {
    value: 'mana_regen',
    label: 'Mana Regen',
    hint: 'Steady mana recovery',
  },
  {
    value: 'physical_damage',
    label: 'Physical Damage',
    hint: 'Weapon and impact offense',
  },
  {
    value: 'arcane_damage',
    label: 'Arcane Damage',
    hint: 'Arcane spell output',
  },
  { value: 'holy_damage', label: 'Holy Damage', hint: 'Radiant offense' },
  {
    value: 'shadow_damage',
    label: 'Shadow Damage',
    hint: 'Umbral offense',
  },
  {
    value: 'fire_resistance',
    label: 'Fire Resistance',
    hint: 'Heat and flame defense',
  },
  {
    value: 'ice_resistance',
    label: 'Ice Resistance',
    hint: 'Cold and frost defense',
  },
  {
    value: 'lightning_resistance',
    label: 'Lightning Resistance',
    hint: 'Storm and shock defense',
  },
  {
    value: 'poison_resistance',
    label: 'Poison Resistance',
    hint: 'Toxin defense',
  },
  {
    value: 'physical_resistance',
    label: 'Physical Resistance',
    hint: 'Armor-like resilience',
  },
  {
    value: 'warding',
    label: 'Warding',
    hint: 'Broad incoming damage resistance',
  },
];

const emptyFormState = (): ShrineTemplateFormState => ({
  zoneKind: '',
  name: '',
  description: '',
  blessingName: '',
  effectDescription: '',
  effectKind: 'strength',
  baseMagnitude: '2',
});

const emptyGenerationForm = (): ShrineTemplateGenerationFormState => ({
  count: '6',
  zoneKind: '',
});

const parseInteger = (value: string, fallback = 0) => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const formatDate = (value?: string | null): string => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const effectKindLabel = (value: string): string =>
  effectKindOptions.find((option) => option.value === value)?.label ??
  value
    .split('_')
    .filter(Boolean)
    .map((token) => token.charAt(0).toUpperCase() + token.slice(1))
    .join(' ');

const formFromRecord = (
  record: ShrineTemplateRecord
): ShrineTemplateFormState => ({
  zoneKind: record.zoneKind ?? '',
  name: record.name ?? '',
  description: record.description ?? '',
  blessingName: record.blessingName ?? '',
  effectDescription: record.effectDescription ?? '',
  effectKind: record.effectKind ?? 'strength',
  baseMagnitude: String(record.baseMagnitude ?? 2),
});

const buildPayloadFromForm = (form: ShrineTemplateFormState) => ({
  zoneKind: form.zoneKind.trim(),
  name: form.name.trim(),
  description: form.description.trim(),
  blessingName: form.blessingName.trim(),
  effectDescription: form.effectDescription.trim(),
  effectKind: form.effectKind,
  baseMagnitude: parseInteger(form.baseMagnitude, 2),
});

const statusClassName = (status: string): string => {
  switch (status) {
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-rose-600';
    case 'in_progress':
      return 'bg-amber-600';
    default:
      return 'bg-slate-600';
  }
};

export const ShrineTemplates = () => {
  const { apiClient } = useAPI();
  const { zoneKinds, zoneKindBySlug } = useZoneKinds();
  const [records, setRecords] = useState<ShrineTemplateRecord[]>([]);
  const [jobs, setJobs] = useState<ShrineTemplateGenerationJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<ShrineTemplateRecord | null>(null);
  const [form, setForm] = useState<ShrineTemplateFormState>(emptyFormState());
  const [generationForm, setGenerationForm] =
    useState<ShrineTemplateGenerationFormState>(emptyGenerationForm());

  const load = useCallback(
    async (suppressLoading = false) => {
      try {
        if (!suppressLoading) setLoading(true);
        const [templateResp, jobResp] = await Promise.all([
          apiClient.get<ShrineTemplateRecord[]>('/sonar/shrine-templates'),
          apiClient.get<ShrineTemplateGenerationJob[]>(
            '/sonar/admin/shrine-template-generation-jobs?limit=20'
          ),
        ]);
        setRecords(Array.isArray(templateResp) ? templateResp : []);
        setJobs(Array.isArray(jobResp) ? jobResp : []);
      } catch (error) {
        console.error('Failed to load shrine templates', error);
        alert('Failed to load shrine templates.');
      } finally {
        if (!suppressLoading) setLoading(false);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!jobs.some((job) => ['queued', 'in_progress'].includes(job.status))) {
      return;
    }
    const interval = window.setInterval(() => {
      void load(true);
    }, 5000);
    return () => clearInterval(interval);
  }, [jobs, load]);

  const sortedRecords = useMemo(
    () =>
      [...records].sort((left, right) =>
        (left.name || '').localeCompare(right.name || '')
      ),
    [records]
  );

  const dashboardMetrics = useMemo(() => {
    const zoneBoundCount = records.filter((record) => record.zoneKind?.trim()).length;
    const strongestTemplate = records.reduce(
      (max, record) => Math.max(max, record.baseMagnitude ?? 0),
      0
    );
    const averageMagnitude =
      records.length > 0
        ? (
            records.reduce(
              (sum, record) => sum + (record.baseMagnitude ?? 0),
              0
            ) / records.length
          ).toFixed(1)
        : '0.0';
    return [
      { label: 'Templates', value: records.length },
      {
        label: 'Zone-Tuned',
        value: zoneBoundCount,
        note:
          zoneBoundCount === 0
            ? 'Global templates only'
            : `${records.length - zoneBoundCount} global templates`,
      },
      {
        label: 'Average Magnitude',
        value: averageMagnitude,
        note: 'Base value before level scaling',
      },
      {
        label: 'Strongest Base',
        value: strongestTemplate,
        note: 'Highest base blessing magnitude',
      },
    ];
  }, [records]);

  const dashboardSections = useMemo(() => {
    const byEffect = effectKindOptions
      .map((option) => ({
        label: option.label,
        value: records.filter((record) => record.effectKind === option.value)
          .length,
        note: option.hint,
      }))
      .filter((bucket) => bucket.value > 0);

    const byZoneKind = [
      {
        label: 'Global',
        value: records.filter((record) => !record.zoneKind?.trim()).length,
        note: 'Available to any zone when selected or seeded',
      },
      ...zoneKinds
        .map((zoneKind) => ({
          label: zoneKind.name,
          value: records.filter((record) => record.zoneKind === zoneKind.slug)
            .length,
          note: zoneKind.description?.trim() || zoneKind.slug,
        }))
        .filter((bucket) => bucket.value > 0),
    ];

    const jobBuckets = jobs.slice(0, 6).map((job) => ({
      label: `${zoneKindLabel(job.zoneKind, zoneKindBySlug)} · ${job.count} requested`,
      value: job.createdCount,
      note:
        job.status === 'failed'
          ? job.errorMessage || 'Generation failed.'
          : `${job.status.replaceAll('_', ' ')} · updated ${formatDate(
              job.updatedAt
            )}`,
    }));

    return [
      {
        title: 'Effect Spread',
        note: 'Which blessings are already covered by current templates.',
        buckets: byEffect,
        emptyLabel: 'No shrine templates yet.',
      },
      {
        title: 'Zone Kind Coverage',
        note: 'Use global templates as overflow and zone-kind templates for flavor.',
        buckets: byZoneKind.filter((bucket) => bucket.value > 0),
        emptyLabel: 'No zone kind coverage yet.',
      },
      {
        title: 'Recent Generation Jobs',
        note: 'Queued and recent shrine template jobs.',
        buckets: jobBuckets,
        emptyLabel: 'No shrine template jobs have been queued yet.',
      },
    ];
  }, [jobs, records, zoneKinds, zoneKindBySlug]);

  const closeForm = () => {
    setShowForm(false);
    setEditing(null);
    setForm(emptyFormState());
  };

  const openCreate = () => {
    setEditing(null);
    setForm(emptyFormState());
    setShowForm(true);
  };

  const openEdit = (record: ShrineTemplateRecord) => {
    setEditing(record);
    setForm(formFromRecord(record));
    setShowForm(true);
  };

  const saveTemplate = async () => {
    const payload = buildPayloadFromForm(form);
    if (!payload.name) {
      alert('Template name is required.');
      return;
    }
    if (!payload.blessingName) {
      alert('Blessing name is required.');
      return;
    }
    if (!payload.effectDescription) {
      alert('Effect description is required.');
      return;
    }
    if (payload.baseMagnitude <= 0) {
      alert('Base magnitude must be greater than zero.');
      return;
    }

    try {
      setSaving(true);
      if (editing) {
        await apiClient.put(`/sonar/shrine-templates/${editing.id}`, payload);
      } else {
        await apiClient.post('/sonar/shrine-templates', payload);
      }
      closeForm();
      await load(true);
    } catch (error) {
      console.error('Failed to save shrine template', error);
      alert('Failed to save shrine template.');
    } finally {
      setSaving(false);
    }
  };

  const deleteTemplate = async (record: ShrineTemplateRecord) => {
    if (
      !window.confirm(
        `Delete ${record.name || record.blessingName || 'this shrine template'}?`
      )
    ) {
      return;
    }
    try {
      await apiClient.delete(`/sonar/shrine-templates/${record.id}`);
      if (editing?.id === record.id) {
        closeForm();
      }
      await load(true);
    } catch (error) {
      console.error('Failed to delete shrine template', error);
      alert('Failed to delete shrine template.');
    }
  };

  const queueGenerationJob = async () => {
    const count = parseInteger(generationForm.count, 0);
    if (count <= 0) {
      alert('Generation count must be at least 1.');
      return;
    }
    try {
      setGenerating(true);
      await apiClient.post('/sonar/admin/shrine-template-generation-jobs', {
        count,
        zoneKind: generationForm.zoneKind.trim(),
      });
      setGenerationForm(emptyGenerationForm());
      await load(true);
    } catch (error) {
      console.error('Failed to queue shrine template generation job', error);
      alert('Failed to queue shrine template generation job.');
    } finally {
      setGenerating(false);
    }
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Shrine Templates</h1>
          <p className="mt-2 max-w-3xl text-sm text-gray-600">
            Author reusable blessing shrines, queue new shrine concepts for a
            zone kind, and keep the pool stocked for zone seeding. Each shrine
            instance inherits its blessing identity from the template while
            keeping its own cooldown per player.
          </p>
        </div>
        <div className="flex flex-wrap gap-3">
          <button
            type="button"
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            onClick={() => void load()}
            disabled={loading}
          >
            Refresh
          </button>
          <button
            type="button"
            className="rounded-md bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-700"
            onClick={openCreate}
          >
            New Shrine Template
          </button>
        </div>
      </div>

      <ContentDashboard
        title="Shrine Template Dashboard"
        subtitle="Track effect coverage, zone-kind specialization, and recent generation activity for blessing shrines."
        status={loading ? 'Loading…' : `${records.length} templates loaded`}
        loading={loading}
        metrics={dashboardMetrics}
        sections={dashboardSections}
      />

      <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Generate Shrine Templates</h2>
            <p className="mt-1 text-sm text-gray-600">
              Queue blessing ideas for a specific zone kind or create general
              shrine templates the seeding job can reuse anywhere.
            </p>
          </div>
        </div>
        <div className="mt-4 grid gap-4 md:grid-cols-3">
          <label className="block">
            <span className="mb-1 block text-sm font-medium text-gray-700">
              Zone Kind
            </span>
            <select
              className="w-full rounded-md border border-gray-300 px-3 py-2"
              value={generationForm.zoneKind}
              onChange={(event) =>
                setGenerationForm((current) => ({
                  ...current,
                  zoneKind: event.target.value,
                }))
              }
            >
              <option value="">Global pool</option>
              {zoneKinds.map((zoneKind) => (
                <option key={zoneKind.id} value={zoneKind.slug}>
                  {zoneKind.name}
                </option>
              ))}
            </select>
          </label>
          <label className="block">
            <span className="mb-1 block text-sm font-medium text-gray-700">
              Count
            </span>
            <input
              type="number"
              min={1}
              max={100}
              className="w-full rounded-md border border-gray-300 px-3 py-2"
              value={generationForm.count}
              onChange={(event) =>
                setGenerationForm((current) => ({
                  ...current,
                  count: event.target.value,
                }))
              }
            />
          </label>
          <div className="flex items-end">
            <button
              type="button"
              className="w-full rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
              onClick={() => void queueGenerationJob()}
              disabled={generating}
            >
              {generating ? 'Queueing…' : 'Queue Shrine Ideas'}
            </button>
          </div>
        </div>
      </section>

      <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold">Recent Jobs</h2>
            <p className="mt-1 text-sm text-gray-600">
              The seeding job will use existing shrine templates first, then
              generate more when a zone kind needs additional variety.
            </p>
          </div>
        </div>
        {jobs.length === 0 ? (
          <p className="mt-4 text-sm text-gray-500">
            No shrine template generation jobs yet.
          </p>
        ) : (
          <div className="mt-4 space-y-3">
            {jobs.map((job) => (
              <div
                key={job.id}
                className="rounded-xl border border-gray-200 px-4 py-3"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <div className="text-sm font-semibold text-gray-900">
                      {zoneKindLabel(job.zoneKind, zoneKindBySlug)}
                    </div>
                    <div className="mt-1 text-xs text-gray-500">
                      Requested {job.count} · created {job.createdCount} · last
                      updated {formatDate(job.updatedAt)}
                    </div>
                  </div>
                  <span
                    className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold uppercase tracking-wide text-white ${statusClassName(
                      job.status
                    )}`}
                  >
                    {job.status.replaceAll('_', ' ')}
                  </span>
                </div>
                {job.errorMessage ? (
                  <div className="mt-2 text-sm text-rose-600">
                    {job.errorMessage}
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        )}
      </section>

      <section className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold">Active Templates</h2>
            <p className="mt-1 text-sm text-gray-600">
              Shrine names should read like the blessing they bestow, while the
              effect description explains how the day-long buff scales with the
              player.
            </p>
          </div>
          <div className="text-sm text-gray-500">
            {sortedRecords.length} template{sortedRecords.length === 1 ? '' : 's'}
          </div>
        </div>
        {sortedRecords.length === 0 ? (
          <p className="mt-4 text-sm text-gray-500">
            No shrine templates yet. Create one or queue a generation job to
            start the pool.
          </p>
        ) : (
          <div className="mt-4 grid gap-4 lg:grid-cols-2">
            {sortedRecords.map((record) => (
              <article
                key={record.id}
                className="rounded-2xl border border-gray-200 bg-gradient-to-br from-violet-50 via-white to-sky-50 p-5 shadow-sm"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <div className="text-lg font-semibold text-gray-900">
                      {record.name || 'Untitled Shrine'}
                    </div>
                    <div className="mt-1 text-sm font-medium text-violet-700">
                      {record.blessingName || 'Unnamed blessing'}
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <span className="rounded-full bg-white/90 px-2.5 py-1 text-xs font-semibold text-gray-700 ring-1 ring-gray-200">
                      {effectKindLabel(record.effectKind)}
                    </span>
                    <span className="rounded-full bg-white/90 px-2.5 py-1 text-xs font-semibold text-gray-700 ring-1 ring-gray-200">
                      Base {record.baseMagnitude}
                    </span>
                  </div>
                </div>
                <div className="mt-3 flex flex-wrap gap-2 text-xs text-gray-600">
                  <span className="rounded-full bg-white/90 px-2.5 py-1 ring-1 ring-gray-200">
                    {zoneKindLabel(record.zoneKind, zoneKindBySlug)}
                  </span>
                  <span className="rounded-full bg-white/90 px-2.5 py-1 ring-1 ring-gray-200">
                    Updated {formatDate(record.updatedAt)}
                  </span>
                </div>
                {record.description ? (
                  <p className="mt-4 text-sm leading-6 text-gray-700">
                    {record.description}
                  </p>
                ) : null}
                <div className="mt-4 rounded-xl border border-violet-100 bg-white/90 p-4">
                  <div className="text-xs font-semibold uppercase tracking-wide text-violet-700">
                    Blessing Effect
                  </div>
                  <p className="mt-2 text-sm leading-6 text-gray-700">
                    {record.effectDescription || 'No effect description yet.'}
                  </p>
                </div>
                <div className="mt-4 flex flex-wrap gap-3">
                  <button
                    type="button"
                    className="rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-white"
                    onClick={() => openEdit(record)}
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    className="rounded-md border border-rose-200 px-3 py-2 text-sm font-medium text-rose-600 hover:bg-rose-50"
                    onClick={() => void deleteTemplate(record)}
                  >
                    Delete
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      {showForm ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/55 px-4 py-8">
          <div className="max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-2xl bg-white p-6 shadow-2xl">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">
                  {editing ? 'Edit Shrine Template' : 'Create Shrine Template'}
                </h2>
                <p className="mt-1 text-sm text-gray-600">
                  Define the blessing identity here. Individual shrine map
                  placements inherit this template and track their cooldown
                  separately per user.
                </p>
              </div>
              <button
                type="button"
                className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                onClick={closeForm}
              >
                Close
              </button>
            </div>

            <div className="mt-6 grid gap-4 md:grid-cols-2">
              <label className="block">
                <span className="mb-1 block text-sm font-medium text-gray-700">
                  Template Name
                </span>
                <input
                  type="text"
                  className="w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="Shrine of Vigor"
                  value={form.name}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      name: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block">
                <span className="mb-1 block text-sm font-medium text-gray-700">
                  Blessing Name
                </span>
                <input
                  type="text"
                  className="w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="Vigor's Blessing"
                  value={form.blessingName}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      blessingName: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block">
                <span className="mb-1 block text-sm font-medium text-gray-700">
                  Zone Kind
                </span>
                <select
                  className="w-full rounded-md border border-gray-300 px-3 py-2"
                  value={form.zoneKind}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      zoneKind: event.target.value,
                    }))
                  }
                >
                  <option value="">Global pool</option>
                  {zoneKinds.map((zoneKind) => (
                    <option key={zoneKind.id} value={zoneKind.slug}>
                      {zoneKind.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block">
                <span className="mb-1 block text-sm font-medium text-gray-700">
                  Effect Kind
                </span>
                <select
                  className="w-full rounded-md border border-gray-300 px-3 py-2"
                  value={form.effectKind}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      effectKind: event.target.value as ShrineEffectKind,
                    }))
                  }
                >
                  {effectKindOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block md:col-span-2">
                <span className="mb-1 block text-sm font-medium text-gray-700">
                  Description
                </span>
                <textarea
                  className="min-h-[96px] w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="A weathered shrine that hums with old strength."
                  value={form.description}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      description: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block md:col-span-2">
                <span className="mb-1 block text-sm font-medium text-gray-700">
                  Effect Description
                </span>
                <textarea
                  className="min-h-[110px] w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="Increases Strength for one day, with the bonus scaling by user level."
                  value={form.effectDescription}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      effectDescription: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block">
                <span className="mb-1 block text-sm font-medium text-gray-700">
                  Base Magnitude
                </span>
                <input
                  type="number"
                  min={1}
                  className="w-full rounded-md border border-gray-300 px-3 py-2"
                  value={form.baseMagnitude}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      baseMagnitude: event.target.value,
                    }))
                  }
                />
              </label>
            </div>

            <div className="mt-6 flex flex-wrap justify-end gap-3">
              <button
                type="button"
                className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                onClick={closeForm}
                disabled={saving}
              >
                Cancel
              </button>
              <button
                type="button"
                className="rounded-md bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-700 disabled:cursor-not-allowed disabled:opacity-60"
                onClick={() => void saveTemplate()}
                disabled={saving}
              >
                {saving
                  ? editing
                    ? 'Saving…'
                    : 'Creating…'
                  : editing
                    ? 'Save Changes'
                    : 'Create Template'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

export default ShrineTemplates;
