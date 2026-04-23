import React, {
  useCallback,
  useDeferredValue,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { useAPI, useInventory } from '@poltergeist/contexts';
import { Spell, ZoneGenre } from '@poltergeist/types';
import ContentDashboard from './ContentDashboard.tsx';
import {
  countBy,
  difficultyBandLabel,
  useAdminAggregateDataset,
} from './contentDashboardUtils.ts';
import { useZoneKinds, zoneKindLabel } from './zoneKindHelpers.ts';

type ScenarioFailureDrainType = 'none' | 'flat' | 'percent';
type ScenarioFailurePenaltyMode = 'shared' | 'individual';
type ScenarioSuccessRewardMode = 'shared' | 'individual';

type ScenarioTemplateRecord = {
  id: string;
  genreId: string;
  genre?: ZoneGenre;
  zoneKind?: string;
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  scaleWithUserLevel: boolean;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  difficulty: number;
  rewardExperience: number;
  rewardGold: number;
  openEnded: boolean;
  successHandoffText?: string;
  failureHandoffText?: string;
  failurePenaltyMode: ScenarioFailurePenaltyMode;
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: number;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: number;
  failureStatuses: unknown[];
  successRewardMode: ScenarioSuccessRewardMode;
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: number;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: number;
  successStatuses: unknown[];
  options: unknown[];
  itemRewards: unknown[];
  itemChoiceRewards: unknown[];
  spellRewards: unknown[];
  createdAt?: string;
  updatedAt?: string;
};

type ScenarioTemplateGenerationJob = {
  id: string;
  genreId: string;
  genre?: ZoneGenre;
  zoneKind?: string;
  status: string;
  count: number;
  openEnded: boolean;
  createdCount: number;
  errorMessage?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type ScenarioTemplateFormState = {
  genreId: string;
  zoneKind: string;
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  scaleWithUserLevel: boolean;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  difficulty: string;
  rewardExperience: string;
  rewardGold: string;
  openEnded: boolean;
  successHandoffText: string;
  failureHandoffText: string;
  failurePenaltyMode: ScenarioFailurePenaltyMode;
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: string;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: string;
  failureStatusesJson: string;
  successRewardMode: ScenarioSuccessRewardMode;
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: string;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: string;
  successStatusesJson: string;
  optionsJson: string;
  itemRewardsJson: string;
  itemChoiceRewardsJson: string;
  spellRewardsJson: string;
};

type ScenarioTemplateGenerationFormState = {
  count: string;
  genreId: string;
  zoneKind: string;
  openEnded: boolean;
};

type PaginatedResponse<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
};

const scenarioTemplateListPageSize = 25;

const PaginationControls = ({
  page,
  pageSize,
  total,
  onPageChange,
}: {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
}) => {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = total === 0 ? 0 : Math.min(total, page * pageSize);

  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-3">
      <p className="text-sm text-gray-600">
        {total === 0
          ? 'No scenario templates.'
          : `Showing ${start}-${end} of ${total} scenario templates`}
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1}
        >
          Previous
        </button>
        <span className="text-sm text-gray-600">
          Page {page} of {totalPages}
        </span>
        <button
          type="button"
          className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
          onClick={() => onPageChange(page + 1)}
          disabled={page >= totalPages}
        >
          Next
        </button>
      </div>
    </div>
  );
};

const emptyFormState = (): ScenarioTemplateFormState => ({
  genreId: '',
  zoneKind: '',
  prompt: '',
  imageUrl: '',
  thumbnailUrl: '',
  scaleWithUserLevel: false,
  rewardMode: 'random',
  randomRewardSize: 'small',
  difficulty: '24',
  rewardExperience: '0',
  rewardGold: '0',
  openEnded: false,
  successHandoffText: '',
  failureHandoffText: '',
  failurePenaltyMode: 'shared',
  failureHealthDrainType: 'none',
  failureHealthDrainValue: '0',
  failureManaDrainType: 'none',
  failureManaDrainValue: '0',
  failureStatusesJson: '[]',
  successRewardMode: 'shared',
  successHealthRestoreType: 'none',
  successHealthRestoreValue: '0',
  successManaRestoreType: 'none',
  successManaRestoreValue: '0',
  successStatusesJson: '[]',
  optionsJson: '[]',
  itemRewardsJson: '[]',
  itemChoiceRewardsJson: '[]',
  spellRewardsJson: '[]',
});

