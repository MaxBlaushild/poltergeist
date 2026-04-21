import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { ZoneAdminSummary, ZoneKind } from '@poltergeist/types';
import { Link } from 'react-router-dom';

type ZoneKindRatioField = {
  key:
    | 'placeCountRatio'
    | 'monsterCountRatio'
    | 'bossEncounterCountRatio'
    | 'raidEncounterCountRatio'
    | 'inputEncounterCountRatio'
    | 'optionEncounterCountRatio'
    | 'treasureChestCountRatio'
    | 'healingFountainCountRatio'
    | 'resourceCountRatio';
  label: string;
  description: string;
};

type ZoneKindFormState = {
  name: string;
  slug: string;
  description: string;
  placeCountRatio: string;
  monsterCountRatio: string;
  bossEncounterCountRatio: string;
  raidEncounterCountRatio: string;
  inputEncounterCountRatio: string;
  optionEncounterCountRatio: string;
  treasureChestCountRatio: string;
  healingFountainCountRatio: string;
  resourceCountRatio: string;
};

const ratioFields: ZoneKindRatioField[] = [
  {
    key: 'placeCountRatio',
    label: 'Places',
    description: 'POIs and place-led encounters',
  },
  {
    key: 'monsterCountRatio',
    label: 'Monsters',
    description: 'Standard encounters',
  },
  {
    key: 'bossEncounterCountRatio',
    label: 'Bosses',
    description: 'Boss encounter density',
  },
  {
    key: 'raidEncounterCountRatio',
    label: 'Raids',
    description: 'Group raid encounter density',
  },
  {
    key: 'inputEncounterCountRatio',
    label: 'Input scenarios',
    description: 'Open-ended scenario prompts',
  },
  {
    key: 'optionEncounterCountRatio',
    label: 'Option scenarios',
    description: 'Choice-driven scenarios',
  },
  {
    key: 'treasureChestCountRatio',
    label: 'Treasure chests',
    description: 'Chest reward density',
  },
  {
    key: 'healingFountainCountRatio',
    label: 'Healing fountains',
    description: 'Restorative nodes',
  },
  {
    key: 'resourceCountRatio',
    label: 'Resources',
    description: 'Gatherable node density',
  },
];

const emptyForm = (): ZoneKindFormState => ({
  name: '',
  slug: '',
  description: '',
  placeCountRatio: '1',
  monsterCountRatio: '1',
  bossEncounterCountRatio: '1',
  raidEncounterCountRatio: '1',
  inputEncounterCountRatio: '1',
  optionEncounterCountRatio: '1',
  treasureChestCountRatio: '1',
  healingFountainCountRatio: '1',
  resourceCountRatio: '1',
});

