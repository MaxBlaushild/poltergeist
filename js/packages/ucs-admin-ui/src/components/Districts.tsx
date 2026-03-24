import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { District } from '@poltergeist/types';
import { Link, useNavigate } from 'react-router-dom';

export const Districts = () => {
  const { apiClient } = useAPI();
  const navigate = useNavigate();
  const [districts, setDistricts] = useState<District[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');

  const loadDistricts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<District[]>('/sonar/districts');
      setDistricts(response);
    } catch (err) {
      console.error('Error fetching districts:', err);
      setError('Unable to load districts right now.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void loadDistricts();
  }, [loadDistricts]);

  const filteredDistricts = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) {
      return districts;
    }

    return districts.filter((district) => {
      const haystack = `${district.name} ${district.description}`.toLowerCase();
      return haystack.includes(query);
    });
  }, [districts, search]);

  const handleCreateDistrict = async (
    event: React.FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault();
    const name = newName.trim();
    if (!name) {
      return;
    }

    setIsCreating(true);
    setError(null);
    try {
      const created = await apiClient.post<District>('/sonar/districts', {
        name,
        description: newDescription.trim(),
        zoneIds: [],
      });
      setNewName('');
      setNewDescription('');
      navigate(`/districts/${created.id}`);
    } catch (err) {
      console.error('Error creating district:', err);
      setError('Unable to create that district right now.');
    } finally {
      setIsCreating(false);
    }
  };

  const handleDeleteDistrict = async (district: District) => {
    const confirmed = window.confirm(`Delete district "${district.name}"?`);
    if (!confirmed) {
      return;
    }

    try {
      await apiClient.delete(`/sonar/districts/${district.id}`);
      setDistricts((current) =>
        current.filter((item) => item.id !== district.id)
      );
    } catch (err) {
      console.error('Error deleting district:', err);
      setError('Unable to delete that district right now.');
    }
  };

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6 p-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold text-gray-900">Districts</h1>
        <p className="text-sm text-gray-600">
          Districts are curated collections of zones. Use them to organize
          neighboring zones into a larger region for admin planning and future
          content.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[360px_minmax(0,1fr)]">
        <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900">
            Create district
          </h2>
          <p className="mt-1 text-sm text-gray-500">
            Start with a name, then open the editor to select zones from the
            map.
          </p>
          <form
            className="mt-4 flex flex-col gap-3"
            onSubmit={handleCreateDistrict}
          >
            <input
              className="rounded-lg border border-gray-300 px-3 py-2 text-sm"
              placeholder="District name"
              value={newName}
              onChange={(event) => setNewName(event.target.value)}
              required
            />
            <textarea
              className="min-h-[120px] rounded-lg border border-gray-300 px-3 py-2 text-sm"
              placeholder="Short description"
              value={newDescription}
              onChange={(event) => setNewDescription(event.target.value)}
            />
            <button
              type="submit"
              className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
              disabled={isCreating}
            >
              {isCreating ? 'Creating...' : 'Create district'}
            </button>
          </form>
        </div>

        <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
          <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">
                Existing districts
              </h2>
              <p className="text-sm text-gray-500">
                Click a district to edit its zone collection and metadata.
              </p>
            </div>
            <input
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm md:max-w-xs"
              placeholder="Search districts"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>

          {error && (
            <div className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          )}

          {loading ? (
            <div className="mt-6 text-sm text-gray-500">
              Loading districts...
            </div>
          ) : filteredDistricts.length === 0 ? (
            <div className="mt-6 rounded-lg border border-dashed border-gray-300 px-4 py-8 text-center text-sm text-gray-500">
              No districts found.
            </div>
          ) : (
            <div className="mt-4 grid gap-4">
              {filteredDistricts.map((district) => (
                <div
                  key={district.id}
                  className="rounded-xl border border-gray-200 bg-gray-50 p-4 transition hover:border-gray-300 hover:bg-white"
                >
                  <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                    <div className="min-w-0 flex-1">
                      <Link
                        to={`/districts/${district.id}`}
                        className="text-lg font-semibold text-slate-900 hover:text-blue-700"
                      >
                        {district.name}
                      </Link>
                      <p className="mt-1 text-sm text-gray-600">
                        {district.description || 'No description yet.'}
                      </p>
                      <div className="mt-3 flex flex-wrap gap-2 text-xs text-gray-500">
                        <span className="rounded-full bg-white px-2 py-1">
                          {district.zones?.length || 0} zones
                        </span>
                        <span className="rounded-full bg-white px-2 py-1">
                          ID: {district.id}
                        </span>
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => void handleDeleteDistrict(district)}
                      className="rounded-lg border border-red-200 px-3 py-2 text-sm font-medium text-red-700 hover:bg-red-50"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
