import React, { useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { useInventory } from '@poltergeist/contexts';

interface InventoryItemFormData {
  name: string;
  imageUrl: string;
  flavorText: string;
  effectText: string;
  rarityTier: string;
  isCaptureType: boolean;
  itemType: string;
  equipmentSlot: string;
  stats: {
    strengthBonus: number;
    dexterityBonus: number;
    constitutionBonus: number;
    intelligenceBonus: number;
    wisdomBonus: number;
    charismaBonus: number;
  };
}

const RARITY_TIERS = ['Common', 'Uncommon', 'Epic', 'Mythic', 'Not Droppable'];
const ITEM_TYPES = ['passive', 'consumable', 'equippable'];
const EQUIPMENT_SLOTS = [
  'head', 'chest', 'legs', 'feet', 'left_hand', 'right_hand', 
  'neck', 'ring', 'belt', 'gloves'
];

export const CreateInventoryItem = () => {
  const { apiClient } = useAPI();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = useState(false);
  
  const [formData, setFormData] = useState<InventoryItemFormData>({
    name: '',
    imageUrl: '',
    flavorText: '',
    effectText: '',
    rarityTier: 'Common',
    isCaptureType: false,
    itemType: 'passive',
    equipmentSlot: '',
    stats: {
      strengthBonus: 0,
      dexterityBonus: 0,
      constitutionBonus: 0,
      intelligenceBonus: 0,
      wisdomBonus: 0,
      charismaBonus: 0,
    }
  });

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleStatChange = (stat: string, value: number) => {
    setFormData(prev => ({
      ...prev,
      stats: {
        ...prev.stats,
        [stat]: value
      }
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setSubmitError(null);
    setSubmitSuccess(false);

    try {
      // Prepare the payload
      const payload: any = {
        name: formData.name,
        imageUrl: formData.imageUrl,
        flavorText: formData.flavorText,
        effectText: formData.effectText,
        rarityTier: formData.rarityTier,
        isCaptureType: formData.isCaptureType,
        itemType: formData.itemType,
      };

      // Add equipment slot if item is equippable
      if (formData.itemType === 'equippable' && formData.equipmentSlot) {
        payload.equipmentSlot = formData.equipmentSlot;
      }

      // Add stats if any bonus is non-zero
      const hasStats = Object.values(formData.stats).some(value => value !== 0);
      if (hasStats) {
        payload.stats = formData.stats;
      }

      await apiClient.post('/sonar/admin/items', payload);
      
      setSubmitSuccess(true);
      // Reset form
      setFormData({
        name: '',
        imageUrl: '',
        flavorText: '',
        effectText: '',
        rarityTier: 'Common',
        isCaptureType: false,
        itemType: 'passive',
        equipmentSlot: '',
        stats: {
          strengthBonus: 0,
          dexterityBonus: 0,
          constitutionBonus: 0,
          intelligenceBonus: 0,
          wisdomBonus: 0,
          charismaBonus: 0,
        }
      });
      
      // Clear success message after 3 seconds
      setTimeout(() => setSubmitSuccess(false), 3000);
      
    } catch (error: any) {
      setSubmitError(error.response?.data?.error || error.message || 'Failed to create item');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Create New Inventory Item</h1>
      
      {submitSuccess && (
        <div className="mb-6 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
          Item created successfully!
        </div>
      )}
      
      {submitError && (
        <div className="mb-6 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
          Error: {submitError}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Information */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Basic Information</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Name *
              </label>
              <input
                type="text"
                required
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="Enter item name"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Image URL *
              </label>
              <input
                type="url"
                required
                value={formData.imageUrl}
                onChange={(e) => handleInputChange('imageUrl', e.target.value)}
                className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="https://example.com/item-image.png"
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Flavor Text
              </label>
              <textarea
                value={formData.flavorText}
                onChange={(e) => handleInputChange('flavorText', e.target.value)}
                rows={3}
                className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="Descriptive text about the item's lore or appearance"
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Effect Text
              </label>
              <textarea
                value={formData.effectText}
                onChange={(e) => handleInputChange('effectText', e.target.value)}
                rows={3}
                className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="Description of what the item does mechanically"
              />
            </div>
          </div>
        </div>

        {/* Item Properties */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Item Properties</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Rarity Tier *
              </label>
              <select
                required
                value={formData.rarityTier}
                onChange={(e) => handleInputChange('rarityTier', e.target.value)}
                className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              >
                {RARITY_TIERS.map(tier => (
                  <option key={tier} value={tier}>{tier}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Item Type *
              </label>
              <select
                required
                value={formData.itemType}
                onChange={(e) => handleInputChange('itemType', e.target.value)}
                className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              >
                {ITEM_TYPES.map(type => (
                  <option key={type} value={type}>
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </option>
                ))}
              </select>
            </div>

            {formData.itemType === 'equippable' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Equipment Slot *
                </label>
                <select
                  required
                  value={formData.equipmentSlot}
                  onChange={(e) => handleInputChange('equipmentSlot', e.target.value)}
                  className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                >
                  <option value="">Select slot...</option>
                  {EQUIPMENT_SLOTS.map(slot => (
                    <option key={slot} value={slot}>
                      {slot.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                    </option>
                  ))}
                </select>
              </div>
            )}
          </div>

          <div className="mt-4">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={formData.isCaptureType}
                onChange={(e) => handleInputChange('isCaptureType', e.target.checked)}
                className="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
              <span className="ml-2 text-sm text-gray-700">
                Is Capture Type (can be used for instant captures)
              </span>
            </label>
          </div>
        </div>

        {/* Stats */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Stat Bonuses</h2>
          <p className="text-sm text-gray-600 mb-4">
            Set stat bonuses for this item. Leave at 0 if no bonus is provided.
          </p>
          
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {Object.entries(formData.stats).map(([stat, value]) => (
              <div key={stat}>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {stat.replace('Bonus', '').charAt(0).toUpperCase() + stat.replace('Bonus', '').slice(1)}
                </label>
                <input
                  type="number"
                  min="-10"
                  max="10"
                  value={value}
                  onChange={(e) => handleStatChange(stat, parseInt(e.target.value) || 0)}
                  className="w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                />
              </div>
            ))}
          </div>
        </div>

        {/* Submit Button */}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={isSubmitting}
            className="inline-flex justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? 'Creating...' : 'Create Item'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default CreateInventoryItem;