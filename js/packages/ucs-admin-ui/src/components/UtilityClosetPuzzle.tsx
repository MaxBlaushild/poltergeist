import { useAPI } from '@poltergeist/contexts';
import { UtilityClosetPuzzle, HueLight, ButtonConfig, PUZZLE_COLORS, COLOR_TO_INDEX, INDEX_TO_COLOR } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const UtilityClosetPuzzleAdmin = () => {
  const { apiClient } = useAPI();
  const [puzzle, setPuzzle] = useState<UtilityClosetPuzzle | null>(null);
  const [hueLights, setHueLights] = useState<HueLight[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [pressing, setPressing] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [toggling, setToggling] = useState(false);
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

  const getColorDisplayStyle = (hue: number, isLarge = false): React.CSSProperties => {
    const colorMap: Record<number, string> = {
      0: '#808080', // Off (grey)
      1: '#0000FF', // Blue
      2: '#00FF00', // Green
      3: '#FFFFFF', // White
      4: '#FF0000', // Red
      5: '#800080', // Purple
      6: '#FFD700', // Gold
    };
    const color = colorMap[hue] || '#808080';
    return {
      backgroundColor: color,
      color: hue === 3 ? '#000000' : '#FFFFFF', // White text on dark colors, black on white
      padding: isLarge ? '16px 24px' : '4px 8px',
      borderRadius: '8px',
      display: 'inline-block',
      minWidth: isLarge ? '100px' : '60px',
      textAlign: 'center' as const,
      border: hue === 3 ? '2px solid #CCCCCC' : '2px solid transparent',
      cursor: isLarge ? 'pointer' : 'default',
      transition: 'all 0.2s',
      fontWeight: isLarge ? 'bold' : 'normal',
      fontSize: isLarge ? '16px' : '12px',
    };
  };

  const handlePressButton = async (slot: number) => {
    setPressing(true);
    setError(null);
    try {
      const updatedPuzzle = await apiClient.get<UtilityClosetPuzzle>(`/final-fete/utility-closet-puzzle/press/${slot}?antiviral=true`);
      setPuzzle(updatedPuzzle);
    } catch (error) {
      console.error('Error pressing button:', error);
      setError('Failed to press button');
      alert('Error pressing button. Please try again.');
    } finally {
      setPressing(false);
    }
  };

  const handleResetPuzzle = async () => {
    setResetting(true);
    setError(null);
    try {
      const updatedPuzzle = await apiClient.post<UtilityClosetPuzzle>('/final-fete/utility-closet-puzzle/reset', {});
      setPuzzle(updatedPuzzle);
      alert('Puzzle reset to base state!');
    } catch (error) {
      console.error('Error resetting puzzle:', error);
      setError('Failed to reset puzzle');
      alert('Error resetting puzzle. Please try again.');
    } finally {
      setResetting(false);
    }
  };

  const handleToggleAchievement = async (achievementType: 'allGreens' | 'allPurples') => {
    setToggling(true);
    setError(null);
    try {
      const updatedPuzzle = await apiClient.post<UtilityClosetPuzzle>('/final-fete/utility-closet-puzzle/toggle-achievement', {
        achievementType,
      });
      setPuzzle(updatedPuzzle);
    } catch (error) {
      console.error('Error toggling achievement:', error);
      setError('Failed to toggle achievement');
      alert('Error toggling achievement. Please try again.');
    } finally {
      setToggling(false);
    }
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

      {/* Interactive Puzzle Buttons */}
      <div className="mb-8 p-6 bg-gray-50 rounded-lg">
        <h2 className="text-xl font-bold mb-4">Puzzle Controls</h2>
        <p className="text-sm text-gray-600 mb-4">
          Click a button to press it. Pressing a button affects that button and its neighbors (wrapping around).
        </p>
        <div className="flex flex-wrap gap-4 justify-center items-center mb-4">
          {[0, 1, 2, 3, 4, 5].map((slot) => {
            const currentHue = puzzle[`button${slot}CurrentHue` as keyof UtilityClosetPuzzle] as number;
            const hueLightId = puzzle[`button${slot}HueLightId` as keyof UtilityClosetPuzzle] as number | null | undefined;
            const hasLight = hueLightId !== null && hueLightId !== undefined;
            const isGold = currentHue === 6;
            const isDisabled = pressing || !hasLight || isGold;
            
            return (
              <div key={slot} className="flex flex-col items-center">
                <button
                  onClick={() => handlePressButton(slot)}
                  disabled={isDisabled}
                  style={getColorDisplayStyle(currentHue, true)}
                  className={`
                    hover:opacity-90 active:scale-95
                    ${isDisabled ? 'opacity-50 cursor-not-allowed' : 'hover:shadow-lg'}
                    ${isGold ? 'ring-4 ring-yellow-300 ring-opacity-75' : ''}
                  `}
                  title={
                    isGold 
                      ? `Button ${slot} - Puzzle Solved! 🎉`
                      : hasLight 
                        ? `Button ${slot} - Click to press`
                        : `Button ${slot} - No light assigned`
                  }
                >
                  {INDEX_TO_COLOR[currentHue]}
                </button>
                <div className="mt-2 text-xs text-gray-500">
                  Button {slot}
                  {!hasLight && <div className="text-red-500">(No light)</div>}
                  {isGold && <div className="text-yellow-600 font-bold">✓ Solved!</div>}
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex flex-col items-center gap-4">
          {/* Achievement State Toggles */}
          <div className="flex gap-6 items-center">
            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-gray-700">
                All Greens Achieved:
              </label>
              <button
                onClick={() => handleToggleAchievement('allGreens')}
                disabled={toggling}
                className={`
                  relative inline-flex h-6 w-11 items-center rounded-full transition-colors
                  ${puzzle.allGreensAchieved ? 'bg-green-500' : 'bg-gray-300'}
                  ${toggling ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                `}
                title={puzzle.allGreensAchieved ? 'Click to disable' : 'Click to enable'}
              >
                <span
                  className={`
                    inline-block h-4 w-4 transform rounded-full bg-white transition-transform
                    ${puzzle.allGreensAchieved ? 'translate-x-6' : 'translate-x-1'}
                  `}
                />
              </button>
              <span className={`text-sm font-semibold ${puzzle.allGreensAchieved ? 'text-green-600' : 'text-gray-500'}`}>
                {puzzle.allGreensAchieved ? 'Enabled' : 'Disabled'}
              </span>
            </div>
            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-gray-700">
                All Purples Achieved:
              </label>
              <button
                onClick={() => handleToggleAchievement('allPurples')}
                disabled={toggling}
                className={`
                  relative inline-flex h-6 w-11 items-center rounded-full transition-colors
                  ${puzzle.allPurplesAchieved ? 'bg-purple-500' : 'bg-gray-300'}
                  ${toggling ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                `}
                title={puzzle.allPurplesAchieved ? 'Click to disable' : 'Click to enable'}
              >
                <span
                  className={`
                    inline-block h-4 w-4 transform rounded-full bg-white transition-transform
                    ${puzzle.allPurplesAchieved ? 'translate-x-6' : 'translate-x-1'}
                  `}
                />
              </button>
              <span className={`text-sm font-semibold ${puzzle.allPurplesAchieved ? 'text-purple-600' : 'text-gray-500'}`}>
                {puzzle.allPurplesAchieved ? 'Enabled' : 'Disabled'}
              </span>
            </div>
          </div>
          {/* Action Buttons */}
          <div className="flex justify-center gap-4">
            <button
              onClick={handleResetPuzzle}
              disabled={resetting}
              className="bg-yellow-500 hover:bg-yellow-600 disabled:bg-gray-400 text-white px-6 py-2 rounded-md font-medium"
            >
              {resetting ? 'Resetting...' : 'Reset to Base State'}
            </button>
            <button
              onClick={fetchPuzzle}
              className="bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-md font-medium"
            >
              Refresh State
            </button>
          </div>
        </div>
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
