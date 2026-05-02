import React, { useCallback, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import type { RewardProfile } from '@poltergeist/types';

type RewardProfileDraft = {
  name: string;
  slug: string;
  description: string;
  active: boolean;
  preferredItemTags: string;
  preferredMaterialKeys: string;
  preferredDamageAffinities: string;
  preferredResourceTypeIds: string;
  preferEquipment: boolean;
  preferUtility: boolean;
  preferKnowledge: boolean;
  preferNonEquipment: boolean;
};

type RewardProfilePayload = {
  name: string;
  slug: string;
  description: string;
  active: boolean;
  preferredItemTags: string[];
  preferredMaterialKeys: string[];
  preferredDamageAffinities: string[];
  preferredResourceTypeIds: string[];
  preferEquipment: boolean;
  preferUtility: boolean;
  preferKnowledge: boolean;
  preferNonEquipment: boolean;
};

const materialKeyHelp =
  'timber, stone, iron, herbs, monster_parts, arcane_dust, relic_shards';

const flagFields: Array<{
  key:
    | 'preferEquipment'
    | 'preferUtility'
    | 'preferKnowledge'
    | 'preferNonEquipment';
  label: string;
  description: string;
}> = [
  {
    key: 'preferEquipment',
    label: 'Prefer equipment',
    description: 'Boost equippable item candidates for this profile.',
  },
  {
    key: 'preferUtility',
    label: 'Prefer utility',
    description:
      'Favor consumables and tools with practical progression value.',
  },
  {
    key: 'preferKnowledge',
    label: 'Prefer knowledge',
    description:
      'Boost recipes, spell-teaching items, and other knowledge rewards.',
  },
  {
    key: 'preferNonEquipment',
    label: 'Prefer non-equipment',
    description:
      'Bias away from gear when the profile should mostly return consumables or materials.',
  },
];

const parseCsv = (value: string): string[] =>
  Array.from(
    new Set(
      value
        .split(',')
        .map((entry) => entry.trim())
        .filter(Boolean)
    )
  );

const formatCsv = (values?: string[] | null): string =>
  Array.isArray(values) ? values.join(', ') : '';

const emptyDraft = (): RewardProfileDraft => ({
  name: '',
  slug: '',
  description: '',
  active: true,
  preferredItemTags: '',
  preferredMaterialKeys: '',
  preferredDamageAffinities: '',
  preferredResourceTypeIds: '',
  preferEquipment: false,
  preferUtility: false,
  preferKnowledge: false,
  preferNonEquipment: false,
});

const draftFromRewardProfile = (
  rewardProfile: RewardProfile
): RewardProfileDraft => ({
  name: rewardProfile.name ?? '',
  slug: rewardProfile.slug ?? '',
  description: rewardProfile.description ?? '',
  active: rewardProfile.active !== false,
  preferredItemTags: formatCsv(rewardProfile.preferredItemTags),
  preferredMaterialKeys: formatCsv(rewardProfile.preferredMaterialKeys),
  preferredDamageAffinities: formatCsv(rewardProfile.preferredDamageAffinities),
  preferredResourceTypeIds: formatCsv(rewardProfile.preferredResourceTypeIds),
  preferEquipment: rewardProfile.preferEquipment === true,
  preferUtility: rewardProfile.preferUtility === true,
  preferKnowledge: rewardProfile.preferKnowledge === true,
  preferNonEquipment: rewardProfile.preferNonEquipment === true,
});

const payloadFromDraft = (draft: RewardProfileDraft): RewardProfilePayload => ({
  name: draft.name.trim(),
  slug: draft.slug.trim(),
  description: draft.description.trim(),
  active: draft.active,
  preferredItemTags: parseCsv(draft.preferredItemTags),
  preferredMaterialKeys: parseCsv(draft.preferredMaterialKeys),
  preferredDamageAffinities: parseCsv(draft.preferredDamageAffinities),
  preferredResourceTypeIds: parseCsv(draft.preferredResourceTypeIds),
  preferEquipment: draft.preferEquipment,
  preferUtility: draft.preferUtility,
  preferKnowledge: draft.preferKnowledge,
  preferNonEquipment: draft.preferNonEquipment,
});

const formatDate = (value?: string) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

type RewardProfileEditorProps = {
  draft: RewardProfileDraft;
  onChange: (next: RewardProfileDraft) => void;
  disabled?: boolean;
};

const RewardProfileEditor = ({
  draft,
  onChange,
  disabled = false,
}: RewardProfileEditorProps) => {
  const updateDraft = <K extends keyof RewardProfileDraft>(
    key: K,
    value: RewardProfileDraft[K]
  ) => onChange({ ...draft, [key]: value });

  return (
    <div className="mt-4 space-y-4">
      <div className="grid gap-3 md:grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)_140px]">
        <label className="block text-sm font-medium text-gray-700">
          Name
          <input
            type="text"
            value={draft.name}
            onChange={(event) => updateDraft('name', event.target.value)}
            disabled={disabled}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            placeholder="Combat"
          />
        </label>
        <label className="block text-sm font-medium text-gray-700">
          Slug
          <input
            type="text"
            value={draft.slug}
            onChange={(event) => updateDraft('slug', event.target.value)}
            disabled={disabled}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            placeholder="combat"
          />
        </label>
        <label className="flex items-center gap-2 pt-7 text-sm font-medium text-gray-700">
          <input
            type="checkbox"
            checked={draft.active}
            onChange={(event) => updateDraft('active', event.target.checked)}
            disabled={disabled}
            className="h-4 w-4 rounded border-gray-300"
          />
          Active
        </label>
      </div>

      <label className="block text-sm font-medium text-gray-700">
        Description
        <textarea
          value={draft.description}
          onChange={(event) => updateDraft('description', event.target.value)}
          disabled={disabled}
          rows={2}
          className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
          placeholder="Frontline and hunt-focused rewards for monster content."
        />
      </label>

      <div className="grid gap-3 md:grid-cols-2">
        <label className="block text-sm font-medium text-gray-700">
          Preferred Item Tags
          <textarea
            value={draft.preferredItemTags}
            onChange={(event) =>
              updateDraft('preferredItemTags', event.target.value)
            }
            disabled={disabled}
            rows={2}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            placeholder="martial, hunter, frontline"
          />
          <span className="mt-1 block text-xs text-gray-500">
            Comma-separated item `internalTags`.
          </span>
        </label>
        <label className="block text-sm font-medium text-gray-700">
          Preferred Material Keys
          <textarea
            value={draft.preferredMaterialKeys}
            onChange={(event) =>
              updateDraft('preferredMaterialKeys', event.target.value)
            }
            disabled={disabled}
            rows={2}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            placeholder="monster_parts, iron, herbs"
          />
          <span className="mt-1 block text-xs text-gray-500">
            Ordered keys: {materialKeyHelp}
          </span>
        </label>
        <label className="block text-sm font-medium text-gray-700">
          Preferred Damage Affinities
          <input
            type="text"
            value={draft.preferredDamageAffinities}
            onChange={(event) =>
              updateDraft('preferredDamageAffinities', event.target.value)
            }
            disabled={disabled}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            placeholder="fire, shadow"
          />
          <span className="mt-1 block text-xs text-gray-500">
            Optional affinity tags that should bias matching items.
          </span>
        </label>
        <label className="block text-sm font-medium text-gray-700">
          Preferred Resource Type IDs
          <input
            type="text"
            value={draft.preferredResourceTypeIds}
            onChange={(event) =>
              updateDraft('preferredResourceTypeIds', event.target.value)
            }
            disabled={disabled}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            placeholder="uuid-1, uuid-2"
          />
          <span className="mt-1 block text-xs text-gray-500">
            Optional gatherable resource-type UUIDs, comma-separated.
          </span>
        </label>
      </div>

      <div className="grid gap-3 md:grid-cols-2">
        {flagFields.map((field) => (
          <label
            key={field.key}
            className="flex items-start gap-3 rounded border border-gray-200 bg-slate-50 px-3 py-3 text-sm text-gray-700"
          >
            <input
              type="checkbox"
              checked={draft[field.key]}
              onChange={(event) => updateDraft(field.key, event.target.checked)}
              disabled={disabled}
              className="mt-0.5 h-4 w-4 rounded border-gray-300"
            />
            <span>
              <span className="block font-medium text-gray-900">
                {field.label}
              </span>
              <span className="block text-xs text-gray-500">
                {field.description}
              </span>
            </span>
          </label>
        ))}
      </div>
    </div>
  );
};

export const RewardProfiles = () => {
  const { apiClient } = useAPI();
  const [rewardProfiles, setRewardProfiles] = useState<RewardProfile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [createDraft, setCreateDraft] =
    useState<RewardProfileDraft>(emptyDraft);
  const [draftsById, setDraftsById] = useState<
    Record<string, RewardProfileDraft>
  >({});
  const [creating, setCreating] = useState(false);
  const [savingId, setSavingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchRewardProfiles = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<RewardProfile[]>(
        '/sonar/admin/reward-profiles'
      );
      const nextProfiles = Array.isArray(response) ? response : [];
      setRewardProfiles(nextProfiles);
      setDraftsById(
        nextProfiles.reduce<Record<string, RewardProfileDraft>>(
          (acc, rewardProfile) => {
            acc[rewardProfile.id] = draftFromRewardProfile(rewardProfile);
            return acc;
          },
          {}
        )
      );
    } catch (err) {
      console.error('Failed to load reward profiles', err);
      setError(
        err instanceof Error ? err.message : 'Failed to load reward profiles.'
      );
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void fetchRewardProfiles();
  }, [fetchRewardProfiles]);

  const handleCreate = useCallback(async () => {
    if (!createDraft.name.trim()) {
      setError('Profile name is required.');
      return;
    }
    try {
      setCreating(true);
      setError(null);
      setMessage(null);
      const created = await apiClient.post<RewardProfile>(
        '/sonar/admin/reward-profiles',
        payloadFromDraft(createDraft)
      );
      setRewardProfiles((prev) => [...prev, created]);
      setDraftsById((prev) => ({
        ...prev,
        [created.id]: draftFromRewardProfile(created),
      }));
      setCreateDraft(emptyDraft());
      setMessage(`Created ${created.name}.`);
      await fetchRewardProfiles();
    } catch (err) {
      console.error('Failed to create reward profile', err);
      setError(
        err instanceof Error ? err.message : 'Failed to create reward profile.'
      );
    } finally {
      setCreating(false);
    }
  }, [apiClient, createDraft, fetchRewardProfiles]);

  const handleSave = useCallback(
    async (rewardProfile: RewardProfile) => {
      const draft =
        draftsById[rewardProfile.id] ?? draftFromRewardProfile(rewardProfile);
      if (!draft.name.trim()) {
        setError('Profile name is required.');
        return;
      }
      try {
        setSavingId(rewardProfile.id);
        setError(null);
        setMessage(null);
        const updated = await apiClient.patch<RewardProfile>(
          `/sonar/admin/reward-profiles/${rewardProfile.id}`,
          payloadFromDraft(draft)
        );
        setRewardProfiles((prev) =>
          prev.map((entry) => (entry.id === updated.id ? updated : entry))
        );
        setDraftsById((prev) => ({
          ...prev,
          [updated.id]: draftFromRewardProfile(updated),
        }));
        setMessage(`Saved ${updated.name}.`);
        await fetchRewardProfiles();
      } catch (err) {
        console.error('Failed to save reward profile', err);
        setError(
          err instanceof Error ? err.message : 'Failed to save reward profile.'
        );
      } finally {
        setSavingId(null);
      }
    },
    [apiClient, draftsById, fetchRewardProfiles]
  );

  const handleDelete = useCallback(
    async (rewardProfile: RewardProfile) => {
      if (
        !window.confirm(
          `Delete ${rewardProfile.name}? Any content still relying on the slug will fall back to code defaults until the profile is recreated.`
        )
      ) {
        return;
      }
      try {
        setDeletingId(rewardProfile.id);
        setError(null);
        setMessage(null);
        await apiClient.delete(
          `/sonar/admin/reward-profiles/${rewardProfile.id}`
        );
        setRewardProfiles((prev) =>
          prev.filter((entry) => entry.id !== rewardProfile.id)
        );
        setDraftsById((prev) => {
          const next = { ...prev };
          delete next[rewardProfile.id];
          return next;
        });
        setMessage(`Deleted ${rewardProfile.name}.`);
      } catch (err) {
        console.error('Failed to delete reward profile', err);
        setError(
          err instanceof Error
            ? err.message
            : 'Failed to delete reward profile.'
        );
      } finally {
        setDeletingId(null);
      }
    },
    [apiClient]
  );

  return (
    <div className="m-10 space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Reward Profiles</h1>
          <p className="mt-2 max-w-3xl text-sm text-gray-600">
            Manage the named random-reward presets that content falls back to at
            runtime. These profiles define the preferred item tags, material
            families, and reward-shape flags used to bias random drops.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void fetchRewardProfiles()}
          disabled={loading}
          className="rounded bg-slate-700 px-3 py-2 text-sm text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {loading ? 'Refreshing...' : 'Refresh Profiles'}
        </button>
      </div>

      {message ? (
        <div className="rounded border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {message}
        </div>
      ) : null}

      {error ? (
        <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}

      <section className="rounded border border-gray-200 bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold text-gray-900">
              Create Profile
            </h2>
            <p className="mt-1 text-xs text-gray-600">
              Seeded profiles mirror the current runtime defaults. Create new
              ones when you want a distinct reward archetype.
            </p>
          </div>
          <button
            type="button"
            onClick={() => void handleCreate()}
            disabled={creating}
            className="rounded bg-slate-800 px-3 py-2 text-sm text-white hover:bg-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {creating ? 'Creating...' : 'Create Profile'}
          </button>
        </div>
        <RewardProfileEditor
          draft={createDraft}
          onChange={setCreateDraft}
          disabled={creating}
        />
      </section>

      <div className="space-y-4">
        {rewardProfiles.map((rewardProfile) => {
          const draft =
            draftsById[rewardProfile.id] ??
            draftFromRewardProfile(rewardProfile);
          const busy =
            savingId === rewardProfile.id || deletingId === rewardProfile.id;
          return (
            <section
              key={rewardProfile.id}
              className="rounded border border-gray-200 bg-white p-4 shadow-sm"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-lg font-semibold text-gray-900">
                      {rewardProfile.name}
                    </h2>
                    <span
                      className={`rounded-full px-2 py-1 text-xs font-medium ${
                        rewardProfile.active
                          ? 'bg-emerald-100 text-emerald-700'
                          : 'bg-slate-200 text-slate-700'
                      }`}
                    >
                      {rewardProfile.active ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                  <div className="mt-1 text-xs text-gray-500">
                    <span>Slug: {rewardProfile.slug}</span>
                    <span className="mx-2">·</span>
                    <span>Updated {formatDate(rewardProfile.updatedAt)}</span>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => void handleSave(rewardProfile)}
                    disabled={busy}
                    className="rounded bg-slate-800 px-3 py-2 text-sm text-white hover:bg-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {savingId === rewardProfile.id ? 'Saving...' : 'Save'}
                  </button>
                  <button
                    type="button"
                    onClick={() => void handleDelete(rewardProfile)}
                    disabled={busy}
                    className="rounded border border-red-300 bg-white px-3 py-2 text-sm text-red-700 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {deletingId === rewardProfile.id ? 'Deleting...' : 'Delete'}
                  </button>
                </div>
              </div>

              <RewardProfileEditor
                draft={draft}
                onChange={(next) =>
                  setDraftsById((prev) => ({
                    ...prev,
                    [rewardProfile.id]: next,
                  }))
                }
                disabled={busy}
              />
            </section>
          );
        })}
      </div>
    </div>
  );
};

export default RewardProfiles;
