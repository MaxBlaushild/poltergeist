import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

type ChallengeRecord = {
  id: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  question: string;
  imageUrl: string;
  thumbnailUrl: string;
  reward: number;
  inventoryItemId?: number | null;
  submissionType: 'photo' | 'text' | 'video';
  difficulty: number;
  statTags: string[];
  proficiency?: string | null;
};

type ChallengeFormState = {
  zoneId: string;
  latitude: string;
  longitude: string;
  question: string;
  imageUrl: string;
  thumbnailUrl: string;
  reward: string;
  inventoryItemId: string;
  submissionType: 'photo' | 'text' | 'video';
  difficulty: string;
  statTags: string;
  proficiency: string;
};

type ImagePreviewState = {
  url: string;
  alt: string;
};

const statTagOptions = [
  'strength',
  'dexterity',
  'constitution',
  'intelligence',
  'wisdom',
  'charisma',
] as const;

const parseIntSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseFloatSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseOptionalInt = (value: string): number | undefined => {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number.parseInt(trimmed, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const parseCsv = (value: string): string[] =>
  value
    .split(',')
    .map((entry) => entry.trim().toLowerCase())
    .filter((entry) => statTagOptions.includes(entry as (typeof statTagOptions)[number]));

const emptyForm = (): ChallengeFormState => ({
  zoneId: '',
  latitude: '',
  longitude: '',
  question: '',
  imageUrl: '',
  thumbnailUrl: '',
  reward: '0',
  inventoryItemId: '',
  submissionType: 'photo',
  difficulty: '0',
  statTags: '',
  proficiency: '',
});

const formFromRecord = (record: ChallengeRecord): ChallengeFormState => ({
  zoneId: record.zoneId ?? '',
  latitude: String(record.latitude ?? ''),
  longitude: String(record.longitude ?? ''),
  question: record.question ?? '',
  imageUrl: record.imageUrl ?? '',
  thumbnailUrl: record.thumbnailUrl ?? '',
  reward: String(record.reward ?? 0),
  inventoryItemId:
    record.inventoryItemId !== undefined && record.inventoryItemId !== null
      ? String(record.inventoryItemId)
      : '',
  submissionType:
    record.submissionType === 'text' || record.submissionType === 'video'
      ? record.submissionType
      : 'photo',
  difficulty: String(record.difficulty ?? 0),
  statTags: (record.statTags ?? []).join(', '),
  proficiency: record.proficiency ?? '',
});

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

export const Challenges = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { inventoryItems } = useInventory();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [records, setRecords] = useState<ChallengeRecord[]>([]);
  const [query, setQuery] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editingChallenge, setEditingChallenge] = useState<ChallengeRecord | null>(
    null
  );
  const [form, setForm] = useState<ChallengeFormState>(emptyForm);
  const [geoLoading, setGeoLoading] = useState(false);
  const [generatingChallengeId, setGeneratingChallengeId] = useState<
    string | null
  >(null);
  const [imagePreview, setImagePreview] = useState<ImagePreviewState | null>(
    null
  );

  const mapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const mapRef = React.useRef<mapboxgl.Map | null>(null);
  const markerRef = React.useRef<mapboxgl.Marker | null>(null);
  const formLatitudeRef = React.useRef(form.latitude);
  const formLongitudeRef = React.useRef(form.longitude);

  const zoneNameById = useMemo(() => {
    return new Map(zones.map((zone) => [zone.id, zone.name]));
  }, [zones]);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<ChallengeRecord[]>('/sonar/challenges');
      setRecords(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to load challenges', err);
      setError('Failed to load challenges.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  const refreshChallengeById = useCallback(
    async (challengeId: string) => {
      const latest = await apiClient.get<ChallengeRecord>(
        `/sonar/challenges/${challengeId}`
      );
      setRecords((prev) =>
        prev.map((record) => (record.id === challengeId ? latest : record))
      );
      return latest;
    },
    [apiClient]
  );

  useEffect(() => {
    void load();
  }, [load]);

  const filteredRecords = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return records;
    return records.filter((record) => {
      const zoneName = zoneNameById.get(record.zoneId) ?? '';
      return (
        record.question.toLowerCase().includes(normalized) ||
        zoneName.toLowerCase().includes(normalized) ||
        record.id.toLowerCase().includes(normalized)
      );
    });
  }, [query, records, zoneNameById]);

  const openCreate = () => {
    setEditingChallenge(null);
    setForm({
      ...emptyForm(),
      zoneId: zones[0]?.id ?? '',
    });
    setShowModal(true);
  };

  const openEdit = (record: ChallengeRecord) => {
    setEditingChallenge(record);
    setForm(formFromRecord(record));
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingChallenge(null);
    setForm(emptyForm());
  };

  useEffect(() => {
    formLatitudeRef.current = form.latitude;
    formLongitudeRef.current = form.longitude;
  }, [form.latitude, form.longitude]);

  const setFormLocation = useCallback((latitude: number, longitude: number) => {
    setForm((prev) => ({
      ...prev,
      latitude: latitude.toFixed(6),
      longitude: longitude.toFixed(6),
    }));
  }, []);

  const handleUseCurrentLocation = useCallback(() => {
    if (!navigator.geolocation) {
      alert('Geolocation is not supported in this browser.');
      return;
    }
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGeoLoading(false);
        setFormLocation(position.coords.latitude, position.coords.longitude);
      },
      (geoError) => {
        setGeoLoading(false);
        alert(`Unable to get current location: ${geoError.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setFormLocation]);

  useEffect(() => {
    if (!showModal) return;
    if (!mapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (mapRef.current) return;

    const parsedLat = Number.parseFloat(formLatitudeRef.current);
    const parsedLng = Number.parseFloat(formLongitudeRef.current);
    const selectedZone = zones.find((zone) => zone.id === form.zoneId);
    const zoneLat = selectedZone
      ? Number.parseFloat(String(selectedZone.latitude ?? ''))
      : Number.NaN;
    const zoneLng = selectedZone
      ? Number.parseFloat(String(selectedZone.longitude ?? ''))
      : Number.NaN;

    const center: [number, number] =
      Number.isFinite(parsedLat) && Number.isFinite(parsedLng)
        ? [parsedLng, parsedLat]
        : Number.isFinite(zoneLat) && Number.isFinite(zoneLng)
          ? [zoneLng, zoneLat]
          : [-73.98513, 40.7589];

    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center,
      zoom: 13,
    });

    map.on('click', (event) => {
      setFormLocation(event.lngLat.lat, event.lngLat.lng);
    });

    mapRef.current = map;

    return () => {
      markerRef.current?.remove();
      markerRef.current = null;
      map.remove();
      mapRef.current = null;
    };
  }, [form.zoneId, setFormLocation, showModal, zones]);

  useEffect(() => {
    if (!showModal) return;
    if (!mapRef.current) return;

    const lat = Number.parseFloat(form.latitude);
    const lng = Number.parseFloat(form.longitude);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
      markerRef.current?.remove();
      markerRef.current = null;
      return;
    }

    if (!markerRef.current) {
      markerRef.current = new mapboxgl.Marker({ color: '#dc2626' })
        .setLngLat([lng, lat])
        .addTo(mapRef.current);
    } else {
      markerRef.current.setLngLat([lng, lat]);
    }

    mapRef.current.easeTo({ center: [lng, lat], duration: 350 });
  }, [form.latitude, form.longitude, showModal]);

  const save = async () => {
    try {
      const payload = {
        zoneId: form.zoneId.trim(),
        latitude: parseFloatSafe(form.latitude, 0),
        longitude: parseFloatSafe(form.longitude, 0),
        question: form.question.trim(),
        imageUrl: form.imageUrl.trim(),
        thumbnailUrl: form.thumbnailUrl.trim(),
        reward: parseIntSafe(form.reward, 0),
        inventoryItemId: parseOptionalInt(form.inventoryItemId),
        submissionType: form.submissionType,
        difficulty: parseIntSafe(form.difficulty, 0),
        statTags: parseCsv(form.statTags),
        proficiency: form.proficiency.trim(),
      };

      if (!payload.zoneId || !payload.question) {
        alert('Zone and question are required.');
        return;
      }
      if (!payload.thumbnailUrl && payload.imageUrl) {
        payload.thumbnailUrl = payload.imageUrl;
      }

      if (editingChallenge) {
        const updated = await apiClient.put<ChallengeRecord>(
          `/sonar/challenges/${editingChallenge.id}`,
          payload
        );
        setRecords((prev) =>
          prev.map((record) => (record.id === updated.id ? updated : record))
        );
      } else {
        const created = await apiClient.post<ChallengeRecord>(
          '/sonar/challenges',
          payload
        );
        setRecords((prev) => [created, ...prev]);
      }
      closeModal();
    } catch (err) {
      console.error('Failed to save challenge', err);
      const message = err instanceof Error ? err.message : 'Failed to save challenge.';
      alert(message);
    }
  };

  const deleteChallenge = async (record: ChallengeRecord) => {
    if (!window.confirm(`Delete challenge "${record.question}"?`)) return;
    try {
      await apiClient.delete(`/sonar/challenges/${record.id}`);
      setRecords((prev) => prev.filter((entry) => entry.id !== record.id));
    } catch (err) {
      console.error('Failed to delete challenge', err);
      alert('Failed to delete challenge.');
    }
  };

  const handleGenerateImage = async (record: ChallengeRecord) => {
    if (generatingChallengeId) return;
    setGeneratingChallengeId(record.id);
    const previousImageURL = (record.imageUrl || record.thumbnailUrl || '').trim();
    try {
      await apiClient.post(`/sonar/challenges/${record.id}/generate-image`, {});

      for (let attempt = 0; attempt < 18; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 1200));
        const latest = await refreshChallengeById(record.id);
        const nextImageURL = (latest.imageUrl || latest.thumbnailUrl || '').trim();
        if (nextImageURL && nextImageURL !== previousImageURL) {
          break;
        }
      }
    } catch (err) {
      console.error('Failed to queue challenge image generation', err);
      alert('Failed to queue challenge image generation.');
    } finally {
      setGeneratingChallengeId(null);
    }
  };

  const openImagePreview = (record: ChallengeRecord) => {
    const url = record.thumbnailUrl || record.imageUrl;
    if (!url) return;
    setImagePreview({
      url,
      alt: `Challenge ${record.question || record.id} image`,
    });
  };

  const closeImagePreview = () => {
    setImagePreview(null);
  };

  if (loading) {
    return <div className="m-10">Loading challenges...</div>;
  }

  return (
    <div className="m-10">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Challenges</h1>
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={openCreate}
        >
          Create Challenge
        </button>
      </div>

      <div className="mb-4">
        <input
          className="w-full max-w-xl border rounded-md p-2"
          placeholder="Search by question, zone, or id"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
      </div>

      {error ? (
        <div className="mb-4 text-sm text-red-600">{error}</div>
      ) : null}

      <div className="overflow-auto border rounded-md">
        <table className="min-w-full text-sm">
          <thead className="bg-gray-100">
            <tr>
              <th className="text-left p-2 border-b">Question</th>
              <th className="text-left p-2 border-b">Zone</th>
              <th className="text-left p-2 border-b">Submission</th>
              <th className="text-left p-2 border-b">Difficulty</th>
              <th className="text-left p-2 border-b">Reward</th>
              <th className="text-left p-2 border-b">Location</th>
              <th className="text-left p-2 border-b">Image</th>
              <th className="text-left p-2 border-b">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredRecords.length === 0 ? (
              <tr>
                <td className="p-3 text-gray-500" colSpan={8}>
                  No challenges found.
                </td>
              </tr>
            ) : (
              filteredRecords.map((record) => (
                <tr key={record.id} className="odd:bg-white even:bg-gray-50">
                  <td className="p-2 border-b align-top max-w-md">
                    <div className="font-medium">{record.question}</div>
                    <div className="text-xs text-gray-500 font-mono mt-1">
                      {record.id}
                    </div>
                    {record.statTags?.length ? (
                      <div className="text-xs text-gray-600 mt-1">
                        Stats: {record.statTags.join(', ')}
                      </div>
                    ) : null}
                  </td>
                  <td className="p-2 border-b align-top">
                    {zoneNameById.get(record.zoneId) ?? record.zoneId}
                  </td>
                  <td className="p-2 border-b align-top">{record.submissionType}</td>
                  <td className="p-2 border-b align-top">{record.difficulty}</td>
                  <td className="p-2 border-b align-top">
                    {record.reward}
                    {record.inventoryItemId ? (
                      <div className="text-xs text-gray-600">
                        Item #{record.inventoryItemId}
                      </div>
                    ) : null}
                  </td>
                  <td className="p-2 border-b align-top">
                    {Number.isFinite(record.latitude) &&
                    Number.isFinite(record.longitude)
                      ? `${record.latitude.toFixed(5)}, ${record.longitude.toFixed(5)}`
                      : 'n/a'}
                  </td>
                  <td className="p-2 border-b align-top">
                    {record.thumbnailUrl || record.imageUrl ? (
                      <button
                        type="button"
                        className="inline-flex"
                        onClick={() => openImagePreview(record)}
                        title="Open image preview"
                      >
                        <img
                          src={record.thumbnailUrl || record.imageUrl}
                          alt="Challenge"
                          className="w-12 h-12 object-cover rounded border"
                        />
                      </button>
                    ) : (
                      <span className="text-xs text-gray-500">No image</span>
                    )}
                  </td>
                  <td className="p-2 border-b align-top">
                    <div className="flex flex-col gap-2">
                      <button
                        type="button"
                        className="bg-indigo-600 text-white px-2 py-1 rounded-md disabled:opacity-60"
                        onClick={() => void handleGenerateImage(record)}
                        disabled={generatingChallengeId !== null}
                      >
                        {generatingChallengeId === record.id
                          ? 'Generating...'
                          : 'Generate Image'}
                      </button>
                      <button
                        type="button"
                        className="bg-blue-600 text-white px-2 py-1 rounded-md"
                        onClick={() => openEdit(record)}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        className="bg-red-600 text-white px-2 py-1 rounded-md"
                        onClick={() => void deleteChallenge(record)}
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-lg p-6 w-full max-w-4xl max-h-[90vh] overflow-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">
                {editingChallenge ? 'Edit Challenge' : 'Create Challenge'}
              </h2>
              <button
                type="button"
                className="text-gray-600 hover:text-black"
                onClick={closeModal}
              >
                Close
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <label className="text-sm">
                Zone
                <select
                  className="w-full border rounded-md p-2"
                  value={form.zoneId}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      zoneId: event.target.value,
                    }))
                  }
                >
                  <option value="">Select zone</option>
                  {zones.map((zone) => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="text-sm">
                Submission Type
                <select
                  className="w-full border rounded-md p-2"
                  value={form.submissionType}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      submissionType: event.target.value as 'photo' | 'text' | 'video',
                    }))
                  }
                >
                  <option value="photo">photo</option>
                  <option value="text">text</option>
                  <option value="video">video</option>
                </select>
              </label>

              <label className="text-sm md:col-span-2">
                Question
                <textarea
                  className="w-full border rounded-md p-2 min-h-[100px]"
                  value={form.question}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, question: event.target.value }))
                  }
                />
              </label>

              <label className="text-sm">
                Difficulty
                <input
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                  value={form.difficulty}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, difficulty: event.target.value }))
                  }
                />
              </label>

              <label className="text-sm">
                Reward (Gold/Score)
                <input
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                  value={form.reward}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, reward: event.target.value }))
                  }
                />
              </label>

              <label className="text-sm">
                Reward Item (Optional)
                <select
                  className="w-full border rounded-md p-2"
                  value={form.inventoryItemId}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      inventoryItemId: event.target.value,
                    }))
                  }
                >
                  <option value="">None</option>
                  {inventoryItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name} (#{item.id})
                    </option>
                  ))}
                </select>
              </label>

              <label className="text-sm">
                Proficiency (Optional)
                <input
                  className="w-full border rounded-md p-2"
                  value={form.proficiency}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, proficiency: event.target.value }))
                  }
                />
              </label>

              <label className="text-sm md:col-span-2">
                Stat Tags (comma separated)
                <input
                  className="w-full border rounded-md p-2"
                  placeholder="strength, dexterity"
                  value={form.statTags}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, statTags: event.target.value }))
                  }
                />
              </label>

              <label className="text-sm">
                Latitude
                <input
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                  value={form.latitude}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, latitude: event.target.value }))
                  }
                />
              </label>

              <label className="text-sm">
                Longitude
                <input
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                  value={form.longitude}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, longitude: event.target.value }))
                  }
                />
              </label>
            </div>

            <div className="mt-3 mb-3">
              <button
                type="button"
                className="bg-gray-700 text-white px-3 py-2 rounded-md disabled:opacity-60"
                onClick={handleUseCurrentLocation}
                disabled={geoLoading}
              >
                {geoLoading ? 'Locating...' : 'Use Current Browser Location'}
              </button>
            </div>

            <div className="mb-4">
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm">Map Location Picker</span>
                <span className="text-xs text-gray-500">
                  Click map to set latitude/longitude
                </span>
              </div>
              {mapboxgl.accessToken ? (
                <div ref={mapContainerRef} className="w-full h-64 border rounded-md" />
              ) : (
                <div className="w-full border rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                  Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                  unavailable.
                </div>
              )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-5">
              <label className="text-sm">
                Image URL (Optional)
                <input
                  className="w-full border rounded-md p-2"
                  value={form.imageUrl}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, imageUrl: event.target.value }))
                  }
                />
              </label>
              <label className="text-sm">
                Thumbnail URL (Optional)
                <input
                  className="w-full border rounded-md p-2"
                  value={form.thumbnailUrl}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      thumbnailUrl: event.target.value,
                    }))
                  }
                />
              </label>
            </div>

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="bg-gray-200 px-4 py-2 rounded-md"
                onClick={closeModal}
              >
                Cancel
              </button>
              <button
                type="button"
                className="bg-blue-600 text-white px-4 py-2 rounded-md"
                onClick={() => void save()}
              >
                Save Challenge
              </button>
            </div>
          </div>
        </div>
      )}

      {imagePreview && (
        <div className="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-6">
          <div className="relative max-w-4xl w-full">
            <button
              type="button"
              className="absolute -top-10 right-0 bg-white text-black px-3 py-1 rounded-md"
              onClick={closeImagePreview}
            >
              Close
            </button>
            <img
              src={imagePreview.url}
              alt={imagePreview.alt}
              className="w-full max-h-[80vh] object-contain rounded-lg bg-black"
            />
          </div>
        </div>
      )}
    </div>
  );
};
