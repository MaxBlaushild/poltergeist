import React, { useState } from 'react';
import { CreatePointOfInterestPayload, useCreatePointOfInterest } from '../hooks/useCreatePointOfInterest.ts';

export const CreatePointsOfInterest = () => {
  const { createPointOfInterest } = useCreatePointOfInterest();
  const [formData, setFormData] = useState<CreatePointOfInterestPayload>({
    name: '',
    description: '',
    imageUrl: '',
    lat: '',
    lon: '', 
    clue: '',
    tierOne: '',
    tierTwo: '',
    tierThree: '',
    tierOneInventoryItemId: 0,
    tierTwoInventoryItemId: 0,
    tierThreeInventoryItemId: 0,
    pointOfInterestGroupMemberId: ''
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createPointOfInterest(formData);
      alert('Point of interest created successfully');
    } catch (error) {
      alert(`Error creating point of interest: ${error}`);
    }
  };

  return (
    <div className="Admin__background">

    <div className="p-4">
      <h2 className="text-xl font-bold mb-4">Create Point of Interest</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block mb-1">Name</label>
          <input
            type="text"
            name="name"
            value={formData.name}
            onChange={handleChange}
            className="w-full p-2 border rounded"
            required
          />
        </div>

        <div>
          <label className="block mb-1">Description</label>
          <textarea
            name="description"
            value={formData.description}
            onChange={handleChange}
            className="w-full p-2 border rounded"
            required
          />
        </div>

        <div>
          <label className="block mb-1">Image URL</label>
          <input
            type="text"
            name="imageUrl"
            value={formData.imageUrl}
            onChange={handleChange}
            className="w-full p-2 border rounded"
            required
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block mb-1">Latitude</label>
            <input
              type="text"
              name="lat"
              value={formData.lat}
              onChange={handleChange}
              className="w-full p-2 border rounded"
              required
            />
          </div>
          <div>
            <label className="block mb-1">Longitude</label>
            <input
              type="text"
              name="lon"
              value={formData.lon}
              onChange={handleChange}
              className="w-full p-2 border rounded"
              required
            />
          </div>
        </div>

        <div>
          <label className="block mb-1">Clue</label>
          <input
            type="text"
            name="clue"
            value={formData.clue}
            onChange={handleChange}
            className="w-full p-2 border rounded"
            required
          />
        </div>

        <div>
          <label className="block mb-1">Tier One</label>
          <input
            type="text"
            name="tierOne"
            value={formData.tierOne}
            onChange={handleChange}
            className="w-full p-2 border rounded"
            required
          />
        </div>

        <div>
          <label className="block mb-1">Tier Two</label>
          <input
            type="text"
            name="tierTwo"
            value={formData.tierTwo}
            onChange={handleChange}
            className="w-full p-2 border rounded"
          />
        </div>

        <div>
          <label className="block mb-1">Tier Three</label>
          <input
            type="text"
            name="tierThree"
            value={formData.tierThree}
            onChange={handleChange}
            className="w-full p-2 border rounded"
          />
        </div>

        <div>
          <label className="block mb-1">Tier One Inventory Item ID</label>
          <input
            type="number"
            name="tierOneInventoryItemId"
            value={formData.tierOneInventoryItemId}
            onChange={handleChange}
            className="w-full p-2 border rounded"
          />
        </div>

        <div>
          <label className="block mb-1">Tier Two Inventory Item ID</label>
          <input
            type="number"
            name="tierTwoInventoryItemId"
            value={formData.tierTwoInventoryItemId}
            onChange={handleChange}
            className="w-full p-2 border rounded"
          />
        </div>

        <div>
          <label className="block mb-1">Tier Three Inventory Item ID</label>
          <input
            type="number"
            name="tierThreeInventoryItemId"
            value={formData.tierThreeInventoryItemId}
            onChange={handleChange}
            className="w-full p-2 border rounded"
          />
        </div>

        <div>
          <label className="block mb-1">Point of Interest Group Member ID</label>
          <input
            type="text"
            name="pointOfInterestGroupMemberId"
            value={formData.pointOfInterestGroupMemberId}
            onChange={handleChange}
            className="w-full p-2 border rounded"
          />
        </div>

        <button
          type="submit"
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Create Point of Interest
        </button>
      </form>
    </div>
    </div>
  );
};
