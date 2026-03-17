import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useInventory } from '@poltergeist/contexts';

type LocationArchetypeRecord = {
  id: string;
  name: string;
};

type ChallengeTemplateRecord = {
  id: string;
  locationArchetypeId: string;
  locationArchetype?: LocationArchetypeRecord | null;
  question: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  scaleWithUserLevel: boolean;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience?: number;
  reward: number;
  inventoryItemId?: number | null;
  itemChoiceRewards: unknown[];
  submissionType: 'photo' | 'text' | 'video';
  difficulty: number;
  statTags: string[];
  proficiency?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type ChallengeTemplateGenerationJob = {
  id: string;
  locationArchetypeId: string;
  status: string;
  count: number;
  createdCount: number;
  errorMessage?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type ChallengeTemplateFormState = {
  locationArchetypeId: string;
  question: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  scaleWithUserLevel: boolean;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  rewardExperience: string;
  reward: string;
  inventoryItemId: string;
  itemChoiceRewardsJson: string;
  submissionType: 'photo' | 'text' | 'video';
  difficulty: string;
  statTags: string;
  proficiency: string;
};

type ChallengeTemplateGenerationFormState = {
  locationArchetypeId: string;
  count: string;
};

const emptyFormState = (): ChallengeTemplateFormState => ({
  locationArchetypeId: '',
  question: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  scaleWithUserLevel: false,
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: '0',
  reward: '0',
  inventoryItemId: '',
  itemChoiceRewardsJson: '[]',
  submissionType: 'photo',
  difficulty: '0',
  statTags: '',
  proficiency: '',
});

const emptyGenerationForm = (): ChallengeTemplateGenerationFormState => ({
  locationArchetypeId: '',
  count: '6',
});

const parseInteger = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const formatDate = (value?: string | null): string => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const formFromRecord = (record: ChallengeTemplateRecord): ChallengeTemplateFormState => ({
  locationArchetypeId: record.locationArchetypeId ?? '',
  question: record.question ?? '',
  description: record.description ?? '',
  imageUrl: record.imageUrl ?? '',
  thumbnailUrl: record.thumbnailUrl ?? '',
  scaleWithUserLevel: Boolean(record.scaleWithUserLevel),
  rewardMode: record.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    record.randomRewardSize === 'medium' || record.randomRewardSize === 'large'
      ? record.randomRewardSize
      : 'small',
  rewardExperience: String(record.rewardExperience ?? 0),
  reward: String(record.reward ?? 0),
  inventoryItemId:
    record.inventoryItemId !== undefined && record.inventoryItemId !== null
      ? String(record.inventoryItemId)
      : '',
  itemChoiceRewardsJson: JSON.stringify(record.itemChoiceRewards ?? [], null, 2),
  submissionType:
    record.submissionType === 'text' || record.submissionType === 'video'
      ? record.submissionType
      : 'photo',
  difficulty: String(record.difficulty ?? 0),
  statTags: (record.statTags ?? []).join(', '),
  proficiency: record.proficiency ?? '',
});

const parseJsonField = <T,>(label: string, value: string): T => {
  try {
    return JSON.parse(value) as T;
  } catch {
    throw new Error(`${label} must be valid JSON.`);
  }
};

const buildPayloadFromForm = (form: ChallengeTemplateFormState) => ({
  locationArchetypeId: form.locationArchetypeId,
  question: form.question,
  description: form.description,
  imageUrl: form.imageUrl,
  thumbnailUrl: form.thumbnailUrl,
  scaleWithUserLevel: form.scaleWithUserLevel,
  rewardMode: form.rewardMode,
  randomRewardSize: form.randomRewardSize,
  rewardExperience: parseInteger(form.rewardExperience, 0),
  reward: parseInteger(form.reward, 0),
  inventoryItemId: form.inventoryItemId.trim()
    ? parseInteger(form.inventoryItemId, 0)
    : null,
  itemChoiceRewards: parseJsonField<unknown[]>(
    'Item choice rewards',
    form.itemChoiceRewardsJson
  ),
  submissionType: form.submissionType,
  difficulty: parseInteger(form.difficulty, 0),
  statTags: form.statTags
    .split(',')
    .map((entry) => entry.trim().toLowerCase())
    .filter(Boolean),
  proficiency: form.proficiency,
});

const jobStatusClassName = (status: string): string => {
  switch (status) {
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    case 'in_progress':
      return 'bg-amber-600';
    default:
      return 'bg-slate-600';
  }
};

export const ChallengeTemplates = () => {
  const { apiClient } = useAPI();
  const { inventoryItems } = useInventory();
  const [locationArchetypes, setLocationArchetypes] = useState<LocationArchetypeRecord[]>([]);
  const [records, setRecords] = useState<ChallengeTemplateRecord[]>([]);
  const [jobs, setJobs] = useState<ChallengeTemplateGenerationJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<ChallengeTemplateRecord | null>(null);
  const [form, setForm] = useState<ChallengeTemplateFormState>(emptyFormState());
  const [generationForm, setGenerationForm] =
    useState<ChallengeTemplateGenerationFormState>(emptyGenerationForm());

  const load = useCallback(async (suppressLoading = false) => {
    try {
      if (!suppressLoading) setLoading(true);
      const [templateResp, jobResp, archetypeResp] = await Promise.all([
        apiClient.get<ChallengeTemplateRecord[]>('/sonar/challenge-templates'),
        apiClient.get<ChallengeTemplateGenerationJob[]>(
          '/sonar/admin/challenge-template-generation-jobs?limit=20'
        ),
        apiClient.get<LocationArchetypeRecord[]>('/sonar/locationArchetypes'),
      ]);
      setRecords(Array.isArray(templateResp) ? templateResp : []);
      setJobs(Array.isArray(jobResp) ? jobResp : []);
      setLocationArchetypes(Array.isArray(archetypeResp) ? archetypeResp : []);
    } catch (error) {
      console.error('Failed to load challenge templates', error);
      alert('Failed to load challenge templates.');
    } finally {
      if (!suppressLoading) setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    if (!jobs.some((job) => ['queued', 'in_progress'].includes(job.status))) return;
    const interval = window.setInterval(() => {
      void load(true);
    }, 5000);
    return () => clearInterval(interval);
  }, [jobs, load]);

  const inventoryHint = useMemo(
    () =>
      inventoryItems
        .slice(0, 12)
        .map((item) => `${item.id}: ${item.name}`)
        .join(', '),
    [inventoryItems]
  );
  const archetypeNameById = useMemo(() => {
    const map = new Map<string, string>();
    locationArchetypes.forEach((archetype) => map.set(archetype.id, archetype.name));
    return map;
  }, [locationArchetypes]);

  const openCreate = () => {
    setEditing(null);
    setForm({
      ...emptyFormState(),
      locationArchetypeId: locationArchetypes[0]?.id ?? '',
    });
    setShowModal(true);
  };

  const openEdit = (record: ChallengeTemplateRecord) => {
    setEditing(record);
    setForm(formFromRecord(record));
    setShowModal(true);
  };

  const closeModal = () => {
    setEditing(null);
    setForm(emptyFormState());
    setShowModal(false);
  };

  const save = async () => {
    try {
      setSaving(true);
      const payload = buildPayloadFromForm(form);
      const response = editing
        ? await apiClient.put<ChallengeTemplateRecord>(
            `/sonar/challenge-templates/${editing.id}`,
            payload
          )
        : await apiClient.post<ChallengeTemplateRecord>(
            '/sonar/challenge-templates',
            payload
          );
      if (editing) {
        setRecords((prev) => prev.map((record) => (record.id === response.id ? response : record)));
      } else {
        setRecords((prev) => [response, ...prev]);
      }
      closeModal();
    } catch (error) {
      console.error('Failed to save challenge template', error);
      alert(error instanceof Error ? error.message : 'Failed to save challenge template.');
    } finally {
      setSaving(false);
    }
  };

  const deleteRecord = async (record: ChallengeTemplateRecord) => {
    if (!window.confirm('Delete this challenge template?')) return;
    try {
      await apiClient.delete(`/sonar/challenge-templates/${record.id}`);
      setRecords((prev) => prev.filter((entry) => entry.id !== record.id));
    } catch (error) {
      console.error('Failed to delete challenge template', error);
      alert('Failed to delete challenge template.');
    }
  };

  const queueGeneration = async () => {
    const count = parseInteger(generationForm.count, 0);
    if (!generationForm.locationArchetypeId) {
      alert('Location archetype is required.');
      return;
    }
    if (count <= 0 || count > 100) {
      alert('Count must be between 1 and 100.');
      return;
    }
    try {
      setGenerating(true);
      await apiClient.post('/sonar/admin/challenge-template-generation-jobs', {
        locationArchetypeId: generationForm.locationArchetypeId,
        count,
      });
      await load(true);
    } catch (error) {
      console.error('Failed to queue challenge template generation', error);
      alert('Failed to queue challenge template generation.');
    } finally {
      setGenerating(false);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Challenge Templates</h1>
          <p className="text-sm text-gray-600">
            Reusable challenges keyed to a location archetype instead of a concrete POI.
          </p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          className="rounded bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700"
        >
          New Template
        </button>
      </div>

      <section className="rounded-lg border bg-white p-4 shadow-sm space-y-4">
        <h2 className="text-lg font-semibold">Generate Templates</h2>
        <div className="grid gap-4 md:grid-cols-3">
          <label className="block text-sm">
            Location Archetype
            <select
              value={generationForm.locationArchetypeId}
              onChange={(event) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  locationArchetypeId: event.target.value,
                }))
              }
              className="mt-1 w-full rounded border p-2"
            >
              <option value="">Select an archetype</option>
              {locationArchetypes.map((archetype) => (
                <option key={archetype.id} value={archetype.id}>
                  {archetype.name}
                </option>
              ))}
            </select>
          </label>
          <label className="block text-sm">
            Count
            <input
              value={generationForm.count}
              onChange={(event) =>
                setGenerationForm((prev) => ({ ...prev, count: event.target.value }))
              }
              className="mt-1 w-full rounded border p-2"
            />
          </label>
          <div className="flex items-end">
            <button
              type="button"
              onClick={queueGeneration}
              disabled={generating}
              className="rounded bg-emerald-600 px-4 py-2 text-white hover:bg-emerald-700 disabled:opacity-60"
            >
              {generating ? 'Queueing...' : 'Queue Generation'}
            </button>
          </div>
        </div>
        <div className="text-xs text-gray-500">
          Inventory item examples for reward JSON: {inventoryHint || 'none loaded'}
        </div>
        <div className="space-y-3">
          {jobs.length === 0 ? (
            <p className="text-sm text-gray-500">No generation jobs yet.</p>
          ) : (
            jobs.map((job) => (
              <div key={job.id} className="rounded border p-3">
                <div className="flex items-center justify-between gap-3">
                  <div className="text-sm">
                    <div className="font-medium">
                      {archetypeNameById.get(job.locationArchetypeId) ?? 'Unknown archetype'} x {job.count}
                    </div>
                    <div className="text-gray-500">
                      Created {job.createdCount} • queued {formatDate(job.createdAt)}
                    </div>
                    {job.errorMessage ? (
                      <div className="text-red-600">{job.errorMessage}</div>
                    ) : null}
                  </div>
                  <span
                    className={`rounded px-2 py-1 text-xs font-semibold text-white ${jobStatusClassName(
                      job.status
                    )}`}
                  >
                    {job.status}
                  </span>
                </div>
              </div>
            ))
          )}
        </div>
      </section>

      <section className="rounded-lg border bg-white p-4 shadow-sm">
        {loading ? (
          <p className="text-sm text-gray-500">Loading templates...</p>
        ) : records.length === 0 ? (
          <p className="text-sm text-gray-500">No challenge templates yet.</p>
        ) : (
          <div className="space-y-4">
            {records.map((record) => (
              <div key={record.id} className="rounded border p-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-2">
                    <div className="text-sm text-gray-500">
                      {record.locationArchetype?.name ??
                        archetypeNameById.get(record.locationArchetypeId) ??
                        'Unknown archetype'}
                    </div>
                    <div className="font-medium">{record.question}</div>
                    <div className="text-sm whitespace-pre-wrap text-gray-700">
                      {record.description}
                    </div>
                    <div className="text-sm text-gray-500">
                      Difficulty {record.difficulty} • Updated {formatDate(record.updatedAt)}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => openEdit(record)}
                      className="rounded border px-3 py-2 text-sm hover:bg-gray-50"
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      onClick={() => deleteRecord(record)}
                      className="rounded border border-red-300 px-3 py-2 text-sm text-red-700 hover:bg-red-50"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {showModal ? (
        <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/40 p-6 overflow-auto">
          <div className="w-full max-w-4xl rounded-lg bg-white p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold">
                {editing ? 'Edit Challenge Template' : 'Create Challenge Template'}
              </h2>
              <button type="button" onClick={closeModal} className="text-sm text-gray-500">
                Close
              </button>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="block text-sm">
                Location Archetype
                <select
                  value={form.locationArchetypeId}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, locationArchetypeId: event.target.value }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="">Select an archetype</option>
                  {locationArchetypes.map((archetype) => (
                    <option key={archetype.id} value={archetype.id}>
                      {archetype.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="block text-sm">
                Submission Type
                <select
                  value={form.submissionType}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      submissionType: event.target.value as 'photo' | 'text' | 'video',
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="photo">Photo</option>
                  <option value="text">Text</option>
                  <option value="video">Video</option>
                </select>
              </label>
              <label className="block text-sm md:col-span-2">
                Question
                <input
                  value={form.question}
                  onChange={(event) => setForm((prev) => ({ ...prev, question: event.target.value }))}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm md:col-span-2">
                Description
                <textarea
                  value={form.description}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, description: event.target.value }))
                  }
                  rows={5}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Image URL
                <input
                  value={form.imageUrl}
                  onChange={(event) => setForm((prev) => ({ ...prev, imageUrl: event.target.value }))}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Thumbnail URL
                <input
                  value={form.thumbnailUrl}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, thumbnailUrl: event.target.value }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Difficulty
                <input
                  value={form.difficulty}
                  onChange={(event) => setForm((prev) => ({ ...prev, difficulty: event.target.value }))}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Reward
                <input
                  value={form.reward}
                  onChange={(event) => setForm((prev) => ({ ...prev, reward: event.target.value }))}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Reward Experience
                <input
                  value={form.rewardExperience}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, rewardExperience: event.target.value }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Reward Mode
                <select
                  value={form.rewardMode}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      rewardMode: event.target.value as 'explicit' | 'random',
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="random">Random</option>
                  <option value="explicit">Explicit</option>
                </select>
              </label>
              <label className="block text-sm">
                Random Reward Size
                <select
                  value={form.randomRewardSize}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      randomRewardSize: event.target.value as 'small' | 'medium' | 'large',
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </label>
              <label className="block text-sm">
                Inventory Item ID
                <input
                  value={form.inventoryItemId}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, inventoryItemId: event.target.value }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Stat Tags CSV
                <input
                  value={form.statTags}
                  onChange={(event) => setForm((prev) => ({ ...prev, statTags: event.target.value }))}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Proficiency
                <input
                  value={form.proficiency}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, proficiency: event.target.value }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="flex items-center gap-2 text-sm mt-6">
                <input
                  type="checkbox"
                  checked={form.scaleWithUserLevel}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, scaleWithUserLevel: event.target.checked }))
                  }
                />
                Scale with user level
              </label>
              <label className="block text-sm md:col-span-2">
                Item Choice Rewards JSON
                <textarea
                  value={form.itemChoiceRewardsJson}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, itemChoiceRewardsJson: event.target.value }))
                  }
                  rows={8}
                  className="mt-1 w-full rounded border p-2 font-mono text-xs"
                />
              </label>
            </div>

            <div className="rounded bg-slate-50 p-3 text-xs text-gray-600">
              Inventory item examples: {inventoryHint || 'none loaded'}
            </div>

            <div className="flex justify-end gap-3">
              <button type="button" onClick={closeModal} className="rounded border px-4 py-2">
                Cancel
              </button>
              <button
                type="button"
                onClick={save}
                disabled={saving}
                className="rounded bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700 disabled:opacity-60"
              >
                {saving ? 'Saving...' : editing ? 'Save Changes' : 'Create Template'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

export default ChallengeTemplates;
