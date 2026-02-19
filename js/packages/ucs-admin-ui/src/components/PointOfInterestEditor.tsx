import React, { useEffect, useMemo, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useAPI, useMediaContext, useZoneContext } from '@poltergeist/contexts';
import { Character, PointOfInterest } from '@poltergeist/types';

const buildCharacterPayload = (character: Character, pointOfInterestId: string | null) => {
  return {
    name: character.name,
    description: character.description,
    mapIconUrl: character.mapIconUrl,
    dialogueImageUrl: character.dialogueImageUrl,
    pointOfInterestId,
    movementPattern: {
      movementPatternType: character.movementPattern.movementPatternType,
      zoneId: character.movementPattern.zoneId ?? null,
      startingLatitude: character.movementPattern.startingLatitude,
      startingLongitude: character.movementPattern.startingLongitude,
      path: character.movementPattern.path ?? [],
    },
  };
};

export const PointOfInterestEditor = () => {
  const { id } = useParams();
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();

  const [pointOfInterest, setPointOfInterest] = useState<PointOfInterest | null>(null);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [selectedZoneId, setSelectedZoneId] = useState<string>('');
  const [selectedCharacterId, setSelectedCharacterId] = useState<string>('');
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [refreshingImage, setRefreshingImage] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isPreviewOpen, setIsPreviewOpen] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    clue: '',
    lat: '',
    lng: '',
    imageURL: '',
    originalName: '',
    googleMapsPlaceId: '',
    unlockTier: '' as string,
  });

  const formatGenerationStatus = (status?: string) => {
    switch (status) {
      case 'queued':
        return 'Queued';
      case 'in_progress':
        return 'Generating';
      case 'complete':
        return 'Complete';
      case 'failed':
        return 'Failed';
      case 'none':
        return 'Not requested';
      default:
        return 'Unknown';
    }
  };

  const assignedCharacters = useMemo(() => {
    if (!pointOfInterest) return [];
    return characters.filter(character => character.pointOfInterestId === pointOfInterest.id);
  }, [characters, pointOfInterest]);

  useEffect(() => {
    if (!id) {
      setError('Point of interest ID is missing.');
      setLoading(false);
      return;
    }

    const fetchPointOfInterest = async () => {
      try {
        const points = await apiClient.get<PointOfInterest[]>('/sonar/pointsOfInterest');
        const selected = points.find(point => point.id === id) || null;
        if (!selected) {
          setError('Point of interest not found.');
          setLoading(false);
          return;
        }
        setPointOfInterest(selected);
        setFormData({
          name: selected.name ?? '',
          description: selected.description ?? '',
          clue: selected.clue ?? '',
          lat: selected.lat ?? '',
          lng: selected.lng ?? '',
          imageURL: selected.imageURL ?? '',
          originalName: selected.originalName ?? '',
          googleMapsPlaceId: selected.googleMapsPlaceId ?? '',
          unlockTier: selected.unlockTier != null ? String(selected.unlockTier) : '',
        });
      } catch (err) {
        console.error('Error loading point of interest:', err);
        setError('Failed to load point of interest.');
      } finally {
        setLoading(false);
      }
    };

    const fetchCharacters = async () => {
      try {
        const response = await apiClient.get<Character[]>('/sonar/characters');
        setCharacters(response);
      } catch (err) {
        console.error('Error fetching characters:', err);
      }
    };

    const fetchZoneForPoint = async () => {
      try {
        const zone = await apiClient.get<{ id: string }>(`/sonar/pointOfInterest/${id}/zone`);
        if (zone?.id) {
          setSelectedZoneId(zone.id);
        }
      } catch (err) {
        // POI might not be in any zone.
      }
    };

    fetchPointOfInterest();
    fetchCharacters();
    fetchZoneForPoint();
  }, [apiClient, id]);

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const handleImageFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] || null;
    setImageFile(file);
    if (!file) {
      setImagePreview(null);
      return;
    }
    const reader = new FileReader();
    reader.onloadend = () => {
      setImagePreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const previewSrc = imagePreview || formData.imageURL || null;

  const uploadImageIfNeeded = async () => {
    if (!imageFile || !id) return null;
    const imageKey = `points-of-interest/${id}-${imageFile.name.toLowerCase().replace(/\s+/g, '-')}`;
    const presignedUrl = await getPresignedUploadURL('crew-points-of-interest', imageKey);
    if (!presignedUrl) {
      throw new Error('Failed to get upload URL');
    }
    await uploadMedia(presignedUrl, imageFile);
    return presignedUrl.split('?')[0];
  };

  const handleSave = async () => {
    if (!id) return;
    setSaving(true);
    setError(null);

    try {
      let imageUrl = formData.imageURL;
      if (imageFile) {
        const uploadedUrl = await uploadImageIfNeeded();
        if (uploadedUrl) {
          imageUrl = uploadedUrl;
        }
      }

      await apiClient.patch(`/sonar/pointsOfInterest/${id}`, {
        name: formData.name,
        description: formData.description,
        clue: formData.clue,
        lat: formData.lat,
        lng: formData.lng,
        imageUrl,
        originalName: formData.originalName,
        googleMapsPlaceId: formData.googleMapsPlaceId,
        unlockTier: formData.unlockTier === '' ? null : Number(formData.unlockTier),
      });

      setFormData(prev => ({ ...prev, imageURL: imageUrl }));
      setImageFile(null);
      setImagePreview(null);
    } catch (err) {
      console.error('Error saving point of interest:', err);
      setError('Failed to save changes.');
    } finally {
      setSaving(false);
    }
  };

  const handleRefreshImage = async () => {
    if (!id) return;
    setRefreshingImage(true);
    setError(null);
    try {
      const updated = await apiClient.post<PointOfInterest>('/sonar/pointOfInterest/image/refresh', {
        pointOfInterestID: id,
      });
      setPointOfInterest(updated);
      setFormData(prev => ({ ...prev, imageURL: updated.imageURL ?? '' }));
      setImageFile(null);
      setImagePreview(updated.imageURL ?? null);
    } catch (err) {
      console.error('Error refreshing point of interest image:', err);
      setError('Failed to refresh point of interest image.');
    } finally {
      setRefreshingImage(false);
    }
  };

  const handleZoneChange = async (nextZoneId: string) => {
    if (!id) return;
    setSelectedZoneId(nextZoneId);
    try {
      if (selectedZoneId) {
        await apiClient.delete(`/sonar/zones/${selectedZoneId}/pointOfInterest/${id}`);
      }
      if (nextZoneId) {
        await apiClient.post(`/sonar/zones/${nextZoneId}/pointOfInterest/${id}`);
      }
    } catch (err) {
      console.error('Error updating zone assignment:', err);
      setError('Failed to update zone assignment.');
    }
  };

  const assignCharacter = async () => {
    if (!selectedCharacterId || !pointOfInterest) return;
    const character = characters.find(c => c.id === selectedCharacterId);
    if (!character) return;

    try {
      const payload = buildCharacterPayload(character, pointOfInterest.id);
      const updatedCharacter = await apiClient.put<Character>(`/sonar/characters/${character.id}`, payload);
      setCharacters(prev => prev.map(c => c.id === updatedCharacter.id ? updatedCharacter : c));
      setSelectedCharacterId('');
    } catch (err) {
      console.error('Error assigning character:', err);
      setError('Failed to assign character.');
    }
  };

  const removeCharacter = async (character: Character) => {
    if (!pointOfInterest) return;
    try {
      const payload = buildCharacterPayload(character, null);
      const updatedCharacter = await apiClient.put<Character>(`/sonar/characters/${character.id}`, payload);
      setCharacters(prev => prev.map(c => c.id === updatedCharacter.id ? updatedCharacter : c));
    } catch (err) {
      console.error('Error removing character:', err);
      setError('Failed to remove character.');
    }
  };

  if (loading) {
    return <div className="p-6">Loading point of interest...</div>;
  }

  if (error) {
    return <div className="p-6 text-red-600">{error}</div>;
  }

  if (!pointOfInterest) {
    return <div className="p-6">Point of interest not found.</div>;
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold">Edit Point of Interest</h1>
          <p className="text-gray-600">{pointOfInterest.name}</p>
        </div>
        <Link to="/points-of-interest" className="text-blue-600 hover:underline">
          Back to list
        </Link>
      </div>

      <div className="bg-white rounded-lg shadow-md p-6 space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Original Name</label>
            <input
              type="text"
              value={formData.originalName}
              onChange={(e) => handleInputChange('originalName', e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Latitude</label>
            <input
              type="text"
              value={formData.lat}
              onChange={(e) => handleInputChange('lat', e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Longitude</label>
            <input
              type="text"
              value={formData.lng}
              onChange={(e) => handleInputChange('lng', e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Unlock Tier</label>
            <input
              type="number"
              value={formData.unlockTier}
              onChange={(e) => handleInputChange('unlockTier', e.target.value)}
              min="1"
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Google Maps Place ID</label>
            <input
              type="text"
              value={formData.googleMapsPlaceId}
              onChange={(e) => handleInputChange('googleMapsPlaceId', e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea
            value={formData.description}
            onChange={(e) => handleInputChange('description', e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2"
            rows={3}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Clue</label>
          <textarea
            value={formData.clue}
            onChange={(e) => handleInputChange('clue', e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2"
            rows={2}
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Image URL</label>
            <input
              type="text"
              value={formData.imageURL}
              onChange={(e) => handleInputChange('imageURL', e.target.value)}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Upload New Image</label>
            <input
              type="file"
              accept="image/*"
              onChange={handleImageFileChange}
              className="w-full"
            />
            {(imagePreview || formData.imageURL) && (
              <img
                src={previewSrc ?? undefined}
                alt="Preview"
                className="mt-2 h-32 w-full object-cover rounded cursor-pointer"
                onClick={() => setIsPreviewOpen(true)}
              />
            )}
            <p className="mt-2 text-sm text-gray-600">
              Image Status: {formatGenerationStatus(pointOfInterest.imageGenerationStatus)}
            </p>
            {pointOfInterest.imageGenerationStatus === 'failed' && pointOfInterest.imageGenerationError && (
              <p className="mt-1 text-xs text-red-600">
                Error: {pointOfInterest.imageGenerationError}
              </p>
            )}
            <div className="mt-2">
              <button
                type="button"
                onClick={handleRefreshImage}
                disabled={refreshingImage || ['queued', 'in_progress'].includes(pointOfInterest.imageGenerationStatus || '')}
                className="px-3 py-2 bg-yellow-500 text-white rounded-md disabled:bg-gray-300"
              >
                {refreshingImage ? 'Refreshing Image...' : 'Refresh Image'}
              </button>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Zone</label>
            <select
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              value={selectedZoneId}
              onChange={(e) => handleZoneChange(e.target.value)}
            >
              <option value="">No zone</option>
              {zones.map(zone => (
                <option key={zone.id} value={zone.id}>
                  {zone.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Assign Character</label>
            <div className="flex gap-2">
              <select
                className="flex-1 border border-gray-300 rounded-md px-3 py-2"
                value={selectedCharacterId}
                onChange={(e) => setSelectedCharacterId(e.target.value)}
              >
                <option value="">Select character</option>
                {characters
                  .filter(character => character.pointOfInterestId !== pointOfInterest.id)
                  .map(character => (
                    <option key={character.id} value={character.id}>
                      {character.name}
                    </option>
                  ))}
              </select>
              <button
                type="button"
                onClick={assignCharacter}
                disabled={!selectedCharacterId}
                className="px-4 py-2 bg-blue-600 text-white rounded-md disabled:bg-gray-300"
              >
                Assign
              </button>
            </div>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Assigned Characters</label>
          <div className="space-y-2">
            {assignedCharacters.length === 0 && (
              <div className="text-sm text-gray-500">No characters assigned.</div>
            )}
            {assignedCharacters.map(character => (
              <div key={character.id} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                <span>{character.name}</span>
                <button
                  type="button"
                  onClick={() => removeCharacter(character)}
                  className="text-sm text-red-600 hover:underline"
                >
                  Remove
                </button>
              </div>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 bg-green-600 text-white rounded-md disabled:bg-gray-300"
          >
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>
      {isPreviewOpen && previewSrc && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-4"
          onClick={() => setIsPreviewOpen(false)}
        >
          <div
            className="relative max-h-[90vh] max-w-4xl w-full"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => setIsPreviewOpen(false)}
              className="absolute -top-10 right-0 text-white text-sm underline"
            >
              Close
            </button>
            <img
              src={previewSrc}
              alt="Point of interest preview"
              className="max-h-[90vh] w-full rounded-lg object-contain bg-black"
            />
          </div>
        </div>
      )}
    </div>
  );
};