const emptyGenerationForm = (): ScenarioTemplateGenerationFormState => ({
  count: '6',
  genreId: '',
  zoneKind: '',
  openEnded: false,
});

const defaultGenreIdFromList = (genres: ZoneGenre[]): string => {
  const fantasy = genres.find(
    (genre) => (genre.name || '').trim().toLowerCase() === 'fantasy'
  );
  return fantasy?.id ?? genres[0]?.id ?? '';
};

const formatGenreLabel = (genre?: ZoneGenre | null): string =>
  genre?.name?.trim() || 'Fantasy';

const prettyJson = (value: unknown): string =>
  JSON.stringify(value ?? [], null, 2);

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

const formFromRecord = (
  record: ScenarioTemplateRecord
): ScenarioTemplateFormState => ({
  genreId: record.genreId ?? record.genre?.id ?? '',
  zoneKind: record.zoneKind?.trim() ?? '',
  prompt: record.prompt ?? '',
  imageUrl: record.imageUrl ?? '',
  thumbnailUrl: record.thumbnailUrl ?? '',
  scaleWithUserLevel: Boolean(record.scaleWithUserLevel),
  rewardMode: record.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    record.randomRewardSize === 'medium' || record.randomRewardSize === 'large'
      ? record.randomRewardSize
      : 'small',
  difficulty: String(record.difficulty ?? 0),
  rewardExperience: String(record.rewardExperience ?? 0),
  rewardGold: String(record.rewardGold ?? 0),
  openEnded: Boolean(record.openEnded),
  successHandoffText: record.successHandoffText?.trim() ?? '',
  failureHandoffText: record.failureHandoffText?.trim() ?? '',
  failurePenaltyMode:
    record.failurePenaltyMode === 'individual' ? 'individual' : 'shared',
  failureHealthDrainType:
    record.failureHealthDrainType === 'flat' ||
    record.failureHealthDrainType === 'percent'
      ? record.failureHealthDrainType
      : 'none',
  failureHealthDrainValue: String(record.failureHealthDrainValue ?? 0),
  failureManaDrainType:
    record.failureManaDrainType === 'flat' ||
    record.failureManaDrainType === 'percent'
      ? record.failureManaDrainType
      : 'none',
  failureManaDrainValue: String(record.failureManaDrainValue ?? 0),
  failureStatusesJson: prettyJson(record.failureStatuses),
  successRewardMode:
    record.successRewardMode === 'individual' ? 'individual' : 'shared',
  successHealthRestoreType:
    record.successHealthRestoreType === 'flat' ||
    record.successHealthRestoreType === 'percent'
      ? record.successHealthRestoreType
      : 'none',
  successHealthRestoreValue: String(record.successHealthRestoreValue ?? 0),
  successManaRestoreType:
    record.successManaRestoreType === 'flat' ||
    record.successManaRestoreType === 'percent'
      ? record.successManaRestoreType
      : 'none',
  successManaRestoreValue: String(record.successManaRestoreValue ?? 0),
  successStatusesJson: prettyJson(record.successStatuses),
  optionsJson: prettyJson(record.options),
  itemRewardsJson: prettyJson(record.itemRewards),
  itemChoiceRewardsJson: prettyJson(record.itemChoiceRewards),
  spellRewardsJson: prettyJson(record.spellRewards),
});

const parseJsonField = <T,>(label: string, value: string): T => {
  try {
    return JSON.parse(value) as T;
  } catch (error) {
    throw new Error(`${label} must be valid JSON.`);
  }
};

