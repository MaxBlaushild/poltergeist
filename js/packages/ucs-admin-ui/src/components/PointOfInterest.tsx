import React, { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI, useMediaContext, useTagContext, useZoneContext } from '@poltergeist/contexts';
import { useCandidates, usePointOfInterestGroups } from '@poltergeist/hooks';
import { Candidate, PointOfInterest as PointOfInterestType, Tag } from '@poltergeist/types';

const flattenTags = (tagGroups: { tags: Tag[] }[]): Tag[] => {
  return tagGroups.flatMap(group => group.tags);
};

export const PointOfInterest = () => {
  const { apiClient } = useAPI();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { zones } = useZoneContext();
  const { tagGroups } = useTagContext();
  const { pointOfInterestGroups } = usePointOfInterestGroups();

  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterestType[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [nameQuery, setNameQuery] = useState('');
  const [selectedZoneId, setSelectedZoneId] = useState('');
  const [selectedTagIds, setSelectedTagIds] = useState<Set<string>>(new Set());
  const [filtersOpen, setFiltersOpen] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [importError, setImportError] = useState<string | null>(null);
  const [createForm, setCreateForm] = useState({
    name: '',
    description: '',
    clue: '',
    lat: '',
    lng: '',
    imageUrl: '',
    unlockTier: '',
    groupId: '',
  });
  const [createImageFile, setCreateImageFile] = useState<File | null>(null);
  const [createImagePreview, setCreateImagePreview] = useState<string | null>(null);

  const [importQuery, setImportQuery] = useState('');
  const [selectedCandidate, setSelectedCandidate] = useState<Candidate | null>(null);
  const [importZoneId, setImportZoneId] = useState('');
  const { candidates } = useCandidates(importQuery);

  const allTags = useMemo(() => flattenTags(tagGroups), [tagGroups]);
  const fetchPointsOfInterest = async () => {
    setLoading(true);
    setError(null);
    try {
      const endpoint = selectedZoneId
        ? `/sonar/zones/${selectedZoneId}/pointsOfInterest`
        : '/sonar/pointsOfInterest';
      const response = await apiClient.get<PointOfInterestType[]>(endpoint);
      setPointsOfInterest(response);
    } catch (err) {
      console.error('Error fetching points of interest:', err);
      setError('Failed to load points of interest');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPointsOfInterest();
  }, [apiClient, selectedZoneId]);

  const toggleTag = (tagId: string) => {
    setSelectedTagIds(prev => {
      const next = new Set(prev);
      if (next.has(tagId)) {
        next.delete(tagId);
      } else {
        next.add(tagId);
      }
      return next;
    });
  };

  const clearTags = () => setSelectedTagIds(new Set());

  const filteredPoints = useMemo(() => {
    const query = nameQuery.trim().toLowerCase();
    const selectedTags = Array.from(selectedTagIds);

    return pointsOfInterest.filter(point => {
      const matchesName = query.length === 0
        || point.name.toLowerCase().includes(query)
        || (point.originalName && point.originalName.toLowerCase().includes(query));

      if (!matchesName) {
        return false;
      }

      if (selectedTags.length === 0) {
        return true;
      }

      const pointTagIds = point.tags?.map(tag => tag.id) ?? [];
      return selectedTags.every(tagId => pointTagIds.includes(tagId));
    });
  }, [pointsOfInterest, nameQuery, selectedTagIds]);

  const resetCreateForm = () => {
    setCreateForm({
      name: '',
      description: '',
      clue: '',
      lat: '',
      lng: '',
      imageUrl: '',
      unlockTier: '',
      groupId: '',
    });
    setCreateImageFile(null);
    setCreateImagePreview(null);
    setCreateError(null);
  };

  const resetImportForm = () => {
    setImportQuery('');
    setSelectedCandidate(null);
    setImportZoneId('');
    setImportError(null);
  };

  const handleCreateImageChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null;
    setCreateImageFile(file);
    if (!file) {
      setCreateImagePreview(null);
      return;
    }
    const reader = new FileReader();
    reader.onloadend = () => {
      setCreateImagePreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const uploadCreateImage = async () => {
    if (!createImageFile) return null;
    const imageKey = `points-of-interest/${createImageFile.name.toLowerCase().replace(/\s+/g, '-')}`;
    const presignedUrl = await getPresignedUploadURL('crew-points-of-interest', imageKey);
    if (!presignedUrl) {
      throw new Error('Failed to get upload URL');
    }
    await uploadMedia(presignedUrl, createImageFile);
    return presignedUrl.split('?')[0];
  };

  const handleCreatePointOfInterest = async () => {
    setCreateError(null);
    if (!createForm.groupId) {
      setCreateError('Please select a point of interest group.');
      return;
    }
    if (!createForm.name || !createForm.description || !createForm.clue) {
      setCreateError('Name, description, and clue are required.');
      return;
    }
    if (!createForm.lat || !createForm.lng) {
      setCreateError('Latitude and longitude are required.');
      return;
    }

    try {
      let imageUrl = createForm.imageUrl;
      if (createImageFile) {
        const uploadedUrl = await uploadCreateImage();
        if (uploadedUrl) {
          imageUrl = uploadedUrl;
        }
      }

      if (!imageUrl) {
        setCreateError('Please provide an image URL or upload an image.');
        return;
      }

      await apiClient.post(`/sonar/pointsOfInterest/group/${createForm.groupId}`, {
        name: createForm.name,
        description: createForm.description,
        latitude: createForm.lat,
        longitude: createForm.lng,
        imageUrl,
        clue: createForm.clue,
        unlockTier: createForm.unlockTier ? Number(createForm.unlockTier) : null,
      });

      setShowCreateModal(false);
      resetCreateForm();
      fetchPointsOfInterest();
    } catch (err) {
      console.error('Error creating point of interest:', err);
      setCreateError('Failed to create point of interest.');
    }
  };

  const handleImportPointOfInterest = async () => {
    setImportError(null);
    if (!selectedCandidate) {
      setImportError('Please select a Google Maps location.');
      return;
    }
    if (!importZoneId) {
      setImportError('Please select a zone.');
      return;
    }
    try {
      await apiClient.post('/sonar/pointOfInterest/import', {
        placeID: selectedCandidate.place_id,
        zoneID: importZoneId,
      });
      setShowImportModal(false);
      resetImportForm();
      fetchPointsOfInterest();
    } catch (err) {
      console.error('Error importing point of interest:', err);
      setImportError('Failed to import point of interest.');
    }
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex flex-col gap-2 mb-6">
        <h1 className="text-3xl font-bold">Points of Interest</h1>
        <p className="text-gray-600">
          Search by name and filter by zone or tags.
        </p>
      </div>

      <div className="flex flex-wrap gap-3 mb-6">
        <button
          type="button"
          className="bg-blue-600 text-white px-4 py-2 rounded-md"
          onClick={() => {
            resetCreateForm();
            setShowCreateModal(true);
          }}
        >
          Create Point of Interest
        </button>
        <button
          type="button"
          className="bg-green-600 text-white px-4 py-2 rounded-md"
          onClick={() => {
            resetImportForm();
            setShowImportModal(true);
          }}
        >
          Import from Google Maps
        </button>
      </div>

      <div className="bg-white rounded-lg shadow-md p-4 mb-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">Filters</h2>
          <button
            type="button"
            onClick={() => setFiltersOpen((prev) => !prev)}
            className="text-sm text-blue-600 hover:underline"
          >
            {filtersOpen ? 'Hide filters' : 'Show filters'}
          </button>
        </div>

        {filtersOpen && (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Search</label>
                <input
                  type="text"
                  placeholder="Search by name..."
                  value={nameQuery}
                  onChange={(e) => setNameQuery(e.target.value)}
                  className="border border-gray-300 rounded-md px-3 py-2"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Zone</label>
                <select
                  className="border border-gray-300 rounded-md px-3 py-2"
                  value={selectedZoneId}
                  onChange={(e) => setSelectedZoneId(e.target.value)}
                >
                  <option value="">All zones</option>
                  {zones.map(zone => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Tags</label>
                <div className="flex flex-wrap gap-2">
                  {allTags.map(tag => {
                    const isSelected = selectedTagIds.has(tag.id);
                    return (
                      <button
                        key={tag.id}
                        type="button"
                        onClick={() => toggleTag(tag.id)}
                        className={`px-3 py-1 rounded-full text-sm border transition-colors ${
                          isSelected
                            ? 'bg-blue-600 text-white border-blue-600'
                            : 'bg-gray-100 text-gray-700 border-gray-200 hover:bg-gray-200'
                        }`}
                      >
                        {tag.name}
                      </button>
                    );
                  })}
                  {allTags.length === 0 && (
                    <span className="text-sm text-gray-500">No tags available</span>
                  )}
                </div>
                {selectedTagIds.size > 0 && (
                  <button
                    type="button"
                    onClick={clearTags}
                    className="text-sm text-blue-600 hover:underline w-fit"
                  >
                    Clear tag filters
                  </button>
                )}
              </div>
            </div>

            <div className="mt-4 text-sm text-gray-600">
              Showing {filteredPoints.length} of {pointsOfInterest.length} points
              {selectedTagIds.size > 0 && ' (matches all selected tags)'}
            </div>
          </>
        )}
      </div>

      {loading && (
        <div className="text-gray-600">Loading points of interest...</div>
      )}
      {error && (
        <div className="text-red-600">{error}</div>
      )}

      {!loading && !error && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredPoints.map(point => {
            const card = (
              <div className="bg-white rounded-lg shadow-md p-4 h-full">
                {point.imageURL && (
                  <img
                    src={point.imageURL}
                    alt={point.name}
                    className="w-full h-40 object-cover rounded mb-3"
                  />
                )}
                <h2 className="text-xl font-semibold mb-1">{point.name}</h2>
                {point.originalName && (
                  <p className="text-sm text-gray-500 mb-2">{point.originalName}</p>
                )}
                <p className="text-sm text-gray-700 mb-2">{point.description}</p>
                <div className="text-sm text-gray-600 space-y-1">
                  <div>Latitude: {point.lat}</div>
                  <div>Longitude: {point.lng}</div>
                  <div>Clue: {point.clue || 'None'}</div>
                  <div>
                    Tags: {point.tags?.length ? point.tags.map(tag => tag.name).join(', ') : 'None'}
                  </div>
                </div>
              </div>
            );

            return (
              <Link key={point.id} to={`/points-of-interest/${point.id}`} className="block hover:shadow-lg transition-shadow">
                {card}
              </Link>
            );
          })}

          {filteredPoints.length === 0 && (
            <div className="col-span-full text-gray-600">
              No points of interest match these filters.
            </div>
          )}
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">Create Point of Interest</h2>

            {createError && (
              <div className="mb-4 text-red-600 text-sm">{createError}</div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Group</label>
                <select
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.groupId}
                  onChange={(e) => setCreateForm({ ...createForm, groupId: e.target.value })}
                >
                  <option value="">Select a group</option>
                  {(pointOfInterestGroups ?? []).map(group => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.name}
                  onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Latitude</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.lat}
                  onChange={(e) => setCreateForm({ ...createForm, lat: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Longitude</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.lng}
                  onChange={(e) => setCreateForm({ ...createForm, lng: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Unlock Tier</label>
                <input
                  type="number"
                  min="1"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.unlockTier}
                  onChange={(e) => setCreateForm({ ...createForm, unlockTier: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Image URL</label>
                <input
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={createForm.imageUrl}
                  onChange={(e) => setCreateForm({ ...createForm, imageUrl: e.target.value })}
                />
              </div>
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                rows={3}
                value={createForm.description}
                onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
              />
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Clue</label>
              <textarea
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                rows={2}
                value={createForm.clue}
                onChange={(e) => setCreateForm({ ...createForm, clue: e.target.value })}
              />
            </div>

            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Upload Image</label>
              <input type="file" accept="image/*" onChange={handleCreateImageChange} />
              {createImagePreview && (
                <img src={createImagePreview} alt="Preview" className="mt-2 h-32 w-full object-cover rounded" />
              )}
            </div>

            <div className="flex justify-end gap-2 mt-6">
              <button
                type="button"
                className="px-4 py-2 rounded-md border border-gray-300"
                onClick={() => {
                  setShowCreateModal(false);
                  resetCreateForm();
                }}
              >
                Cancel
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-md bg-blue-600 text-white"
                onClick={handleCreatePointOfInterest}
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {showImportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">Import Point of Interest</h2>

            {importError && (
              <div className="mb-4 text-red-600 text-sm">{importError}</div>
            )}

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Zone</label>
              <select
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                value={importZoneId}
                onChange={(e) => setImportZoneId(e.target.value)}
              >
                <option value="">Select a zone</option>
                {zones.map(zone => (
                  <option key={zone.id} value={zone.id}>
                    {zone.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Search Google Maps</label>
              <input
                type="text"
                className="w-full border border-gray-300 rounded-md px-3 py-2"
                value={importQuery}
                onChange={(e) => setImportQuery(e.target.value)}
                placeholder="Search for a place..."
              />
            </div>

            <div className="border border-gray-200 rounded-md max-h-64 overflow-y-auto">
              {candidates.length === 0 && (
                <div className="p-4 text-sm text-gray-500">No results yet.</div>
              )}
              {candidates.map(candidate => (
                <button
                  key={candidate.place_id}
                  type="button"
                  className={`w-full text-left px-4 py-3 border-b border-gray-100 hover:bg-gray-50 ${
                    selectedCandidate?.place_id === candidate.place_id ? 'bg-blue-50' : ''
                  }`}
                  onClick={() => setSelectedCandidate(candidate)}
                >
                  <div className="font-medium">{candidate.name}</div>
                  <div className="text-xs text-gray-500">{candidate.formatted_address}</div>
                </button>
              ))}
            </div>

            <div className="flex justify-end gap-2 mt-6">
              <button
                type="button"
                className="px-4 py-2 rounded-md border border-gray-300"
                onClick={() => {
                  setShowImportModal(false);
                  resetImportForm();
                }}
              >
                Cancel
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-md bg-green-600 text-white"
                onClick={handleImportPointOfInterest}
              >
                Import
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
