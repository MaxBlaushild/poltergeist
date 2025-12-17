import { useAPI } from '@poltergeist/contexts';
import { UtilityClosetPuzzle, HueLight, ButtonConfig, PUZZLE_COLORS, COLOR_TO_INDEX, INDEX_TO_COLOR } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const UtilityClosetPuzzleAdmin = () => {
  const { apiClient } = useAPI();
  const [puzzle, setPuzzle] = useState<UtilityClosetPuzzle | null>(null);
  const [hueLights, setHueLights] = useState<HueLight[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [buttonConfigs, setButtonConfigs] = useState<ButtonConfig[]>([]);

  useEffect(() => {
    fetchPuzzle();
    fetchHueLights();
  }, []);

  useEffect(() => {
    if (puzzle) {
      const configs: ButtonConfig[] = [];
      for (let i = 0; i < 6; i++) {
        const hueLightId = puzzle[`button${i}HueLightId` as keyof UtilityClosetPuzzle] as number | null | undefined;
        const baseHue = puzzle[`button${i}BaseHue` as keyof UtilityClosetPuzzle] as number;
        configs.push({
          slot: i,
          hueLightId: hueLightId ?? null,
          baseHue,
        });
      }
      setButtonConfigs(configs);
    }
  }, [puzzle]);

  const fetchPuzzle = async () => {
    try {
      const response = await apiClient.get<UtilityClosetPuzzle>('/final-fete/utility-closet-puzzle');
      setPuzzle(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching puzzle:', error);
      setError('Failed to load puzzle configuration');
      setLoading(false);
    }
  };

  const fetchHueLights = async () => {
    try {
      const response = await apiClient.get<HueLight[]>('/final-fete/hue-lights');
      setHueLights(response);
    } catch (error) {
      console.error('Error fetching hue lights:', error);
      // Don't show error to user, just log it - hue lights are optional
    }
  };

  const updateButtonConfig = (slot: number, field: 'hueLightId' | 'baseHue', value: number | null) => {
    setButtonConfigs(prev => prev.map(config => {
      if (config.slot === slot) {
        return { ...config, [field]: value };
      }
      return config;
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      const updatedPuzzle = await apiClient.put<UtilityClosetPuzzle>('/final-fete/utility-closet-puzzle', {
        buttons: buttonConfigs,
      });
      setPuzzle(updatedPuzzle);
      alert('Puzzle configuration saved successfully!');
    } catch (error) {
      console.error('Error saving puzzle:', error);
      setError('Failed to save puzzle configuration');
      alert('Error saving puzzle configuration. Please try again.');
    } finally {
      setSaving(false);
    }
  };

  const getColorDisplayStyle = (baseHue: number): React.CSSProperties => {
    const colorMap: Record<number, string> = {
      0: '#808080', // Off (grey)
      1: '#0000FF', // Blue
      2: '#00FF00', // Green
      3: '#FFFFFF', // White
      4: '#FF0000', // Red
      5: '#800080', // Purple
    };
    const color = colorMap[baseHue] || '#808080';
    return {
      backgroundColor: color,
      color: baseHue === 3 ? '#000000' : '#FFFFFF', // White text on dark colors, black on white
      padding: '4px 8px',
      borderRadius: '4px',
      display: 'inline-block',
      minWidth: '60px',
      textAlign: 'center' as const,
    };
  };

  if (loading) {
    return <div className="m-10">Loading puzzle configuration...</div>;
  }

  if (!puzzle) {
    return <div className="m-10 text-red-500">Failed to load puzzle configuration</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Utility Closet Puzzle Configuration</h1>
      
      {error && (
        <div className="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      <div className="mb-6">
        <p className="text-gray-600 mb-4">
          Configure the Hue light IDs and base colors for each of the 6 puzzle buttons.
          Colors: Off (grey), Blue, Green, White, Red, Purple
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
        {buttonConfigs.map((config) => (
          <div key={config.slot} className="p-4 border rounded-lg bg-white shadow">
            <h2 className="text-lg font-semibold mb-3">Button {config.slot}</h2>
            
            <div className="mb-4">
              <label className="block mb-2 text-sm font-medium">Hue Light</label>
              <select
                value={config.hueLightId?.toString() || ''}
                onChange={(e) => {
                  const value = e.target.value === '' ? null : parseInt(e.target.value, 10);
                  updateButtonConfig(config.slot, 'hueLightId', value);
                }}
                className="w-full p-2 border rounded-md"
              >
                <option value="">None</option>
                {hueLights.map(light => (
                  <option key={light.id} value={light.id.toString()}>
                    {light.name} (ID: {light.id})
                  </option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block mb-2 text-sm font-medium">Base Color</label>
              <select
                value={INDEX_TO_COLOR[config.baseHue]}
                onChange={(e) => {
                  const color = e.target.value as typeof PUZZLE_COLORS[number];
                  const index = COLOR_TO_INDEX[color];
                  updateButtonConfig(config.slot, 'baseHue', index);
                }}
                className="w-full p-2 border rounded-md"
              >
                {PUZZLE_COLORS.map(color => (
                  <option key={color} value={color}>
                    {color}
                  </option>
                ))}
              </select>
              <div className="mt-2">
                <span style={getColorDisplayStyle(config.baseHue)}>
                  {INDEX_TO_COLOR[config.baseHue]}
                </span>
              </div>
            </div>

            <div className="text-xs text-gray-500">
              <div>Current Hue: {INDEX_TO_COLOR[puzzle[`button${config.slot}CurrentHue` as keyof UtilityClosetPuzzle] as number]}</div>
            </div>
          </div>
        ))}
      </div>

      <div className="flex gap-4">
        <button
          onClick={handleSave}
          disabled={saving}
          className="bg-blue-500 hover:bg-blue-600 disabled:bg-gray-400 text-white px-6 py-2 rounded-md font-medium"
        >
          {saving ? 'Saving...' : 'Save Configuration'}
        </button>
        <button
          onClick={fetchPuzzle}
          className="bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-md font-medium"
        >
          Cancel / Reload
        </button>
      </div>
    </div>
  );
};
