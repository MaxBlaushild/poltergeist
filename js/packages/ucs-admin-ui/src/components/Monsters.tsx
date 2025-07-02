import React, { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Link } from 'react-router-dom';

interface Monster {
  id: string;
  name: string;
  size: string;
  type: string;
  subtype?: string;
  alignment: string;
  armorClass: number;
  hitPoints: number;
  hitDice?: string;
  speed: number;
  speedModifiers?: { [key: string]: number };
  
  strength: number;
  dexterity: number;
  constitution: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
  
  proficiencyBonus: number;
  challengeRating: number;
  experiencePoints: number;
  
  savingThrowProficiencies?: string[];
  skillProficiencies?: { [key: string]: number };
  
  damageVulnerabilities?: string[];
  damageResistances?: string[];
  damageImmunities?: string[];
  conditionImmunities?: string[];
  
  blindsight?: number;
  darkvision?: number;
  tremorsense?: number;
  truesight?: number;
  passivePerception: number;
  
  languages?: string[];
  
  specialAbilities?: MonsterAbility[];
  actions?: MonsterAbility[];
  legendaryActions?: MonsterAbility[];
  legendaryActionsPerTurn?: number;
  reactions?: MonsterAbility[];
  
  imageUrl?: string;
  description?: string;
  flavorText?: string;
  environment?: string;
  source: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

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

export const Monsters = () => {
  const { apiClient } = useAPI();
  const [monsters, setMonsters] = useState<Monster[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filteredMonsters, setFilteredMonsters] = useState<Monster[]>([]);

  const fetchMonsters = async () => {
    setIsLoading(true);
    try {
      const response = await apiClient.get<Monster[]>('/sonar/admin/monsters');
      setMonsters(response);
      setFilteredMonsters(response);
    } catch (error) {
      console.error('Failed to fetch monsters:', error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMonsters();
  }, []);

  useEffect(() => {
    if (searchQuery.trim() === '') {
      setFilteredMonsters(monsters);
    } else {
      const filtered = monsters.filter(monster =>
        monster.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        monster.type.toLowerCase().includes(searchQuery.toLowerCase()) ||
        monster.size.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (monster.description && monster.description.toLowerCase().includes(searchQuery.toLowerCase()))
      );
      setFilteredMonsters(filtered);
    }
  }, [searchQuery, monsters]);

  const getChallengeRatingColor = (cr: number) => {
    if (cr < 1) return 'text-green-600 bg-green-100';
    if (cr < 5) return 'text-yellow-600 bg-yellow-100';
    if (cr < 11) return 'text-orange-600 bg-orange-100';
    if (cr < 17) return 'text-red-600 bg-red-100';
    return 'text-purple-600 bg-purple-100';
  };

  const getSizeColor = (size: string) => {
    switch (size) {
      case 'Tiny': return 'text-blue-600 bg-blue-100';
      case 'Small': return 'text-green-600 bg-green-100';
      case 'Medium': return 'text-gray-600 bg-gray-100';
      case 'Large': return 'text-yellow-600 bg-yellow-100';
      case 'Huge': return 'text-orange-600 bg-orange-100';
      case 'Gargantuan': return 'text-red-600 bg-red-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'beast': return 'text-green-600 bg-green-100';
      case 'humanoid': return 'text-blue-600 bg-blue-100';
      case 'dragon': return 'text-red-600 bg-red-100';
      case 'undead': return 'text-purple-600 bg-purple-100';
      case 'fiend': return 'text-red-800 bg-red-200';
      case 'celestial': return 'text-yellow-600 bg-yellow-100';
      case 'elemental': return 'text-indigo-600 bg-indigo-100';
      case 'fey': return 'text-pink-600 bg-pink-100';
      case 'aberration': return 'text-purple-800 bg-purple-200';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getStatModifier = (score: number) => {
    const modifier = Math.floor((score - 10) / 2);
    return modifier >= 0 ? `+${modifier}` : `${modifier}`;
  };

  const deleteMonster = async (id: string) => {
    if (!confirm('Are you sure you want to delete this monster?')) return;
    
    try {
      await apiClient.delete(`/sonar/admin/monsters/${id}`);
      setMonsters(monsters.filter(m => m.id !== id));
    } catch (error) {
      console.error('Failed to delete monster:', error);
      alert('Failed to delete monster');
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <div className="flex justify-between items-center mb-4">
          <h1 className="text-3xl font-bold text-gray-900">Monsters</h1>
          <Link
            to="/monsters/create"
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            Create New Monster
          </Link>
        </div>

        {/* Search */}
        <div className="max-w-md">
          <input
            type="text"
            placeholder="Search monsters..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-8">
          <div className="text-gray-500">Loading monsters...</div>
        </div>
      ) : (
        <>
          <div className="mb-4 text-sm text-gray-600">
            Showing {filteredMonsters.length} of {monsters.length} monsters
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredMonsters.map((monster) => (
              <div key={monster.id} className="bg-white rounded-lg shadow-md overflow-hidden">
                <div className="h-48 bg-gray-200 flex items-center justify-center">
                  {monster.imageUrl ? (
                    <img 
                      src={monster.imageUrl} 
                      alt={monster.name}
                      className="h-full w-full object-cover"
                      onError={(e) => {
                        const target = e.target as HTMLImageElement;
                        target.style.display = 'none';
                      }}
                    />
                  ) : (
                    <div className="text-gray-400">No Image</div>
                  )}
                </div>
                
                <div className="p-4">
                  <div className="flex items-start justify-between mb-2">
                    <h3 className="text-lg font-semibold text-gray-900">{monster.name}</h3>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getChallengeRatingColor(monster.challengeRating)}`}>
                      CR {monster.challengeRating}
                    </span>
                  </div>
                  
                  <div className="flex items-center gap-2 mb-3">
                    <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getSizeColor(monster.size)}`}>
                      {monster.size}
                    </span>
                    <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getTypeColor(monster.type)}`}>
                      {monster.type}
                    </span>
                    {monster.subtype && (
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium text-gray-600 bg-gray-100">
                        {monster.subtype}
                      </span>
                    )}
                  </div>

                  <div className="text-sm text-gray-600 mb-3">
                    <div><strong>AC:</strong> {monster.armorClass}</div>
                    <div><strong>HP:</strong> {monster.hitPoints} {monster.hitDice && `(${monster.hitDice})`}</div>
                    <div><strong>Speed:</strong> {monster.speed} ft.</div>
                    <div><strong>XP:</strong> {monster.experiencePoints.toLocaleString()}</div>
                  </div>

                  {/* Ability Scores */}
                  <div className="grid grid-cols-6 gap-1 text-xs mb-3">
                    <div className="text-center">
                      <div className="font-medium">STR</div>
                      <div>{monster.strength} ({getStatModifier(monster.strength)})</div>
                    </div>
                    <div className="text-center">
                      <div className="font-medium">DEX</div>
                      <div>{monster.dexterity} ({getStatModifier(monster.dexterity)})</div>
                    </div>
                    <div className="text-center">
                      <div className="font-medium">CON</div>
                      <div>{monster.constitution} ({getStatModifier(monster.constitution)})</div>
                    </div>
                    <div className="text-center">
                      <div className="font-medium">INT</div>
                      <div>{monster.intelligence} ({getStatModifier(monster.intelligence)})</div>
                    </div>
                    <div className="text-center">
                      <div className="font-medium">WIS</div>
                      <div>{monster.wisdom} ({getStatModifier(monster.wisdom)})</div>
                    </div>
                    <div className="text-center">
                      <div className="font-medium">CHA</div>
                      <div>{monster.charisma} ({getStatModifier(monster.charisma)})</div>
                    </div>
                  </div>

                  {monster.description && (
                    <p className="text-sm text-gray-600 mb-3 line-clamp-3">{monster.description}</p>
                  )}

                  <div className="flex justify-between items-center">
                    <div className="text-xs text-gray-500">
                      Source: {monster.source}
                    </div>
                    <div className="flex gap-2">
                      <Link
                        to={`/monsters/${monster.id}`}
                        className="text-indigo-600 hover:text-indigo-500 text-sm font-medium"
                      >
                        View
                      </Link>
                      <Link
                        to={`/monsters/${monster.id}/edit`}
                        className="text-green-600 hover:text-green-500 text-sm font-medium"
                      >
                        Edit
                      </Link>
                      <button
                        onClick={() => deleteMonster(monster.id)}
                        className="text-red-600 hover:text-red-500 text-sm font-medium"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {filteredMonsters.length === 0 && !isLoading && (
            <div className="text-center py-8">
              <div className="text-gray-500 mb-4">No monsters found</div>
              <Link
                to="/monsters/create"
                className="text-indigo-600 hover:text-indigo-500"
              >
                Create the first monster
              </Link>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default Monsters;