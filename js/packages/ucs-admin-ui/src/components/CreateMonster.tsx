import React, { useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { useNavigate } from 'react-router-dom';

interface MonsterAbility {
  name: string;
  description: string;
  attack_bonus?: number;
  damage?: string;
  damage_type?: string;
  additional_damage?: string;
  save_dc?: number;
  save_ability?: string;
  area?: string;
  special?: string;
  recharge?: string;
}

export const CreateMonster = () => {
  const navigate = useNavigate();
  const { apiClient } = useAPI();
  
  const [formData, setFormData] = useState({
    name: '',
    size: 'Medium',
    type: '',
    subtype: '',
    alignment: 'unaligned',
    
    armorClass: 10,
    hitPoints: 1,
    hitDice: '',
    speed: 30,
    speedModifiers: {} as { [key: string]: number },
    
    strength: 10,
    dexterity: 10,
    constitution: 10,
    intelligence: 10,
    wisdom: 10,
    charisma: 10,
    
    proficiencyBonus: 2,
    challengeRating: 0,
    experiencePoints: 0,
    
    savingThrowProficiencies: [] as string[],
    skillProficiencies: {} as { [key: string]: number },
    
    damageVulnerabilities: [] as string[],
    damageResistances: [] as string[],
    damageImmunities: [] as string[],
    conditionImmunities: [] as string[],
    
    blindsight: 0,
    darkvision: 0,
    tremorsense: 0,
    truesight: 0,
    passivePerception: 10,
    
    languages: [] as string[],
    
    specialAbilities: [] as MonsterAbility[],
    actions: [] as MonsterAbility[],
    legendaryActions: [] as MonsterAbility[],
    legendaryActionsPerTurn: 0,
    reactions: [] as MonsterAbility[],
    
    imageUrl: '',
    description: '',
    flavorText: '',
    environment: '',
    source: 'Custom',
  });

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  const sizeOptions = ['Tiny', 'Small', 'Medium', 'Large', 'Huge', 'Gargantuan'];
  const creatureTypes = [
    'aberration', 'beast', 'celestial', 'construct', 'dragon', 'elemental',
    'fey', 'fiend', 'giant', 'humanoid', 'monstrosity', 'ooze', 'plant', 'undead'
  ];
  const alignmentOptions = [
    'lawful good', 'neutral good', 'chaotic good',
    'lawful neutral', 'neutral', 'chaotic neutral',
    'lawful evil', 'neutral evil', 'chaotic evil',
    'unaligned'
  ];
  const abilityScores = ['Strength', 'Dexterity', 'Constitution', 'Intelligence', 'Wisdom', 'Charisma'];
  const skills = [
    'Acrobatics', 'Animal Handling', 'Arcana', 'Athletics', 'Deception',
    'History', 'Insight', 'Intimidation', 'Investigation', 'Medicine',
    'Nature', 'Perception', 'Performance', 'Persuasion', 'Religion',
    'Sleight of Hand', 'Stealth', 'Survival'
  ];
  const damageTypes = [
    'acid', 'bludgeoning', 'cold', 'fire', 'force', 'lightning', 'necrotic',
    'piercing', 'poison', 'psychic', 'radiant', 'slashing', 'thunder'
  ];
  const conditions = [
    'blinded', 'charmed', 'deafened', 'exhaustion', 'frightened', 'grappled',
    'incapacitated', 'invisible', 'paralyzed', 'petrified', 'poisoned',
    'prone', 'restrained', 'stunned', 'unconscious'
  ];
  const languages = [
    'Common', 'Draconic', 'Elvish', 'Dwarvish', 'Halfling', 'Orc', 'Giant',
    'Gnomish', 'Goblin', 'Abyssal', 'Celestial', 'Infernal', 'Sylvan',
    'Primordial', 'Deep Speech', 'Telepathy'
  ];

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  const handleArrayChange = (field: string, values: string[]) => {
    setFormData(prev => ({ ...prev, [field]: values }));
  };

  const addAbility = (abilityType: 'specialAbilities' | 'actions' | 'legendaryActions' | 'reactions') => {
    const newAbility: MonsterAbility = { name: '', description: '' };
    setFormData(prev => ({
      ...prev,
      [abilityType]: [...prev[abilityType], newAbility]
    }));
  };

  const updateAbility = (
    abilityType: 'specialAbilities' | 'actions' | 'legendaryActions' | 'reactions',
    index: number,
    field: string,
    value: any
  ) => {
    setFormData(prev => ({
      ...prev,
      [abilityType]: prev[abilityType].map((ability, i) => 
        i === index ? { ...ability, [field]: value } : ability
      )
    }));
  };

  const removeAbility = (abilityType: 'specialAbilities' | 'actions' | 'legendaryActions' | 'reactions', index: number) => {
    setFormData(prev => ({
      ...prev,
      [abilityType]: prev[abilityType].filter((_, i) => i !== index)
    }));
  };

  const validateForm = () => {
    const newErrors: { [key: string]: string } = {};

    if (!formData.name.trim()) newErrors.name = 'Name is required';
    if (!formData.type.trim()) newErrors.type = 'Creature type is required';
    if (formData.armorClass < 1) newErrors.armorClass = 'Armor class must be at least 1';
    if (formData.hitPoints < 1) newErrors.hitPoints = 'Hit points must be at least 1';
    if (formData.speed < 0) newErrors.speed = 'Speed cannot be negative';
    if (formData.challengeRating < 0) newErrors.challengeRating = 'Challenge rating cannot be negative';
    if (formData.experiencePoints < 0) newErrors.experiencePoints = 'Experience points cannot be negative';
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) return;
    
    setIsSubmitting(true);
    
    try {
      // Clean up form data - remove empty arrays and objects
      const cleanedData = {
        ...formData,
        subtype: formData.subtype.trim() || undefined,
        hitDice: formData.hitDice.trim() || undefined,
        imageUrl: formData.imageUrl.trim() || undefined,
        description: formData.description.trim() || undefined,
        flavorText: formData.flavorText.trim() || undefined,
        environment: formData.environment.trim() || undefined,
        speedModifiers: Object.keys(formData.speedModifiers).length > 0 ? formData.speedModifiers : undefined,
        savingThrowProficiencies: formData.savingThrowProficiencies.length > 0 ? formData.savingThrowProficiencies : undefined,
        skillProficiencies: Object.keys(formData.skillProficiencies).length > 0 ? formData.skillProficiencies : undefined,
        damageVulnerabilities: formData.damageVulnerabilities.length > 0 ? formData.damageVulnerabilities : undefined,
        damageResistances: formData.damageResistances.length > 0 ? formData.damageResistances : undefined,
        damageImmunities: formData.damageImmunities.length > 0 ? formData.damageImmunities : undefined,
        conditionImmunities: formData.conditionImmunities.length > 0 ? formData.conditionImmunities : undefined,
        languages: formData.languages.length > 0 ? formData.languages : undefined,
        specialAbilities: formData.specialAbilities.length > 0 ? formData.specialAbilities : undefined,
        actions: formData.actions.length > 0 ? formData.actions : undefined,
        legendaryActions: formData.legendaryActions.length > 0 ? formData.legendaryActions : undefined,
        reactions: formData.reactions.length > 0 ? formData.reactions : undefined,
      };

      await apiClient.post('/sonar/admin/monsters', cleanedData);
      navigate('/monsters');
    } catch (error: any) {
      console.error('Failed to create monster:', error);
      setErrors({ submit: error.response?.data?.error || 'Failed to create monster' });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Create New Monster</h1>
        <p className="text-gray-600 mt-2">Add a new monster to the database</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Information */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Basic Information</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="Enter monster name"
              />
              {errors.name && <p className="text-red-500 text-sm mt-1">{errors.name}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Size *</label>
              <select
                value={formData.size}
                onChange={(e) => handleInputChange('size', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              >
                {sizeOptions.map(size => (
                  <option key={size} value={size}>{size}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Type *</label>
              <select
                value={formData.type}
                onChange={(e) => handleInputChange('type', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              >
                <option value="">Select type</option>
                {creatureTypes.map(type => (
                  <option key={type} value={type}>{type}</option>
                ))}
              </select>
              {errors.type && <p className="text-red-500 text-sm mt-1">{errors.type}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Subtype</label>
              <input
                type="text"
                value={formData.subtype}
                onChange={(e) => handleInputChange('subtype', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="e.g. elf, dwarf, red dragon"
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700">Alignment *</label>
              <select
                value={formData.alignment}
                onChange={(e) => handleInputChange('alignment', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              >
                {alignmentOptions.map(alignment => (
                  <option key={alignment} value={alignment}>{alignment}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Core Stats */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Core Stats</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Armor Class *</label>
              <input
                type="number"
                min="1"
                value={formData.armorClass}
                onChange={(e) => handleInputChange('armorClass', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
              {errors.armorClass && <p className="text-red-500 text-sm mt-1">{errors.armorClass}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Hit Points *</label>
              <input
                type="number"
                min="1"
                value={formData.hitPoints}
                onChange={(e) => handleInputChange('hitPoints', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
              {errors.hitPoints && <p className="text-red-500 text-sm mt-1">{errors.hitPoints}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Hit Dice</label>
              <input
                type="text"
                value={formData.hitDice}
                onChange={(e) => handleInputChange('hitDice', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="e.g. 2d8 + 2"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Speed (ft) *</label>
              <input
                type="number"
                min="0"
                value={formData.speed}
                onChange={(e) => handleInputChange('speed', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
              {errors.speed && <p className="text-red-500 text-sm mt-1">{errors.speed}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Challenge Rating *</label>
              <input
                type="number"
                min="0"
                step="0.125"
                value={formData.challengeRating}
                onChange={(e) => handleInputChange('challengeRating', parseFloat(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
              {errors.challengeRating && <p className="text-red-500 text-sm mt-1">{errors.challengeRating}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Experience Points *</label>
              <input
                type="number"
                min="0"
                value={formData.experiencePoints}
                onChange={(e) => handleInputChange('experiencePoints', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
              {errors.experiencePoints && <p className="text-red-500 text-sm mt-1">{errors.experiencePoints}</p>}
            </div>
          </div>
        </div>

        {/* Ability Scores */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Ability Scores</h2>
          
          <div className="grid grid-cols-3 md:grid-cols-6 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Strength</label>
              <input
                type="number"
                min="1"
                max="30"
                value={formData.strength}
                onChange={(e) => handleInputChange('strength', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Dexterity</label>
              <input
                type="number"
                min="1"
                max="30"
                value={formData.dexterity}
                onChange={(e) => handleInputChange('dexterity', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Constitution</label>
              <input
                type="number"
                min="1"
                max="30"
                value={formData.constitution}
                onChange={(e) => handleInputChange('constitution', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Intelligence</label>
              <input
                type="number"
                min="1"
                max="30"
                value={formData.intelligence}
                onChange={(e) => handleInputChange('intelligence', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Wisdom</label>
              <input
                type="number"
                min="1"
                max="30"
                value={formData.wisdom}
                onChange={(e) => handleInputChange('wisdom', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Charisma</label>
              <input
                type="number"
                min="1"
                max="30"
                value={formData.charisma}
                onChange={(e) => handleInputChange('charisma', parseInt(e.target.value))}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>
          </div>
        </div>

        {/* Additional Information */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Additional Information</h2>
          
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Image URL</label>
              <input
                type="url"
                value={formData.imageUrl}
                onChange={(e) => handleInputChange('imageUrl', e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="https://example.com/monster-image.jpg"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Description</label>
              <textarea
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                rows={4}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="Describe the monster's appearance, behavior, and lore..."
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">Flavor Text</label>
              <textarea
                value={formData.flavorText}
                onChange={(e) => handleInputChange('flavorText', e.target.value)}
                rows={2}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                placeholder="A memorable quote or short description..."
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Environment</label>
                <input
                  type="text"
                  value={formData.environment}
                  onChange={(e) => handleInputChange('environment', e.target.value)}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  placeholder="e.g. Forest, Grassland, Underground"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Source</label>
                <input
                  type="text"
                  value={formData.source}
                  onChange={(e) => handleInputChange('source', e.target.value)}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  placeholder="e.g. Monster Manual, Custom"
                />
              </div>
            </div>
          </div>
        </div>

        {/* Submit */}
        <div className="flex justify-end gap-4">
          <button
            type="button"
            onClick={() => navigate('/monsters')}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {isSubmitting ? 'Creating...' : 'Create Monster'}
          </button>
        </div>

        {errors.submit && (
          <div className="text-red-500 text-sm mt-2">{errors.submit}</div>
        )}
      </form>
    </div>
  );
};

export default CreateMonster;