import React, { useState } from 'react';
import { usePlaces } from '@poltergeist/hooks';
import { Place } from '@poltergeist/types';

interface AddNewPointOfInterestProps {
  onSave: (name: string, description: string, lat: number, lng: number, image: File | null, clue: string, unlockTier?: number | null) => void;
  onCancel: () => void;
}

export const AddNewPointOfInterest = ({ onSave, onCancel }: AddNewPointOfInterestProps) => {
  const [name, setName] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [lat, setLat] = React.useState(0);
  const [lng, setLng] = React.useState(0);
  const [address, setAddress] = React.useState('');
  const [image, setImage] = React.useState<File | null>(null);
  const [imagePreview, setImagePreview] = React.useState<string | null>(null);
  const [clue, setClue] = React.useState('');
  const [unlockTier, setUnlockTier] = React.useState<number | null>(null);
  const timeoutRef = React.useRef<number>();
  const { places } = usePlaces(address);
  const [showPlaces, setShowPlaces] = React.useState(false);

  const handleAddressChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      setAddress(e.target.value);
      setShowPlaces(true);
    }, 500);
  };

  const handlePlaceSelect = (place: Place) => {
    setLat(place.latLong.lat);
    setLng(place.latLong.lng);
    setAddress(place.name);
    setShowPlaces(false);
  };

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setImage(file);
      
      // Create preview URL
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  return <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
  <div className="bg-white p-6 rounded-lg w-96">
    <h3 className="text-xl font-bold mb-4">Add New Point of Interest</h3>
    <input
      type="text"
      placeholder="Name"
      value={name}
      onChange={(e) => setName(e.target.value)}
      className="border rounded px-2 py-1 w-full mb-2"
    />
    <textarea
      placeholder="Description"
      value={description}
      onChange={(e) => setDescription(e.target.value)}
      className="border rounded px-2 py-1 w-full mb-2"
    />
    <textarea
      placeholder="Clue"
      value={clue}
      onChange={(e) => setClue(e.target.value)}
      className="border rounded px-2 py-1 w-full mb-2"
    />
    <div className="mb-4">
      <input
        type="file"
        accept="image/*"
        onChange={handleImageChange}
        className="mb-2"
      />
      {imagePreview && (
        <div className="mt-2">
          <img 
            src={imagePreview} 
            alt="Preview" 
            className="max-w-full h-auto rounded"
            style={{ maxHeight: '200px' }}
          />
        </div>
      )}
    </div>
    <div className="relative">
      <input
        type="text"
        placeholder="Search address"
        onChange={handleAddressChange}
        className="border rounded px-2 py-1 w-full mb-2"
      />
      {showPlaces && places && places.length > 0 && (
        <div className="absolute z-10 w-full bg-white border rounded-lg shadow-lg max-h-48 overflow-y-auto">
          {places.map((place, index) => (
            <div
              key={index}
              className="px-4 py-2 hover:bg-gray-100 cursor-pointer"
              onClick={() => handlePlaceSelect(place)}
            >
              {place.name}
            </div>
          ))}
        </div>
      )}
    </div>
    <input
      type="number"
      placeholder="Latitude"
      value={lat}
      onChange={(e) => setLat(parseFloat(e.target.value))}
      className="border rounded px-2 py-1 w-full mb-2"
      step="0.000001"
    />
    <input
      type="number"
      placeholder="Longitude"
      value={lng}
      onChange={(e) => setLng(parseFloat(e.target.value))}
      className="border rounded px-2 py-1 w-full mb-2"
      step="0.000001"
    />
    <input
      type="number"
      placeholder="Unlock Tier (optional)"
      value={unlockTier ?? ''}
      onChange={(e) => setUnlockTier(e.target.value ? parseInt(e.target.value) : null)}
      className="border rounded px-2 py-1 w-full mb-4"
      min="1"
    />
    <div className="flex gap-2 justify-end">
      <button
        onClick={onCancel}
        className="bg-gray-500 text-white px-4 py-2 rounded"
      >
        Cancel
      </button>
      <button
        onClick={() => onSave(name, description, lat, lng, image, clue, unlockTier)}
        className="bg-green-500 text-white px-4 py-2 rounded"
      >
        Add Point
      </button>
    </div>
  </div>
</div>;
};