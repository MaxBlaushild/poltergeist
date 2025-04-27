import React, { useState } from 'react';
import { useZoneContext } from '../contexts/zones.tsx';
import { v4 as uuidv4 } from 'uuid';
import { useParams } from 'react-router-dom';
import { useZonePointsOfInterest } from '../hooks/useZonePointsOfInterest.ts';
import { usePlaceTypes } from '../hooks/usePlaceTypes.ts';
import { useGeneratePointsOfInterest } from '../hooks/useGeneratePointsOfInterest.ts';
import { useCandidates } from '@poltergeist/hooks';
import { Candidate } from '@poltergeist/types';
import { useQuestArchtypes } from '../hooks/useQuestArchtypes.ts';
import { useAPI } from '@poltergeist/contexts';
export const Zone = () => {
  const { id } = useParams();
  const { apiClient } = useAPI();
  const { zones, selectedZone, setSelectedZone, createZone, deleteZone } =
    useZoneContext();
  const zone = zones.find((zone) => zone.id === id);
  const { pointsOfInterest, loading, error } = useZonePointsOfInterest(id!);
  const {
    placeTypes,
    loading: placeTypesLoading,
    error: placeTypesError,
  } = usePlaceTypes();
  const [isGenerating, setIsGenerating] = useState(false);
  const [selectedIncludedPlaceTypes, setSelectedIncludedPlaceTypes] = useState<
    string[]
  >([]);
  const [selectedExcludedPlaceTypes, setSelectedExcludedPlaceTypes] = useState<
    string[]
  >([]);
  const {
    questArchtypes,
    loading: questArchtypesLoading,
    error: questArchtypesError,
  } = useQuestArchtypes();
  const [numPlaces, setNumPlaces] = useState(1);
  const [address, setAddress] = useState('');
  const [showPlaces, setShowPlaces] = useState(false);
  const timeoutRef = React.useRef<number>();
  const [query, setQuery] = useState('');
  const [shouldShowImage, setShouldShowImage] = useState(false);
  const [isImporting, setIsImporting] = useState(false);
  const [importedPlaces, setImportedPlaces] = useState<string[]>([]);
  const [isGeneratingQuest, setIsGeneratingQuest] = useState(false);
  const [selectedQuestArchtype, setSelectedQuestArchtype] = useState<
    string | null
  >(null);
  const [nameFilter, setNameFilter] = useState('');
  const {
    candidates,
    loading: candidatesLoading,
    error: candidatesError,
  } = useCandidates(query);

  const {
    loading: generatePointsOfInterestLoading,
    error: generatePointsOfInterestError,
    generatePointsOfInterest,
    refreshPointOfInterestImage,
    refreshPointOfInterest,
    importPointOfInterest,
    generateQuest,
  } = useGeneratePointsOfInterest(id!);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);

  const handleQueryChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      setQuery(e.target.value);
      setShowPlaces(true);
    }, 500);
  };

  const handleCandidateSelect = (candidate: Candidate) => {
    importPointOfInterest(candidate.place_id, id!);
  };

  const handleGenerateQuest = async () => {
    // const/sonar/zones/:id/questArchetypes
    await apiClient.post(`/sonar/zones/${id}/questArchetypes`, {});
  };

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>Error: {error.message}</div>;
  }
  if (!zone) {
    return <div>Zone not found</div>;
  }

  if (generatePointsOfInterestLoading) {
    return <div>Generating points of interest...</div>;
  }
  if (generatePointsOfInterestError) {
    return <div>Error: {generatePointsOfInterestError.message}</div>;
  }

  const filteredPoints = pointsOfInterest.filter(point => 
    point.name.toLowerCase().includes(nameFilter.toLowerCase())
  );

  return (
    <div className="m-10 p-8 bg-white rounded-lg shadow-lg">
      <h1 className="text-3xl font-bold mb-6 text-gray-800">{zone?.name}</h1>
      <p className="text-lg text-gray-600 mb-3">Latitude: {zone?.latitude}</p>
      <p className="text-lg text-gray-600 mb-3">Longitude: {zone?.longitude}</p>
      <p className="text-lg text-gray-600 mb-3">Radius: {zone?.radius}m</p>
      
      <div className="mb-4">
        <input
          type="text"
          placeholder="Filter points by name..."
          value={nameFilter}
          onChange={(e) => setNameFilter(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredPoints.map((point) => (
          <div key={point.id} className="bg-gray-100 p-6 rounded-lg shadow-md">
            {point.imageURL && (
              <>
                <div className="relative">
                  <img
                    src={point.imageURL}
                    alt={point.name}
                    className="float-left mr-4 mb-4 w-32 h-32 object-cover rounded cursor-pointer"
                    onClick={() => setSelectedImage(point.imageURL)}
                  />
                </div>
                {selectedImage === point.imageURL && (
                  <div
                    className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
                    onClick={() => setSelectedImage(null)}
                  >
                    <div className="relative">
                      <img
                        src={selectedImage}
                        alt={point.name}
                        className="max-h-[90vh] max-w-[90vw] object-contain"
                      />
                      <button
                        className="absolute top-4 right-4 text-white text-xl font-bold"
                        onClick={() => setSelectedImage(null)}
                      >
                        ✕
                      </button>
                    </div>
                  </div>
                )}
              </>
            )}
            <h2 className="text-xl font-bold mb-2 text-gray-800">
              {point.name}
            </h2>
            <p className="text-gray-600 mb-3">
              Description: {point.description}
            </p>
            <p className="text-gray-600 mb-3">
              Type: {point.originalName}
            </p>
            <p className="text-gray-600 mb-3">
              Tags: {point.tags?.map(tag => tag.name).join(', ') || 'No tags'}
            </p>
            <p className="text-gray-600 mb-3">Latitude: {point.lat}</p>
            <p className="text-gray-600 mb-3">Longitude: {point.lng}</p>
            <p className="text-gray-600 mb-3">Clue: {point.clue}</p>
            <p className="text-gray-600 mb-3">
              Created At: {point.createdAt.toLocaleString()}
            </p>
            <p className="text-gray-600 mb-3">
              Updated At: {point.updatedAt.toLocaleString()}
            </p>
            <p className="text-gray-600 mb-3">
              Place ID:{' '}
              <a
                href={`/place/${point.googleMapsPlaceId}`}
                className="text-blue-500 hover:underline"
              >
                {point.googleMapsPlaceId}
              </a>
            </p>
            <button
              className="bg-blue-500 hover:bg-blue-600 text-white rounded-md px-3 py-1 text-sm font-medium shadow-md mr-2 transition duration-200"
              onClick={(e) => {
                e.stopPropagation();
                refreshPointOfInterestImage(point.id);
              }}
            >
              Refresh Image
            </button>
            <button
              className="bg-green-500 hover:bg-green-600 text-white rounded-md px-3 py-1 text-sm font-medium shadow-md transition duration-200"
              onClick={(e) => {
                e.stopPropagation();
                refreshPointOfInterest(point.id);
              }}
            >
              Refresh POI
            </button>
          </div>
        ))}
      </div>
      <button
        onClick={() => setIsGenerating(!isGenerating)}
        className="bg-blue-500 text-white px-4 py-2 rounded-md"
      >
        {isGenerating ? 'Stop Generating' : 'Generate Points of Interest'}
      </button>
      <button
        onClick={() => setIsImporting(!isImporting)}
        className="bg-blue-500 text-white px-4 py-2 rounded-md"
      >
        Import Point of Interest
      </button>
      <button
        onClick={() => setIsGeneratingQuest(!isGeneratingQuest)}
        className="bg-blue-500 text-white px-4 py-2 rounded-md"
      >
        Generate Quest
      </button>
      <button
        onClick={handleGenerateQuest}
        className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600 disabled:bg-blue-300"
      >
        Generate Quests for Zone
      </button>
      {isImporting && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg shadow-xl w-96">
            <h2 className="text-xl font-bold mb-4">Import Point of Interest</h2>
            <div className="space-y-4">
              <div>
                <label
                  htmlFor="importText"
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
                  Enter Import Text
                </label>
                <input
                  id="importText"
                  type="text"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  onChange={handleQueryChange}
                />
                {showPlaces && candidates && candidates.length > 0 && (
                  <div className="w-full bg-white border rounded-lg shadow-lg max-h-48 overflow-y-auto mt-1">
                    {candidates.map((candidate, index) => (
                      <div
                        key={index}
                        className="px-4 py-2 hover:bg-gray-100 cursor-pointer"
                        onClick={() => handleCandidateSelect(candidate)}
                      >
                        {candidate.name}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {isGeneratingQuest && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg shadow-xl w-96">
            <h2 className="text-xl font-bold mb-4">Generate Quest</h2>
            <div className="space-y-4">
              <div>
                <label
                  htmlFor="questType"
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
                  Quest Archetype
                </label>
                <select
                  id="questType"
                  className="w-full border border-gray-300 rounded-md px-3 py-2 [&>*]:w-full"
                  onChange={(e) => setSelectedQuestArchtype(e.target.value)}
                >
                  {questArchtypes.map((questArchtype) => (
                    <option key={questArchtype.id} value={questArchtype.id}>
                      {questArchtype.id}
                    </option>
                  ))}
                </select>
              </div>
              <button
                disabled={!selectedQuestArchtype}
                onClick={() => {
                  if (selectedQuestArchtype) {
                    generateQuest(id!, selectedQuestArchtype);
                    setIsGeneratingQuest(false);
                  }
                }}
                className="w-full bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600"
              >
                Generate Quest
              </button>
              <button
                onClick={() => setIsGeneratingQuest(false)}
                className="w-full bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {isGenerating && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg shadow-xl w-96">
            <h2 className="text-xl font-bold mb-4">
              Generate Points of Interest
            </h2>

            {placeTypesLoading ? (
              <p>Loading place types...</p>
            ) : placeTypesError ? (
              <p className="text-red-500">
                Error loading place types: {placeTypesError.message}
              </p>
            ) : (
              <div className="space-y-4">
                <div>
                  <div className="mb-4">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">
                      Selected Included Types:
                    </h3>
                    <div className="flex flex-wrap gap-2">
                      {selectedIncludedPlaceTypes.map((type) => (
                        <span
                          key={type}
                          className="px-2 py-1 bg-blue-100 text-blue-800 rounded-md text-sm"
                        >
                          {type}
                        </span>
                      ))}
                      {selectedIncludedPlaceTypes.length === 0 && (
                        <span className="text-gray-500 text-sm">
                          No types selected
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="mb-4">
                    <h3 className="text-sm font-medium text-gray-700 mb-2">
                      Selected Excluded Types:
                    </h3>
                    <div className="flex flex-wrap gap-2">
                      {selectedExcludedPlaceTypes.map((type) => (
                        <span
                          key={type}
                          className="px-2 py-1 bg-red-100 text-red-800 rounded-md text-sm"
                        >
                          {type}
                        </span>
                      ))}
                      {selectedExcludedPlaceTypes.length === 0 && (
                        <span className="text-gray-500 text-sm">
                          No types selected
                        </span>
                      )}
                    </div>
                  </div>
                  <label
                    htmlFor="placeType"
                    className="block text-sm font-medium text-gray-700 mb-1"
                  >
                    Included Place Types
                  </label>
                  <select
                    id="placeType"
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    multiple
                    value={selectedIncludedPlaceTypes}
                    onChange={(e) => {
                      const selectedOptions = Array.from(
                        e.target.selectedOptions,
                        (option) => option.value
                      );
                      setSelectedIncludedPlaceTypes(selectedOptions);
                    }}
                  >
                    {placeTypes.map((type) => (
                      <option key={type} value={type}>
                        {type}
                      </option>
                    ))}
                  </select>
                  <label
                    htmlFor="excludedPlaceTypes"
                    className="block text-sm font-medium text-gray-700 mb-1 mt-4"
                  >
                    Excluded Place Types
                  </label>
                  <select
                    id="excludedPlaceTypes"
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    multiple
                    value={selectedExcludedPlaceTypes}
                    onChange={(e) => {
                      const selectedOptions = Array.from(
                        e.target.selectedOptions,
                        (option) => option.value
                      );
                      setSelectedExcludedPlaceTypes(selectedOptions);
                    }}
                  >
                    {placeTypes.map((type) => (
                      <option key={type} value={type}>
                        {type}
                      </option>
                    ))}
                  </select>
                  <label
                    htmlFor="numPlaces"
                    className="block text-sm font-medium text-gray-700 mb-1"
                  >
                    Number of Places
                  </label>
                  <input
                    id="numPlaces"
                    type="number"
                    min="1"
                    max="20"
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    value={numPlaces}
                    onChange={(e) =>
                      setNumPlaces(
                        Math.min(20, Math.max(1, parseInt(e.target.value)))
                      )
                    }
                  />
                </div>

                {generatePointsOfInterestError && (
                  <p className="text-red-500">
                    Error: {generatePointsOfInterestError}
                  </p>
                )}

                <div className="flex justify-end gap-2">
                  <button
                    onClick={() => setIsGenerating(false)}
                    className="bg-gray-200 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-300"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => {
                      if (
                        selectedIncludedPlaceTypes &&
                        selectedExcludedPlaceTypes
                      ) {
                        generatePointsOfInterest(
                          id!,
                          selectedIncludedPlaceTypes,
                          selectedExcludedPlaceTypes,
                          numPlaces
                        );
                        setIsGenerating(false);
                      }
                    }}
                    disabled={
                      !selectedIncludedPlaceTypes || !selectedExcludedPlaceTypes
                    }
                    className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600 disabled:bg-blue-300"
                  >
                    Generate
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