const normalizeSlugDraft = (value: string) =>
  value
    .trim()
    .toLowerCase()
    .replace(/[_\s]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');

const formatRatio = (value: number) =>
  `${value.toFixed(Number.isInteger(value) ? 1 : 2)}x`;

const zoneSearchText = (zone: ZoneAdminSummary) =>
  [
    zone.name,
    zone.description,
    zone.kind,
    zone.importMetroName || '',
    ...(zone.internalTags || []),
  ]
    .join(' ')
    .toLowerCase();

const formFromZoneKind = (zoneKind: ZoneKind): ZoneKindFormState => ({
  name: zoneKind.name,
  slug: zoneKind.slug,
  description: zoneKind.description || '',
  placeCountRatio: String(zoneKind.placeCountRatio ?? 1),
  monsterCountRatio: String(zoneKind.monsterCountRatio ?? 1),
  bossEncounterCountRatio: String(zoneKind.bossEncounterCountRatio ?? 1),
  raidEncounterCountRatio: String(zoneKind.raidEncounterCountRatio ?? 1),
  inputEncounterCountRatio: String(zoneKind.inputEncounterCountRatio ?? 1),
  optionEncounterCountRatio: String(zoneKind.optionEncounterCountRatio ?? 1),
  treasureChestCountRatio: String(zoneKind.treasureChestCountRatio ?? 1),
  healingFountainCountRatio: String(zoneKind.healingFountainCountRatio ?? 1),
  resourceCountRatio: String(zoneKind.resourceCountRatio ?? 1),
});

const parseZoneKindForm = (
  form: ZoneKindFormState
): {
  payload?: Omit<ZoneKind, 'id' | 'createdAt' | 'updatedAt'>;
  error?: string;
} => {
  const name = form.name.trim();
  if (!name) {
    return { error: 'Name is required.' };
  }

  const payload: Omit<ZoneKind, 'id' | 'createdAt' | 'updatedAt'> = {
    name,
    slug: normalizeSlugDraft(form.slug || name),
    description: form.description.trim(),
    placeCountRatio: 1,
    monsterCountRatio: 1,
    bossEncounterCountRatio: 1,
    raidEncounterCountRatio: 1,
    inputEncounterCountRatio: 1,
    optionEncounterCountRatio: 1,
    treasureChestCountRatio: 1,
    healingFountainCountRatio: 1,
    resourceCountRatio: 1,
  };

  for (const field of ratioFields) {
    const parsed = Number.parseFloat(form[field.key]);
    if (!Number.isFinite(parsed) || parsed < 0) {
      return {
        error: `${field.label} ratio must be a number greater than or equal to 0.`,
      };
    }
    payload[field.key] = parsed;
  }

  return { payload };
};

export const ZoneKinds = () => {
  const { apiClient } = useAPI();
  const [zoneKinds, setZoneKinds] = useState<ZoneKind[]>([]);
  const [zones, setZones] = useState<ZoneAdminSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [assigningZoneId, setAssigningZoneId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [zoneSearch, setZoneSearch] = useState('');
  const [createForm, setCreateForm] = useState<ZoneKindFormState>(emptyForm());
  const [createSlugTouched, setCreateSlugTouched] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<ZoneKindFormState>(emptyForm());
  const [editSlugTouched, setEditSlugTouched] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [zoneKindsResponse, zonesResponse] = await Promise.all([
        apiClient.get<ZoneKind[]>('/sonar/zoneKinds'),
        apiClient.get<ZoneAdminSummary[]>('/sonar/admin/zones'),
      ]);
      setZoneKinds(zoneKindsResponse);
      setZones(zonesResponse);
      setError(null);
    } catch (err) {
      console.error('Failed to load zone kinds page data', err);
      setError('Unable to load zone kinds right now.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const zoneKindBySlug = useMemo(() => {
    const next = new Map<string, ZoneKind>();
    zoneKinds.forEach((zoneKind) => next.set(zoneKind.slug, zoneKind));
    return next;
  }, [zoneKinds]);

  const assignedZonesByKind = useMemo(() => {
    const next = new Map<string, ZoneAdminSummary[]>();
    zoneKinds.forEach((zoneKind) => next.set(zoneKind.slug, []));
    zones.forEach((zone) => {
      if (!zone.kind) {
        return;
      }
      const current = next.get(zone.kind) || [];
      current.push(zone);
      next.set(zone.kind, current);
    });
    return next;
  }, [zoneKinds, zones]);

  const filteredZones = useMemo(() => {
    const normalizedQuery = zoneSearch.trim().toLowerCase();
    if (!normalizedQuery) {
      return zones;
    }
    return zones.filter((zone) =>
      zoneSearchText(zone).includes(normalizedQuery)
    );
  }, [zoneSearch, zones]);

  const setCreateField = <K extends keyof ZoneKindFormState>(
    key: K,
    value: ZoneKindFormState[K]
  ) => {
    setCreateForm((prev) => ({ ...prev, [key]: value }));
  };

  const setEditField = <K extends keyof ZoneKindFormState>(
    key: K,
    value: ZoneKindFormState[K]
  ) => {
    setEditForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleCreate = async () => {
    const { payload, error: parseError } = parseZoneKindForm(createForm);
    if (!payload) {
      setError(parseError || 'Unable to build zone kind payload.');
      return;
    }

    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      const created = await apiClient.post<ZoneKind>(
        '/sonar/zoneKinds',
        payload
      );
      setZoneKinds((prev) =>
        [...prev, created].sort((a, b) => a.name.localeCompare(b.name))
      );
      setCreateForm(emptyForm());
      setCreateSlugTouched(false);
      setSuccess(`Created ${created.name}.`);
    } catch (err) {
      console.error('Failed to create zone kind', err);
      setError('Unable to create this zone kind.');
    } finally {
      setSaving(false);
    }
  };

  const startEditing = (zoneKind: ZoneKind) => {
    setEditingId(zoneKind.id);
    setEditForm(formFromZoneKind(zoneKind));
    setEditSlugTouched(true);
    setError(null);
    setSuccess(null);
  };

  const cancelEditing = () => {
    setEditingId(null);
    setEditForm(emptyForm());
    setEditSlugTouched(false);
  };

  const handleUpdate = async () => {
    if (!editingId) {
      return;
    }
    const editingZoneKind = zoneKinds.find(
      (zoneKind) => zoneKind.id === editingId
    );
    const { payload, error: parseError } = parseZoneKindForm(editForm);
    if (!payload) {
      setError(parseError || 'Unable to build zone kind payload.');
      return;
    }

    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      const updated = await apiClient.patch<ZoneKind>(
        `/sonar/zoneKinds/${editingId}`,
        payload
      );
      setZoneKinds((prev) =>
        prev
          .map((zoneKind) => (zoneKind.id === updated.id ? updated : zoneKind))
          .sort((a, b) => a.name.localeCompare(b.name))
      );
      setZones((prev) =>
        prev.map((zone) =>
          zone.kind === editingZoneKind?.slug && zone.kind !== updated.slug
            ? { ...zone, kind: updated.slug }
            : zone
        )
      );
      setSuccess(`Updated ${updated.name}.`);
      cancelEditing();
    } catch (err) {
      console.error('Failed to update zone kind', err);
      setError('Unable to update this zone kind.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (zoneKind: ZoneKind) => {
    const confirmed = window.confirm(
      `Delete ${zoneKind.name}? Any zones using ${zoneKind.slug} will be cleared.`
    );
    if (!confirmed) {
      return;
    }

    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.delete(`/sonar/zoneKinds/${zoneKind.id}`);
      setZoneKinds((prev) => prev.filter((entry) => entry.id !== zoneKind.id));
      setZones((prev) =>
        prev.map((zone) =>
          zone.kind === zoneKind.slug ? { ...zone, kind: '' } : zone
        )
      );
      if (editingId === zoneKind.id) {
        cancelEditing();
      }
      setSuccess(`Deleted ${zoneKind.name}.`);
    } catch (err) {
      console.error('Failed to delete zone kind', err);
      setError('Unable to delete this zone kind.');
    } finally {
      setSaving(false);
    }
  };

  const assignZoneKind = async (zoneId: string, kind: string) => {
    setAssigningZoneId(zoneId);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.post('/sonar/zoneKinds/assign-zones', {
        zoneIds: [zoneId],
        kind,
      });
      setZones((prev) =>
        prev.map((zone) => (zone.id === zoneId ? { ...zone, kind } : zone))
      );
      const assignedName =
        kind === ''
          ? 'Cleared zone kind.'
          : `Assigned ${zoneKindBySlug.get(kind)?.name ?? kind}.`;
      setSuccess(assignedName);
    } catch (err) {
      console.error('Failed to assign zone kind', err);
      setError('Unable to update that zone assignment.');
    } finally {
      setAssigningZoneId(null);
    }
  };

  if (loading) {
    return <div className="p-6 text-gray-500">Loading zone kinds...</div>;
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Zone Kinds</h1>
          <p className="mt-1 max-w-3xl text-sm text-gray-500">
            Create reusable zone presets, tune the auto-seed ratios for every
            generated content type, and assign those presets directly to zones.
            A ratio of 1.0 keeps the baseline area recommendation, 2.0 doubles
            it, and 0.5 halves it.
          </p>
        </div>
        <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          <div className="font-medium">{zoneKinds.length} zone kinds</div>
          <div>{zones.filter((zone) => zone.kind).length} assigned zones</div>
        </div>
      </div>

      {(error || success) && (
        <div className="space-y-2">
          {error && (
            <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {error}
            </div>
          )}
          {success && (
            <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
              {success}
            </div>
          )}
        </div>
      )}

      <div className="grid gap-6 xl:grid-cols-[minmax(320px,420px)_minmax(0,1fr)]">
        <section className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">
                Create zone kind
              </h2>
              <p className="mt-1 text-sm text-gray-500">
                Start with a readable label, then tune how heavily each content
                category should be represented in auto-seeding.
              </p>
            </div>
          </div>

          <div className="mt-4 space-y-4">
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                Name
              </label>
              <input
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                value={createForm.name}
                onChange={(event) => {
                  const nextName = event.target.value;
                  setCreateField('name', nextName);
                  if (!createSlugTouched) {
                    setCreateField('slug', normalizeSlugDraft(nextName));
                  }
                }}
                placeholder="Forest"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                Slug
              </label>
              <input
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                value={createForm.slug}
                onChange={(event) => {
                  setCreateSlugTouched(true);
                  setCreateField(
                    'slug',
                    normalizeSlugDraft(event.target.value)
                  );
                }}
                placeholder="forest"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                Description
              </label>
              <textarea
                className="min-h-[88px] w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                value={createForm.description}
                onChange={(event) =>
                  setCreateField('description', event.target.value)
                }
                placeholder="High-herbalism wilderness with more beasts, shrines, and restorative nodes."
              />
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              {ratioFields.map((field) => (
                <label key={field.key} className="block">
                  <span className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                    {field.label}
                  </span>
                  <input
                    type="number"
                    min="0"
                    step="0.1"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                    value={createForm[field.key]}
                    onChange={(event) =>
                      setCreateField(field.key, event.target.value)
                    }
                  />
                  <span className="mt-1 block text-[11px] text-gray-500">
                    {field.description}
                  </span>
                </label>
              ))}
            </div>

            <button
              type="button"
              className="w-full rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
              onClick={handleCreate}
              disabled={saving}
            >
              {saving ? 'Saving...' : 'Create Zone Kind'}
            </button>
          </div>
        </section>

        <section className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">
                Zone kind library
              </h2>
              <p className="mt-1 text-sm text-gray-500">
                These presets are available anywhere we assign or seed a zone.
              </p>
            </div>
          </div>

          {zoneKinds.length === 0 ? (
            <div className="mt-4 rounded-xl border border-dashed border-gray-300 bg-gray-50 px-4 py-10 text-center text-sm text-gray-500">
              No zone kinds yet. Create one to start defining reusable seeding
              identities.
            </div>
          ) : (
            <div className="mt-4 space-y-4">
              {zoneKinds.map((zoneKind) => {
                const assignedZones =
                  assignedZonesByKind.get(zoneKind.slug) || [];
                const isEditing = editingId === zoneKind.id;

                return (
                  <article
                    key={zoneKind.id}
                    className="rounded-2xl border border-gray-200 bg-gray-50 p-4"
                  >
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                      <div>
                        <div className="flex flex-wrap items-center gap-2">
                          <h3 className="text-base font-semibold text-gray-900">
                            {zoneKind.name}
                          </h3>
                          <span className="rounded-full bg-slate-900 px-2.5 py-1 text-[11px] font-medium uppercase tracking-wide text-white">
                            {zoneKind.slug}
                          </span>
                          <span className="rounded-full bg-white px-2.5 py-1 text-[11px] font-medium text-slate-600">
                            {assignedZones.length} assigned zone
                            {assignedZones.length === 1 ? '' : 's'}
                          </span>
                        </div>
                        {zoneKind.description && (
                          <p className="mt-2 text-sm text-gray-600">
                            {zoneKind.description}
                          </p>
                        )}
                      </div>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          className="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-slate-400 hover:text-slate-900"
                          onClick={() =>
                            isEditing ? cancelEditing() : startEditing(zoneKind)
                          }
                        >
                          {isEditing ? 'Cancel' : 'Edit'}
                        </button>
                        <button
                          type="button"
                          className="rounded-lg border border-red-200 bg-red-50 px-3 py-1.5 text-sm font-medium text-red-700 transition hover:border-red-300 hover:bg-red-100"
                          onClick={() => handleDelete(zoneKind)}
                          disabled={saving}
                        >
                          Delete
                        </button>
                      </div>
                    </div>

                    <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                      {ratioFields.map((field) => (
                        <div
                          key={`${zoneKind.id}-${field.key}`}
                          className="rounded-xl border border-white bg-white px-3 py-3 shadow-sm"
                        >
                          <div className="text-xs font-medium uppercase tracking-wide text-gray-500">
                            {field.label}
                          </div>
                          <div className="mt-1 text-lg font-semibold text-gray-900">
                            {formatRatio(zoneKind[field.key])}
                          </div>
                          <div className="mt-1 text-xs text-gray-500">
                            {field.description}
                          </div>
                        </div>
                      ))}
                    </div>

                    {assignedZones.length > 0 && (
                      <div className="mt-4 text-sm text-gray-600">
                        Assigned zones:{' '}
                        {assignedZones
                          .slice(0, 4)
                          .map((zone) => zone.name)
                          .join(', ')}
                        {assignedZones.length > 4
                          ? ` +${assignedZones.length - 4} more`
                          : ''}
                      </div>
                    )}

                    {isEditing && (
                      <div className="mt-4 rounded-2xl border border-slate-200 bg-white p-4">
                        <div className="grid gap-4">
                          <div className="grid gap-4 md:grid-cols-2">
                            <div>
                              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                                Name
                              </label>
                              <input
                                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                                value={editForm.name}
                                onChange={(event) => {
                                  const nextName = event.target.value;
                                  setEditField('name', nextName);
                                  if (!editSlugTouched) {
                                    setEditField(
                                      'slug',
                                      normalizeSlugDraft(nextName)
                                    );
                                  }
                                }}
                              />
                            </div>
                            <div>
                              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                                Slug
                              </label>
                              <input
                                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                                value={editForm.slug}
                                onChange={(event) => {
                                  setEditSlugTouched(true);
                                  setEditField(
                                    'slug',
                                    normalizeSlugDraft(event.target.value)
                                  );
                                }}
                              />
                            </div>
                          </div>

                          <div>
                            <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                              Description
                            </label>
                            <textarea
                              className="min-h-[88px] w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                              value={editForm.description}
                              onChange={(event) =>
                                setEditField('description', event.target.value)
                              }
                            />
                          </div>

                          <div className="grid gap-3 sm:grid-cols-2">
                            {ratioFields.map((field) => (
                              <label
                                key={`${zoneKind.id}-edit-${field.key}`}
                                className="block"
                              >
                                <span className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                                  {field.label}
                                </span>
                                <input
                                  type="number"
                                  min="0"
                                  step="0.1"
                                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                                  value={editForm[field.key]}
                                  onChange={(event) =>
                                    setEditField(field.key, event.target.value)
                                  }
                                />
                              </label>
                            ))}
                          </div>

                          <div className="flex justify-end gap-2">
                            <button
                              type="button"
                              className="rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 transition hover:border-slate-400 hover:text-slate-900"
                              onClick={cancelEditing}
                            >
                              Cancel
                            </button>
                            <button
                              type="button"
                              className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800"
                              onClick={handleUpdate}
                              disabled={saving}
                            >
                              Save changes
                            </button>
                          </div>
                        </div>
                      </div>
                    )}
                  </article>
                );
              })}
            </div>
          )}
        </section>
      </div>

      <section className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">
              Assign kinds to zones
            </h2>
            <p className="mt-1 text-sm text-gray-500">
              Pick a zone kind per zone. Seed jobs will use the assigned kind by
              default whenever no explicit override is provided.
            </p>
          </div>
          <div className="w-full max-w-sm">
            <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
              Search zones
            </label>
            <input
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
              value={zoneSearch}
              onChange={(event) => setZoneSearch(event.target.value)}
              placeholder="Forest, downtown, tag, metro..."
            />
          </div>
        </div>

        <div className="mt-4 overflow-hidden rounded-2xl border border-gray-200">
          <div className="grid grid-cols-[minmax(0,1.4fr)_minmax(0,0.9fr)_minmax(180px,0.9fr)] gap-4 bg-slate-50 px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <div>Zone</div>
            <div>Current kind</div>
            <div>Assignment</div>
          </div>

          {filteredZones.length === 0 ? (
            <div className="px-4 py-10 text-center text-sm text-gray-500">
              No zones match this search.
            </div>
          ) : (
            filteredZones.map((zone) => {
              const matchedKind = zone.kind
                ? zoneKindBySlug.get(zone.kind)
                : null;
              const currentKindLabel = zone.kind
                ? matchedKind?.name ?? `Unknown (${zone.kind})`
                : 'Unassigned';

              return (
                <div
                  key={zone.id}
                  className="grid grid-cols-[minmax(0,1.4fr)_minmax(0,0.9fr)_minmax(180px,0.9fr)] gap-4 border-t border-gray-200 px-4 py-4 text-sm text-gray-700"
                >
                  <div>
                    <div className="font-medium text-gray-900">
                      <Link
                        to={`/zones/${zone.id}`}
                        className="transition hover:text-slate-700 hover:underline"
                      >
                        {zone.name}
                      </Link>
                    </div>
                    <div className="mt-1 text-xs text-gray-500">
                      {zone.importMetroName || 'Custom zone'}
                    </div>
                  </div>
                  <div className="flex items-center">
                    <span
                      className={`rounded-full px-2.5 py-1 text-xs font-medium ${
                        zone.kind && !matchedKind
                          ? 'bg-amber-100 text-amber-800'
                          : zone.kind
                            ? 'bg-slate-100 text-slate-700'
                            : 'bg-gray-100 text-gray-500'
                      }`}
                    >
                      {currentKindLabel}
                    </span>
                  </div>
                  <div>
                    <select
                      className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                      value={zone.kind || ''}
                      onChange={(event) =>
                        void assignZoneKind(zone.id, event.target.value)
                      }
                      disabled={assigningZoneId === zone.id}
                    >
                      <option value="">Unassigned</option>
                      {zoneKinds.map((zoneKind) => (
                        <option
                          key={`${zone.id}-${zoneKind.id}`}
                          value={zoneKind.slug}
                        >
                          {zoneKind.name}
                        </option>
                      ))}
                    </select>
                    {assigningZoneId === zone.id && (
                      <div className="mt-1 text-xs text-slate-500">
                        Saving...
                      </div>
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </section>
    </div>
  );
};

export default ZoneKinds;
