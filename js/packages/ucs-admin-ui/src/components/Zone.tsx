import React, { useState } from 'react';
import { useZoneContext } from '../contexts/zones.tsx';
import { v4 as uuidv4 } from 'uuid';
import { useParams } from 'react-router-dom';
import { useZonePointsOfInterest } from '../hooks/useZonePointsOfInterest.ts';
import { usePlaceTypes } from '../hooks/usePlaceTypes.ts';
import { useGeneratePointsOfInterest } from '../hooks/useGeneratePointsOfInterest.ts';
export const Zone = () => {
  const { id } = useParams();
  const { zones, selectedZone, setSelectedZone, createZone, deleteZone } = useZoneContext();
  const zone = zones.find((zone) => zone.id === id);
  const { pointsOfInterest, loading, error } = useZonePointsOfInterest(id!);
  const { placeTypes, loading: placeTypesLoading, error: placeTypesError } = usePlaceTypes();
  const [isGenerating, setIsGenerating] = useState(false);
  const [selectedPlaceType, setSelectedPlaceType] = useState('');
  const { loading: generatePointsOfInterestLoading, error: generatePointsOfInterestError, generatePointsOfInterest } = useGeneratePointsOfInterest(id!);
  
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
  
  return <div className="m-10 p-8 bg-white rounded-lg shadow-lg">
    <h1 className="text-3xl font-bold mb-6 text-gray-800">{zone?.name}</h1>
    <p className="text-lg text-gray-600 mb-3">Latitude: {zone?.latitude}</p>
    <p className="text-lg text-gray-600 mb-3">Longitude: {zone?.longitude}</p>
    <p className="text-lg text-gray-600 mb-3">Radius: {zone?.radius}m</p>
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {pointsOfInterest.map((point) => (
        <div key={point.id} className="bg-gray-100 p-6 rounded-lg shadow-md">
          {point.imageURL && (
            <img 
              src={point.imageURL} 
              alt={point.name}
              className="float-left mr-4 mb-4 w-32 h-32 object-cover rounded"
            />
          )}
          <h2 className="text-xl font-bold mb-2 text-gray-800">{point.name}</h2>
          <p className="text-gray-600 mb-3">Description: {point.description}</p>
          <p className="text-gray-600 mb-3">Latitude: {point.lat}</p>
          <p className="text-gray-600 mb-3">Longitude: {point.lng}</p>
          <p className="text-gray-600 mb-3">Clue: {point.clue}</p>
          <p className="text-gray-600 mb-3">Created At: {point.createdAt.toLocaleString()}</p>
          <p className="text-gray-600 mb-3">Updated At: {point.updatedAt.toLocaleString()}</p>
        </div>
      ))}
    </div>
    <button
      onClick={() => setIsGenerating(!isGenerating)}
      className="bg-blue-500 text-white px-4 py-2 rounded-md"
    >
      {isGenerating ? 'Stop Generating' : 'Generate Points of Interest'}
    </button>
    {isGenerating && (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div className="bg-white p-6 rounded-lg shadow-xl w-96">
          <h2 className="text-xl font-bold mb-4">Generate Points of Interest</h2>
          
          {placeTypesLoading ? (
            <p>Loading place types...</p>
          ) : placeTypesError ? (
            <p className="text-red-500">Error loading place types: {placeTypesError.message}</p>
          ) : (
            <div className="space-y-4">
              <div>
                <label htmlFor="placeType" className="block text-sm font-medium text-gray-700 mb-1">
                  Place Type
                </label>
                <select
                  id="placeType"
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  value={selectedPlaceType}
                  onChange={(e) => setSelectedPlaceType(e.target.value)}
                >
                  <option value="">Select a place type...</option>
                  {placeTypes.map((type) => (
                    <option key={type} value={type}>
                      {type}
                    </option>
                  ))}
                </select>
              </div>

              {generatePointsOfInterestError && (
                <p className="text-red-500">Error: {generatePointsOfInterestError}</p>
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
                    if (selectedPlaceType) {
                      generatePointsOfInterest(id!, selectedPlaceType);
                      setIsGenerating(false);
                    }
                  }}
                  disabled={!selectedPlaceType}
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
}