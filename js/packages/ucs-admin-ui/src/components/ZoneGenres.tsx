import React, { useCallback, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import type { ZoneGenre } from '@poltergeist/types';

type ZoneGenreDraft = {
  name: string;
  sortOrder: string;
  active: boolean;
  promptSeed: string;
};

const emptyDraft = (): ZoneGenreDraft => ({
  name: '',
  sortOrder: '0',
  active: true,
  promptSeed: '',
});

const draftFromGenre = (genre: ZoneGenre): ZoneGenreDraft => ({
  name: genre.name ?? '',
  sortOrder: String(genre.sortOrder ?? 0),
  active: genre.active !== false,
  promptSeed: genre.promptSeed ?? '',
});

const parseSortOrder = (value: string) => {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : 0;
};

const formatDate = (value?: string) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

export const ZoneGenres = () => {
  const { apiClient } = useAPI();
  const [genres, setGenres] = useState<ZoneGenre[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [createDraft, setCreateDraft] = useState<ZoneGenreDraft>(emptyDraft);
  const [draftsById, setDraftsById] = useState<Record<string, ZoneGenreDraft>>(
    {}
  );
  const [creating, setCreating] = useState(false);
  const [savingId, setSavingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchGenres = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<ZoneGenre[]>(
        '/sonar/zone-genres?includeInactive=true'
      );
      const nextGenres = Array.isArray(response) ? response : [];
      setGenres(nextGenres);
      setDraftsById(
        nextGenres.reduce<Record<string, ZoneGenreDraft>>((acc, genre) => {
          acc[genre.id] = draftFromGenre(genre);
          return acc;
        }, {})
      );
    } catch (err) {
      console.error('Failed to load zone genres', err);
      setError(
        err instanceof Error ? err.message : 'Failed to load zone genres.'
      );
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void fetchGenres();
  }, [fetchGenres]);

  const handleCreate = useCallback(async () => {
    const name = createDraft.name.trim();
    if (!name) {
      setError('Genre name is required.');
      return;
    }
    try {
      setCreating(true);
      setError(null);
      setMessage(null);
      const created = await apiClient.post<ZoneGenre>(
        '/sonar/admin/zone-genres',
        {
          name,
          sortOrder: parseSortOrder(createDraft.sortOrder),
          active: createDraft.active,
          promptSeed: createDraft.promptSeed.trim(),
        }
      );
      setGenres((prev) => [...prev, created]);
      setDraftsById((prev) => ({
        ...prev,
        [created.id]: draftFromGenre(created),
      }));
      setCreateDraft(emptyDraft());
      setMessage(`Created ${created.name}.`);
      await fetchGenres();
    } catch (err) {
      console.error('Failed to create zone genre', err);
      setError(
        err instanceof Error ? err.message : 'Failed to create zone genre.'
      );
    } finally {
      setCreating(false);
    }
  }, [apiClient, createDraft, fetchGenres]);

  const handleSave = useCallback(
    async (genre: ZoneGenre) => {
      const draft = draftsById[genre.id] ?? draftFromGenre(genre);
      const name = draft.name.trim();
      if (!name) {
        setError('Genre name is required.');
        return;
      }
      try {
        setSavingId(genre.id);
        setError(null);
        setMessage(null);
        const updated = await apiClient.patch<ZoneGenre>(
          `/sonar/admin/zone-genres/${genre.id}`,
          {
            name,
            sortOrder: parseSortOrder(draft.sortOrder),
            active: draft.active,
            promptSeed: draft.promptSeed.trim(),
          }
        );
        setGenres((prev) =>
          prev.map((entry) => (entry.id === updated.id ? updated : entry))
        );
        setDraftsById((prev) => ({
          ...prev,
          [updated.id]: draftFromGenre(updated),
        }));
        setMessage(`Saved ${updated.name}.`);
        await fetchGenres();
      } catch (err) {
        console.error('Failed to save zone genre', err);
        setError(
          err instanceof Error ? err.message : 'Failed to save zone genre.'
        );
      } finally {
        setSavingId(null);
      }
    },
    [apiClient, draftsById, fetchGenres]
  );

  const handleDelete = useCallback(
    async (genre: ZoneGenre) => {
      if (
        !window.confirm(
          `Delete ${genre.name}? This will remove its zone score rows and may fail if monsters, scenarios, characters, or templates still use it.`
        )
      ) {
        return;
      }
      try {
        setDeletingId(genre.id);
        setError(null);
        setMessage(null);
        await apiClient.delete(`/sonar/admin/zone-genres/${genre.id}`);
        setGenres((prev) => prev.filter((entry) => entry.id !== genre.id));
        setDraftsById((prev) => {
          const next = { ...prev };
          delete next[genre.id];
          return next;
        });
        setMessage(`Deleted ${genre.name}.`);
      } catch (err) {
        console.error('Failed to delete zone genre', err);
        setError(
          err instanceof Error ? err.message : 'Failed to delete zone genre.'
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
          <h1 className="text-2xl font-bold">Zone Genres</h1>
          <p className="mt-2 max-w-3xl text-sm text-gray-600">
            Manage the shared genres used by the Chaos Engine Room, monster
            generation, scenario generation, character portraits, and other
            shared content systems. Prompt seeds steer how AI-generated content
            should feel for each genre.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void fetchGenres()}
          disabled={loading}
          className="rounded bg-slate-700 px-3 py-2 text-sm text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {loading ? 'Refreshing...' : 'Refresh Genres'}
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
              Create Genre
            </h2>
            <p className="mt-1 text-xs text-gray-600">
              Examples: Fantasy, Science Fiction, Gothic Horror.
            </p>
          </div>
          <button
            type="button"
            onClick={() => void handleCreate()}
            disabled={creating}
            className="rounded bg-slate-800 px-3 py-2 text-sm text-white hover:bg-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {creating ? 'Creating...' : 'Create Genre'}
          </button>
        </div>
        <div className="mt-4 grid gap-3 md:grid-cols-[minmax(0,2fr)_140px_140px]">
          <label className="block text-sm font-medium text-gray-700">
            Name
            <input
              type="text"
              value={createDraft.name}
              onChange={(event) =>
                setCreateDraft((prev) => ({
                  ...prev,
                  name: event.target.value,
                }))
              }
              className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
              placeholder="Fantasy"
            />
          </label>
          <label className="block text-sm font-medium text-gray-700">
            Sort Order
            <input
              type="number"
              value={createDraft.sortOrder}
              onChange={(event) =>
                setCreateDraft((prev) => ({
                  ...prev,
                  sortOrder: event.target.value,
                }))
              }
              className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            />
          </label>
          <label className="flex items-end gap-2 rounded border border-gray-200 px-3 py-2 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={createDraft.active}
              onChange={(event) =>
                setCreateDraft((prev) => ({
                  ...prev,
                  active: event.target.checked,
                }))
              }
            />
            Active
          </label>
        </div>
        <label className="mt-3 block text-sm font-medium text-gray-700">
          Prompt Seed
          <textarea
            value={createDraft.promptSeed}
            onChange={(event) =>
              setCreateDraft((prev) => ({
                ...prev,
                promptSeed: event.target.value,
              }))
            }
            className="mt-1 min-h-[120px] w-full rounded border border-gray-300 px-3 py-2 text-sm"
            placeholder="Describe the tone, motifs, aesthetics, and creature logic for this genre."
          />
        </label>
      </section>

      <section className="rounded border border-gray-200 bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold text-gray-900">
              Configured Genres
            </h2>
            <p className="mt-1 text-xs text-gray-600">
              {genres.length} genre{genres.length === 1 ? '' : 's'} configured.
            </p>
          </div>
        </div>

        {loading ? (
          <div className="mt-4 text-sm text-gray-500">Loading genres...</div>
        ) : genres.length === 0 ? (
          <div className="mt-4 text-sm text-gray-500">
            No genres configured yet.
          </div>
        ) : (
          <div className="mt-4 space-y-3">
            {genres.map((genre) => {
              const draft = draftsById[genre.id] ?? draftFromGenre(genre);
              return (
                <div
                  key={genre.id}
                  className="rounded border border-gray-200 bg-gray-50 p-4"
                >
                  <div className="grid gap-3 lg:grid-cols-[minmax(0,2fr)_140px_140px_auto]">
                    <label className="block text-sm font-medium text-gray-700">
                      Name
                      <input
                        type="text"
                        value={draft.name}
                        onChange={(event) =>
                          setDraftsById((prev) => ({
                            ...prev,
                            [genre.id]: {
                              ...draft,
                              name: event.target.value,
                            },
                          }))
                        }
                        className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                      />
                    </label>
                    <label className="block text-sm font-medium text-gray-700">
                      Sort Order
                      <input
                        type="number"
                        value={draft.sortOrder}
                        onChange={(event) =>
                          setDraftsById((prev) => ({
                            ...prev,
                            [genre.id]: {
                              ...draft,
                              sortOrder: event.target.value,
                            },
                          }))
                        }
                        className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                      />
                    </label>
                    <label className="flex items-end gap-2 rounded border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700">
                      <input
                        type="checkbox"
                        checked={draft.active}
                        onChange={(event) =>
                          setDraftsById((prev) => ({
                            ...prev,
                            [genre.id]: {
                              ...draft,
                              active: event.target.checked,
                            },
                          }))
                        }
                      />
                      Active
                    </label>
                    <div className="flex items-end justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => void handleSave(genre)}
                        disabled={savingId === genre.id}
                        className="rounded bg-slate-800 px-3 py-2 text-sm text-white hover:bg-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {savingId === genre.id ? 'Saving...' : 'Save'}
                      </button>
                      <button
                        type="button"
                        onClick={() => void handleDelete(genre)}
                        disabled={deletingId === genre.id}
                        className="rounded border border-red-200 px-3 py-2 text-sm text-red-700 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {deletingId === genre.id ? 'Deleting...' : 'Delete'}
                      </button>
                    </div>
                  </div>
                  <label className="mt-3 block text-sm font-medium text-gray-700">
                    Prompt Seed
                    <textarea
                      value={draft.promptSeed}
                      onChange={(event) =>
                        setDraftsById((prev) => ({
                          ...prev,
                          [genre.id]: {
                            ...draft,
                            promptSeed: event.target.value,
                          },
                        }))
                      }
                      className="mt-1 min-h-[120px] w-full rounded border border-gray-300 px-3 py-2 text-sm"
                      placeholder="Describe the tone, motifs, aesthetics, and creature logic for this genre."
                    />
                  </label>
                  <div className="mt-3 text-xs text-gray-500">
                    Updated {formatDate(genre.updatedAt)}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
};

export default ZoneGenres;
