import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useTagContext } from '@poltergeist/contexts';
import { Tag } from '@poltergeist/types';

type SelectOption = {
  value: string;
  label: string;
};

type ShopkeeperSeedCandidate = {
  tag: string;
  weight: number;
};

type ShopkeeperSeedProfileResponse = {
  category: string;
  label: string;
  spawnChanceBasisPoints: number;
  candidates: ShopkeeperSeedCandidate[];
};

type ShopkeeperSeedConfigResponse = {
  id: number;
  profiles: ShopkeeperSeedProfileResponse[];
};

type CandidateRow = {
  id: string;
  tag: string;
  weight: number;
};

type EditableProfile = {
  category: string;
  label: string;
  spawnChancePercent: number;
  candidates: CandidateRow[];
};

const flattenTags = (tagGroups: { tags: Tag[] }[]): Tag[] => {
  return tagGroups.flatMap((group) => group.tags);
};

const createRowId = () =>
  `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;

const makeCandidateRow = (tag = '', weight = 1): CandidateRow => ({
  id: createRowId(),
  tag,
  weight,
});

const basisPointsToPercent = (basisPoints: number) => {
  if (!Number.isFinite(basisPoints)) return 0;
  return Number((basisPoints / 100).toFixed(2));
};

const percentToBasisPoints = (percent: number) => {
  if (!Number.isFinite(percent)) return 0;
  const clamped = Math.max(0, Math.min(100, percent));
  return Math.round(clamped * 100);
};

const extractApiErrorMessage = (error: unknown, fallback: string): string => {
  if (
    typeof error === 'object' &&
    error !== null &&
    'response' in error &&
    typeof (error as { response?: unknown }).response === 'object'
  ) {
    const response = (error as { response?: { data?: unknown } }).response;
    const data = response?.data;
    if (typeof data === 'object' && data !== null) {
      const maybeMessage = (data as { error?: unknown; message?: unknown })
        .error;
      if (typeof maybeMessage === 'string' && maybeMessage.trim() !== '') {
        return maybeMessage;
      }
      const maybeFallback = (data as { message?: unknown }).message;
      if (typeof maybeFallback === 'string' && maybeFallback.trim() !== '') {
        return maybeFallback;
      }
    }
  }
  if (error instanceof Error && error.message.trim() !== '') {
    return error.message;
  }
  return fallback;
};

const SearchableSelect = ({
  label,
  placeholder,
  options,
  value,
  onChange,
}: {
  label: string;
  placeholder: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const selected = options.find((option) => option.value === value);
  const filtered = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    if (!normalizedQuery) return options;
    return options.filter((option) =>
      option.label.toLowerCase().includes(normalizedQuery)
    );
  }, [options, query]);

  const displayValue = open ? query : selected?.label ?? value;

  return (
    <div className="relative">
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      <input
        value={displayValue}
        onChange={(event) => {
          setQuery(event.target.value);
          setOpen(true);
        }}
        onFocus={() => {
          setOpen(true);
          setQuery('');
        }}
        onBlur={() => {
          window.setTimeout(() => setOpen(false), 150);
        }}
        placeholder={placeholder}
        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
      />
      {open && (
        <div className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg">
          {filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-500">
              No matches found
            </div>
          )}
          {filtered.map((option) => (
            <button
              type="button"
              key={option.value}
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                onChange(option.value);
                setOpen(false);
                setQuery('');
              }}
              className="flex w-full items-center px-3 py-2 text-left text-sm hover:bg-indigo-50"
            >
              <span className="font-medium text-gray-900">{option.label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

const mapResponseToProfiles = (
  profiles: ShopkeeperSeedProfileResponse[]
): EditableProfile[] => {
  return (profiles || []).map((profile) => ({
    category: profile.category,
    label: profile.label,
    spawnChancePercent: basisPointsToPercent(
      profile.spawnChanceBasisPoints ?? 0
    ),
    candidates: Array.isArray(profile.candidates)
      ? profile.candidates.map((candidate) =>
          makeCandidateRow(candidate.tag ?? '', candidate.weight ?? 1)
        )
      : [],
  }));
};

export const PoiShopkeeperSeedConfigPanel = () => {
  const { apiClient } = useAPI();
  const { tagGroups } = useTagContext();
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [profiles, setProfiles] = useState<EditableProfile[]>([]);

  const tagOptions = useMemo(() => {
    const names = flattenTags(tagGroups)
      .map((tag) => tag.name.trim())
      .filter((name): name is string => name.length > 0);
    return Array.from(new Set(names))
      .sort((left, right) => left.localeCompare(right))
      .map((name) => ({
        value: name,
        label: name,
      }));
  }, [tagGroups]);

  const loadConfig = useCallback(
    async (showMessage = false) => {
      try {
        setLoading(true);
        setError(null);
        const response = await apiClient.get<ShopkeeperSeedConfigResponse>(
          '/sonar/admin/point-of-interest-shopkeeper-seed-config'
        );
        setProfiles(mapResponseToProfiles(response?.profiles || []));
        if (showMessage) {
          setMessage('Reloaded saved shopkeeper profiles.');
        }
      } catch (nextError) {
        console.error(
          'Failed to load point of interest shopkeeper seed config',
          nextError
        );
        setError(
          extractApiErrorMessage(
            nextError,
            'Failed to load point of interest shopkeeper seed config.'
          )
        );
      } finally {
        setLoading(false);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    void loadConfig();
  }, [loadConfig]);

  const updateProfile = useCallback(
    (category: string, updater: (profile: EditableProfile) => EditableProfile) => {
      setProfiles((previous) =>
        previous.map((profile) =>
          profile.category === category ? updater(profile) : profile
        )
      );
    },
    []
  );

  const handleSave = useCallback(async () => {
    try {
      setSaving(true);
      setError(null);
      setMessage(null);

      const payload = {
        profiles: profiles.map((profile) => ({
          category: profile.category,
          spawnChanceBasisPoints: percentToBasisPoints(
            profile.spawnChancePercent
          ),
          candidates: profile.candidates
            .filter(
              (candidate) =>
                candidate.tag.trim().length > 0 && candidate.weight > 0
            )
            .map((candidate) => ({
              tag: candidate.tag,
              weight: Math.max(1, Math.round(candidate.weight)),
            })),
        })),
      };

      const response = await apiClient.put<ShopkeeperSeedConfigResponse>(
        '/sonar/admin/point-of-interest-shopkeeper-seed-config',
        payload
      );
      setProfiles(mapResponseToProfiles(response?.profiles || []));
      setMessage('Point of interest shopkeeper profiles saved.');
    } catch (nextError) {
      console.error(
        'Failed to save point of interest shopkeeper seed config',
        nextError
      );
        setError(
          extractApiErrorMessage(
            nextError,
            'Failed to save point of interest shopkeeper profiles.'
          )
        );
    } finally {
      setSaving(false);
    }
  }, [apiClient, profiles]);

  return (
    <div className="mb-6 rounded-lg bg-white p-4 shadow-md">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div className="max-w-3xl">
          <h2 className="text-lg font-semibold text-gray-900">
            POI Shopkeeper Seeding
          </h2>
          <p className="mt-1 text-sm text-gray-600">
            Configure which shopkeeper tags can roll for each point of interest
            marker category during zone seeding. Each POI gets at most one
            shopkeeper, and weights decide which tag wins when the category
            spawn chance hits.
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={() => setOpen((previous) => !previous)}
            className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            {open ? 'Hide Editor' : 'Show Editor'}
          </button>
          {open && (
            <>
              <button
                type="button"
                onClick={() => void loadConfig(true)}
                disabled={loading || saving}
                className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Reload Saved
              </button>
              <button
                type="button"
                onClick={() => void handleSave()}
                disabled={loading || saving}
                className="rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {saving ? 'Saving...' : 'Save Profiles'}
              </button>
            </>
          )}
        </div>
      </div>

      {error && (
        <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </div>
      )}
      {message && (
        <div className="mt-4 rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
          {message}
        </div>
      )}

      {open && (
        <div className="mt-4">
          {loading ? (
            <div className="text-sm text-gray-500">
              Loading shopkeeper seeding profiles...
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
              {profiles.map((profile) => (
                <div
                  key={profile.category}
                  className="rounded-lg border border-gray-200 bg-gray-50 p-4"
                >
                  <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                    <div>
                      <h3 className="text-base font-semibold text-gray-900">
                        {profile.label}
                      </h3>
                      <p className="text-xs uppercase tracking-wide text-gray-500">
                        {profile.category}
                      </p>
                    </div>
                    <div className="w-full md:w-40">
                      <label className="block text-sm font-medium text-gray-700">
                        Spawn Chance (%)
                      </label>
                      <input
                        type="number"
                        min="0"
                        max="100"
                        step="0.1"
                        value={profile.spawnChancePercent}
                        onChange={(event) =>
                          updateProfile(profile.category, (current) => ({
                            ...current,
                            spawnChancePercent:
                              Number.parseFloat(event.target.value) || 0,
                          }))
                        }
                        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                    </div>
                  </div>

                  <div className="mt-4 flex items-center justify-between">
                    <div>
                      <h4 className="text-sm font-semibold text-gray-800">
                        Candidate Tags
                      </h4>
                      <p className="text-xs text-gray-500">
                        Higher weights make that shopkeeper tag more likely when
                        this category spawns a patron.
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() =>
                        updateProfile(profile.category, (current) => ({
                          ...current,
                          candidates: [...current.candidates, makeCandidateRow()],
                        }))
                      }
                      className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                    >
                      Add Tag
                    </button>
                  </div>

                  {profile.candidates.length === 0 ? (
                    <div className="mt-3 rounded-md border border-dashed border-gray-300 bg-white px-3 py-4 text-sm text-gray-500">
                      No candidate shopkeeper tags configured for this category.
                    </div>
                  ) : (
                    <div className="mt-3 space-y-3">
                      {profile.candidates.map((candidate) => (
                        <div
                          key={candidate.id}
                          className="rounded-md border border-gray-200 bg-white p-3"
                        >
                          <div className="grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,1fr)_120px_auto]">
                            <SearchableSelect
                              label="Shopkeeper Tag"
                              placeholder="Search tag..."
                              options={tagOptions}
                              value={candidate.tag}
                              onChange={(value) =>
                                updateProfile(profile.category, (current) => ({
                                  ...current,
                                  candidates: current.candidates.map((row) =>
                                    row.id === candidate.id
                                      ? { ...row, tag: value }
                                      : row
                                  ),
                                }))
                              }
                            />
                            <div>
                              <label className="block text-sm font-medium text-gray-700">
                                Weight
                              </label>
                              <input
                                type="number"
                                min="1"
                                step="1"
                                value={candidate.weight}
                                onChange={(event) =>
                                  updateProfile(
                                    profile.category,
                                    (current) => ({
                                      ...current,
                                      candidates: current.candidates.map((row) =>
                                        row.id === candidate.id
                                          ? {
                                              ...row,
                                              weight:
                                                Number.parseInt(
                                                  event.target.value,
                                                  10
                                                ) || 1,
                                            }
                                          : row
                                      ),
                                    })
                                  )
                                }
                                className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                              />
                            </div>
                            <div className="flex items-end">
                              <button
                                type="button"
                                onClick={() =>
                                  updateProfile(
                                    profile.category,
                                    (current) => ({
                                      ...current,
                                      candidates: current.candidates.filter(
                                        (row) => row.id !== candidate.id
                                      ),
                                    })
                                  )
                                }
                                className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm font-medium text-red-700 hover:bg-red-100"
                              >
                                Remove
                              </button>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default PoiShopkeeperSeedConfigPanel;