const buildPayloadFromForm = (form: ScenarioTemplateFormState) => ({
  genreId: form.genreId.trim(),
  zoneKind: form.zoneKind.trim(),
  prompt: form.prompt,
  imageUrl: form.imageUrl,
  thumbnailUrl: form.thumbnailUrl,
  scaleWithUserLevel: form.scaleWithUserLevel,
  rewardMode: form.rewardMode,
  randomRewardSize: form.randomRewardSize,
  difficulty: parseInteger(form.difficulty, 0),
  rewardExperience: parseInteger(form.rewardExperience, 0),
  rewardGold: parseInteger(form.rewardGold, 0),
  openEnded: form.openEnded,
  successHandoffText: form.successHandoffText.trim(),
  failureHandoffText: form.failureHandoffText.trim(),
  failurePenaltyMode: form.failurePenaltyMode,
  failureHealthDrainType: form.failureHealthDrainType,
  failureHealthDrainValue: parseInteger(form.failureHealthDrainValue, 0),
  failureManaDrainType: form.failureManaDrainType,
  failureManaDrainValue: parseInteger(form.failureManaDrainValue, 0),
  failureStatuses: parseJsonField<unknown[]>(
    'Failure statuses',
    form.failureStatusesJson
  ),
  successRewardMode: form.successRewardMode,
  successHealthRestoreType: form.successHealthRestoreType,
  successHealthRestoreValue: parseInteger(form.successHealthRestoreValue, 0),
  successManaRestoreType: form.successManaRestoreType,
  successManaRestoreValue: parseInteger(form.successManaRestoreValue, 0),
  successStatuses: parseJsonField<unknown[]>(
    'Success statuses',
    form.successStatusesJson
  ),
  options: parseJsonField<unknown[]>('Options', form.optionsJson),
  itemRewards: parseJsonField<unknown[]>('Item rewards', form.itemRewardsJson),
  itemChoiceRewards: parseJsonField<unknown[]>(
    'Item choice rewards',
    form.itemChoiceRewardsJson
  ),
  spellRewards: parseJsonField<unknown[]>(
    'Spell rewards',
    form.spellRewardsJson
  ),
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

export const ScenarioTemplates = () => {
  const { apiClient } = useAPI();
  const { inventoryItems } = useInventory();
  const { zoneKinds, zoneKindBySlug } = useZoneKinds();
  const [spells, setSpells] = useState<Spell[]>([]);
  const [genres, setGenres] = useState<ZoneGenre[]>([]);
  const [records, setRecords] = useState<ScenarioTemplateRecord[]>([]);
  const [jobs, setJobs] = useState<ScenarioTemplateGenerationJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [genreFilter, setGenreFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [saving, setSaving] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<ScenarioTemplateRecord | null>(null);
  const [form, setForm] = useState<ScenarioTemplateFormState>(emptyFormState());
  const [generationForm, setGenerationForm] =
    useState<ScenarioTemplateGenerationFormState>(emptyGenerationForm());
  const deferredQuery = useDeferredValue(query);
  const defaultGenreId = useMemo(
    () => defaultGenreIdFromList(genres),
    [genres]
  );

  const load = useCallback(
    async (suppressLoading = false) => {
      try {
        if (!suppressLoading) setLoading(true);
        const [templateResp, jobResp, spellResp, genreResp] = await Promise.all([
          apiClient.get<PaginatedResponse<ScenarioTemplateRecord>>(
            '/sonar/admin/scenario-templates',
            {
              page,
              pageSize: scenarioTemplateListPageSize,
              query: deferredQuery.trim(),
              genreId: genreFilter === 'all' ? '' : genreFilter,
            }
          ),
          apiClient.get<ScenarioTemplateGenerationJob[]>(
            '/sonar/admin/scenario-template-generation-jobs?limit=20'
          ),
          apiClient.get<Spell[]>('/sonar/spells'),
          apiClient.get<ZoneGenre[]>('/sonar/zone-genres?includeInactive=true'),
        ]);
        setRecords(
          Array.isArray(templateResp?.items) ? templateResp.items : []
        );
        setTotal(templateResp?.total ?? 0);
        setJobs(Array.isArray(jobResp) ? jobResp : []);
        setSpells(Array.isArray(spellResp) ? spellResp : []);
        setGenres(Array.isArray(genreResp) ? genreResp : []);
      } catch (error) {
        console.error('Failed to load scenario templates', error);
        alert('Failed to load scenario templates.');
      } finally {
        if (!suppressLoading) setLoading(false);
      }
    },
    [apiClient, deferredQuery, genreFilter, page]
  );

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    setPage(1);
  }, [genreFilter, query]);

  useEffect(() => {
    const totalPages = Math.max(
      1,
      Math.ceil(total / scenarioTemplateListPageSize)
    );
    if (page > totalPages) {
      setPage(totalPages);
    }
  }, [page, total]);

  useEffect(() => {
    if (!jobs.some((job) => ['queued', 'in_progress'].includes(job.status)))
      return;
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
  const spellHint = useMemo(
    () =>
      spells
        .slice(0, 12)
        .map((spell) => `${spell.id}: ${spell.name}`)
        .join(', '),
    [spells]
  );
  const dashboardParams = useMemo(
    () => ({
      query: deferredQuery.trim(),
      genreId: genreFilter === 'all' ? '' : genreFilter,
    }),
    [deferredQuery, genreFilter]
  );
  const {
    items: dashboardRecords,
    loading: dashboardLoading,
    error: dashboardError,
  } = useAdminAggregateDataset<ScenarioTemplateRecord>(
    '/sonar/admin/scenario-templates',
    dashboardParams
  );
  const dashboardMetrics = useMemo(() => {
    const totalTemplates = dashboardRecords.length;
    const openEndedCount = dashboardRecords.filter(
      (record) => record.openEnded
    ).length;
    const explicitRewardCount = dashboardRecords.filter(
      (record) => record.rewardMode === 'explicit'
    ).length;
    const scaledCount = dashboardRecords.filter(
      (record) => record.scaleWithUserLevel
    ).length;

    return [
      { label: 'Templates', value: totalTemplates },
      {
        label: 'Open-ended',
        value: openEndedCount,
        note: `${Math.max(0, totalTemplates - openEndedCount)} choice-based`,
      },
      { label: 'Explicit Rewards', value: explicitRewardCount },
      { label: 'Scaled Difficulty', value: scaledCount },
    ];
  }, [dashboardRecords]);
  const dashboardSections = useMemo(
    () => [
      {
        title: 'Genre Mix',
        note: 'Reusable scenario templates grouped by story genre.',
        buckets: countBy(
          dashboardRecords,
          (record) =>
            formatGenreLabel(
              record.genre ??
                genres.find((genre) => genre.id === record.genreId) ??
                null
            ),
          { emptyLabel: 'Fantasy' }
        ),
      },
      {
        title: 'Zone Kind Coverage',
        note: 'Which environments the current template pool is tagged to support.',
        buckets: countBy(
          dashboardRecords,
          (record) =>
            record.zoneKind?.trim()
              ? zoneKindLabel(record.zoneKind, zoneKindBySlug)
              : 'Unassigned',
          { emptyLabel: 'Unassigned' }
        ),
      },
      {
        title: 'Difficulty Bands',
        note: 'How hard the current template pool skews.',
        buckets: countBy(
          dashboardRecords,
          (record) => difficultyBandLabel(record.difficulty),
          { limit: 4 }
        ),
      },
      {
        title: 'Reward Model',
        note: 'Whether templates carry explicit or randomized rewards.',
        buckets: countBy(dashboardRecords, (record) =>
          record.rewardMode === 'explicit' ? 'Explicit rewards' : 'Randomized'
        ),
      },
    ],
    [dashboardRecords, genres, zoneKindBySlug]
  );

  const openCreate = () => {
    setEditing(null);
    setForm(emptyFormState());
    setShowModal(true);
  };

  const openEdit = (record: ScenarioTemplateRecord) => {
    setEditing(record);
    setForm(formFromRecord(record));
    setShowModal(true);
  };

  const closeModal = () => {
    setEditing(null);
    setForm(emptyFormState());
    setShowModal(false);
  };

  useEffect(() => {
    if (!generationForm.genreId && defaultGenreId) {
      setGenerationForm((prev) =>
        prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
      );
    }
  }, [defaultGenreId, generationForm.genreId]);

  useEffect(() => {
    if (!showModal) return;
    if (!form.genreId && defaultGenreId) {
      setForm((prev) =>
        prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
      );
    }
  }, [defaultGenreId, form.genreId, showModal]);

  const save = async () => {
    try {
      setSaving(true);
      const payload = buildPayloadFromForm(form);
      if (editing) {
        await apiClient.put<ScenarioTemplateRecord>(
          `/sonar/scenario-templates/${editing.id}`,
          payload
        );
      } else {
        await apiClient.post<ScenarioTemplateRecord>(
          '/sonar/scenario-templates',
          payload
        );
      }
      await load(true);
      closeModal();
    } catch (error) {
      console.error('Failed to save scenario template', error);
      alert(
        error instanceof Error
          ? error.message
          : 'Failed to save scenario template.'
      );
    } finally {
      setSaving(false);
    }
  };

  const deleteRecord = async (record: ScenarioTemplateRecord) => {
    if (!window.confirm('Delete this scenario template?')) return;
    try {
      await apiClient.delete(`/sonar/scenario-templates/${record.id}`);
      await load(true);
    } catch (error) {
      console.error('Failed to delete scenario template', error);
      alert('Failed to delete scenario template.');
    }
  };

  const queueGeneration = async () => {
    const count = parseInteger(generationForm.count, 0);
    if (count <= 0 || count > 100) {
      alert('Count must be between 1 and 100.');
      return;
    }
    try {
      setGenerating(true);
      await apiClient.post('/sonar/admin/scenario-template-generation-jobs', {
        count,
        genreId: generationForm.genreId.trim(),
        zoneKind: generationForm.zoneKind,
        openEnded: generationForm.openEnded,
      });
      await load(true);
    } catch (error) {
      console.error('Failed to queue scenario template generation', error);
      alert('Failed to queue scenario template generation.');
    } finally {
      setGenerating(false);
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Scenario Templates</h1>
          <p className="text-sm text-gray-600">
            Reusable scenarios without map location binding.
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

      <ContentDashboard
        title="Scenario Template Dashboard"
        subtitle="Aggregate template coverage for the current search and genre filters."
        status={
          query.trim() || genreFilter !== 'all'
            ? 'Reflects current filters'
            : 'All reusable templates'
        }
        loading={dashboardLoading}
        error={dashboardError}
        metrics={dashboardMetrics}
        sections={dashboardSections}
      />

      <section className="rounded-lg border bg-white p-4 shadow-sm space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold">Generate Templates</h2>
            <p className="text-sm text-gray-600">
              Create reusable scenario templates asynchronously.
            </p>
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-4">
          <label className="block text-sm">
            Count
            <input
              value={generationForm.count}
              onChange={(event) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  count: event.target.value,
                }))
              }
              className="mt-1 w-full rounded border p-2"
            />
          </label>
          <label className="block text-sm">
            Genre
            <select
              value={generationForm.genreId}
              onChange={(event) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  genreId: event.target.value,
                }))
              }
              className="mt-1 w-full rounded border p-2"
            >
              {genres.length === 0 ? (
                <option value="">Fantasy</option>
              ) : (
                genres.map((genre) => (
                  <option key={`template-generation-${genre.id}`} value={genre.id}>
                    {formatGenreLabel(genre)}
                    {genre.active === false ? ' (inactive)' : ''}
                  </option>
                ))
              )}
            </select>
          </label>
          <label className="block text-sm">
            Zone Kind
            <select
              value={generationForm.zoneKind}
              onChange={(event) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  zoneKind: event.target.value,
                }))
              }
              className="mt-1 w-full rounded border p-2"
            >
              <option value="">Any zone kind</option>
              {zoneKinds.map((zoneKind) => (
                <option
                  key={`template-generation-kind-${zoneKind.id}`}
                  value={zoneKind.slug}
                >
                  {zoneKind.name}
                </option>
              ))}
            </select>
          </label>
          <label className="flex items-center gap-2 text-sm mt-6">
            <input
              type="checkbox"
              checked={generationForm.openEnded}
              onChange={(event) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  openEnded: event.target.checked,
                }))
              }
            />
            Open ended
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
          JSON reward IDs can reference inventory items like:{' '}
          {inventoryHint || 'none loaded'}
          <br />
          Spell reward IDs can reference spells like:{' '}
          {spellHint || 'none loaded'}
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
                      {job.openEnded ? 'Open-ended' : 'Choice-based'} x{' '}
                      {job.count}
                    </div>
                    <div className="text-gray-500">
                      {formatGenreLabel(
                        job.genre ??
                          genres.find((genre) => genre.id === job.genreId)
                      )}
                      {job.zoneKind ? (
                        <> • {zoneKindLabel(job.zoneKind, zoneKindBySlug)}</>
                      ) : null}{' '}
                      • Created {job.createdCount} • queued{' '}
                      {formatDate(job.createdAt)}
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
        <div className="mb-4 flex flex-wrap gap-3">
          <input
            type="text"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search by prompt or ID..."
            className="w-full max-w-xl rounded border p-2"
          />
          <select
            value={genreFilter}
            onChange={(event) => setGenreFilter(event.target.value)}
            className="w-full max-w-xs rounded border p-2"
            aria-label="Filter scenario templates by genre"
          >
            <option value="all">All Genres</option>
            {genres.map((genre) => (
              <option key={`template-filter-${genre.id}`} value={genre.id}>
                {formatGenreLabel(genre)}
              </option>
            ))}
          </select>
        </div>
        {loading ? (
          <p className="text-sm text-gray-500">Loading templates...</p>
        ) : records.length === 0 ? (
          <p className="text-sm text-gray-500">No scenario templates yet.</p>
        ) : (
          <div className="space-y-4">
            {records.map((record) => (
              <div key={record.id} className="rounded border p-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-2">
                    <div className="text-sm text-gray-500">
                      Genre:{' '}
                      {formatGenreLabel(
                        record.genre ??
                          genres.find((genre) => genre.id === record.genreId)
                      )}{' '}
                      • Zone kind:{' '}
                      {record.zoneKind?.trim()
                        ? zoneKindLabel(record.zoneKind, zoneKindBySlug)
                        : 'Unassigned'}{' '}
                      •{' '}
                      {record.openEnded ? 'Open ended' : 'Choice based'} •
                      difficulty {record.difficulty}
                    </div>
                    <div className="font-medium whitespace-pre-wrap">
                      {record.prompt}
                    </div>
                    <div className="text-sm text-gray-500">
                      Reward mode: {record.rewardMode ?? 'random'} • Updated{' '}
                      {formatDate(record.updatedAt)}
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
            <PaginationControls
              page={page}
              pageSize={scenarioTemplateListPageSize}
              total={total}
              onPageChange={setPage}
            />
          </div>
        )}
      </section>

      {showModal ? (
        <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/40 p-6 overflow-auto">
          <div className="w-full max-w-5xl rounded-lg bg-white p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold">
                {editing
                  ? 'Edit Scenario Template'
                  : 'Create Scenario Template'}
              </h2>
              <button
                type="button"
                onClick={closeModal}
                className="text-sm text-gray-500"
              >
                Close
              </button>
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <label className="block text-sm">
                Genre
                <select
                  value={form.genreId}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      genreId: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  {genres.length === 0 ? (
                    <option value="">Fantasy</option>
                  ) : (
                    genres.map((genre) => (
                      <option key={`template-genre-${genre.id}`} value={genre.id}>
                        {formatGenreLabel(genre)}
                        {genre.active === false ? ' (inactive)' : ''}
                      </option>
                    ))
                  )}
                </select>
              </label>
              <label className="block text-sm">
                Zone Kind
                <select
                  value={form.zoneKind}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      zoneKind: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="">Unassigned</option>
                  {zoneKinds.map((zoneKind) => (
                    <option
                      key={`template-zone-kind-${zoneKind.id}`}
                      value={zoneKind.slug}
                    >
                      {zoneKind.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="block text-sm md:col-span-2">
                Prompt
                <textarea
                  value={form.prompt}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, prompt: event.target.value }))
                  }
                  rows={4}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Success Handoff
                <textarea
                  value={form.successHandoffText}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      successHandoffText: event.target.value,
                    }))
                  }
                  rows={3}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Failure Handoff
                <textarea
                  value={form.failureHandoffText}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHandoffText: event.target.value,
                    }))
                  }
                  rows={3}
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Image URL
                <input
                  value={form.imageUrl}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      imageUrl: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Thumbnail URL
                <input
                  value={form.thumbnailUrl}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      thumbnailUrl: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Difficulty
                <input
                  value={form.difficulty}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      difficulty: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Reward Experience
                <input
                  value={form.rewardExperience}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      rewardExperience: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Reward Gold
                <input
                  value={form.rewardGold}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      rewardGold: event.target.value,
                    }))
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
                      randomRewardSize: event.target.value as
                        | 'small'
                        | 'medium'
                        | 'large',
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </label>
              <label className="flex items-center gap-2 text-sm mt-6">
                <input
                  type="checkbox"
                  checked={form.openEnded}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      openEnded: event.target.checked,
                    }))
                  }
                />
                Open ended
              </label>
              <label className="flex items-center gap-2 text-sm mt-6">
                <input
                  type="checkbox"
                  checked={form.scaleWithUserLevel}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      scaleWithUserLevel: event.target.checked,
                    }))
                  }
                />
                Scale with user level
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <label className="block text-sm">
                Failure Penalty Mode
                <select
                  value={form.failurePenaltyMode}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      failurePenaltyMode: event.target
                        .value as ScenarioFailurePenaltyMode,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="shared">Shared</option>
                  <option value="individual">Individual</option>
                </select>
              </label>
              <label className="block text-sm">
                Success Reward Mode
                <select
                  value={form.successRewardMode}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      successRewardMode: event.target
                        .value as ScenarioSuccessRewardMode,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="shared">Shared</option>
                  <option value="individual">Individual</option>
                </select>
              </label>
              <label className="block text-sm">
                Failure Health Drain Type
                <select
                  value={form.failureHealthDrainType}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHealthDrainType: event.target
                        .value as ScenarioFailureDrainType,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="none">None</option>
                  <option value="flat">Flat</option>
                  <option value="percent">Percent</option>
                </select>
              </label>
              <label className="block text-sm">
                Failure Health Drain Value
                <input
                  value={form.failureHealthDrainValue}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHealthDrainValue: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Failure Mana Drain Type
                <select
                  value={form.failureManaDrainType}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      failureManaDrainType: event.target
                        .value as ScenarioFailureDrainType,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="none">None</option>
                  <option value="flat">Flat</option>
                  <option value="percent">Percent</option>
                </select>
              </label>
              <label className="block text-sm">
                Failure Mana Drain Value
                <input
                  value={form.failureManaDrainValue}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      failureManaDrainValue: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Success Health Restore Type
                <select
                  value={form.successHealthRestoreType}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      successHealthRestoreType: event.target
                        .value as ScenarioFailureDrainType,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="none">None</option>
                  <option value="flat">Flat</option>
                  <option value="percent">Percent</option>
                </select>
              </label>
              <label className="block text-sm">
                Success Health Restore Value
                <input
                  value={form.successHealthRestoreValue}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      successHealthRestoreValue: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
              <label className="block text-sm">
                Success Mana Restore Type
                <select
                  value={form.successManaRestoreType}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      successManaRestoreType: event.target
                        .value as ScenarioFailureDrainType,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                >
                  <option value="none">None</option>
                  <option value="flat">Flat</option>
                  <option value="percent">Percent</option>
                </select>
              </label>
              <label className="block text-sm">
                Success Mana Restore Value
                <input
                  value={form.successManaRestoreValue}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      successManaRestoreValue: event.target.value,
                    }))
                  }
                  className="mt-1 w-full rounded border p-2"
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <label className="block text-sm">
                Failure Statuses JSON
                <textarea
                  value={form.failureStatusesJson}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      failureStatusesJson: event.target.value,
                    }))
                  }
                  rows={8}
                  className="mt-1 w-full rounded border p-2 font-mono text-xs"
                />
              </label>
              <label className="block text-sm">
                Success Statuses JSON
                <textarea
                  value={form.successStatusesJson}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      successStatusesJson: event.target.value,
                    }))
                  }
                  rows={8}
                  className="mt-1 w-full rounded border p-2 font-mono text-xs"
                />
              </label>
              <label className="block text-sm md:col-span-2">
                Options JSON
                <div className="mt-1 text-xs text-gray-500">
                  Choice options can include `successHandoffText` and
                  `failureHandoffText` in addition to `successText` and
                  `failureText`.
                </div>
                <textarea
                  value={form.optionsJson}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      optionsJson: event.target.value,
                    }))
                  }
                  rows={10}
                  className="mt-1 w-full rounded border p-2 font-mono text-xs"
                />
              </label>
              <label className="block text-sm">
                Item Rewards JSON
                <textarea
                  value={form.itemRewardsJson}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      itemRewardsJson: event.target.value,
                    }))
                  }
                  rows={8}
                  className="mt-1 w-full rounded border p-2 font-mono text-xs"
                />
              </label>
              <label className="block text-sm">
                Item Choice Rewards JSON
                <textarea
                  value={form.itemChoiceRewardsJson}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      itemChoiceRewardsJson: event.target.value,
                    }))
                  }
                  rows={8}
                  className="mt-1 w-full rounded border p-2 font-mono text-xs"
                />
              </label>
              <label className="block text-sm md:col-span-2">
                Spell Rewards JSON
                <textarea
                  value={form.spellRewardsJson}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      spellRewardsJson: event.target.value,
                    }))
                  }
                  rows={8}
                  className="mt-1 w-full rounded border p-2 font-mono text-xs"
                />
              </label>
            </div>

            <div className="rounded bg-slate-50 p-3 text-xs text-gray-600">
              Inventory item examples: {inventoryHint || 'none loaded'}
              <br />
              Spell examples: {spellHint || 'none loaded'}
            </div>

            <div className="flex justify-end gap-3">
              <button
                type="button"
                onClick={closeModal}
                className="rounded border px-4 py-2"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={save}
                disabled={saving}
                className="rounded bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700 disabled:opacity-60"
              >
                {saving
                  ? 'Saving...'
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

export default ScenarioTemplates;
