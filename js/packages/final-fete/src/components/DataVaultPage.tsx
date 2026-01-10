import { useAPI } from '@poltergeist/contexts';
import type { UtilityClosetPuzzle } from '@poltergeist/types';
import { useEffect, useState } from 'react';

const solved = (puzzle: UtilityClosetPuzzle | null) => {
  if (!puzzle) return false;
  const hues = [
    puzzle.button0CurrentHue,
    puzzle.button1CurrentHue,
    puzzle.button2CurrentHue,
    puzzle.button3CurrentHue,
    puzzle.button4CurrentHue,
    puzzle.button5CurrentHue,
  ];
  return hues.every((hue) => hue === 6);
};

export const DataVaultPage = () => {
  const { apiClient } = useAPI();
  const [puzzle, setPuzzle] = useState<UtilityClosetPuzzle | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [resetting, setResetting] = useState(false);

  useEffect(() => {
    const fetchPuzzle = async () => {
      try {
        setLoading(true);
        const response = await apiClient.get<UtilityClosetPuzzle>('/final-fete/utility-closet-puzzle');
        setPuzzle(response);
        setError(null);
      } catch (err) {
        console.error('Error fetching puzzle state:', err);
        setError('Failed to load data vault status. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchPuzzle();
  }, [apiClient]);

  const handleReset = async () => {
    try {
      setResetting(true);
      setError(null);
      await apiClient.post('/final-fete/utility-closet-puzzle/reset', {});
      // Refresh puzzle state after reset
      const response = await apiClient.get<UtilityClosetPuzzle>('/final-fete/utility-closet-puzzle');
      setPuzzle(response);
    } catch (err) {
      console.error('Error resetting puzzle:', err);
      setError('Failed to reset puzzle. Please try again.');
    } finally {
      setResetting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6">
        <div className="bg-black/80 border border-[#00ff00]/40 rounded-lg p-6 text-center text-[#00ff00] shadow-[0_0_20px_rgba(0,255,0,0.3)]">
          Checking data vault status...
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6">
        <div className="bg-black/80 border border-red-500/60 rounded-lg p-6 text-center text-red-200 shadow-[0_0_20px_rgba(255,0,0,0.3)] max-w-md">
          <h1 className="text-2xl font-bold mb-3 text-red-400">Error</h1>
          <p className="text-sm text-red-200">{error}</p>
        </div>
      </div>
    );
  }

  const puzzleSolved = solved(puzzle);

  if (!puzzleSolved) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6">
        <div className="bg-black/85 border border-red-500/70 rounded-xl p-6 md:p-8 text-center shadow-[0_0_30px_rgba(255,0,0,0.35)] max-w-xl">
          <h1 className="text-2xl md:text-3xl font-bold mb-4 text-red-400">Access Denied</h1>
          <p className="text-sm md:text-base text-red-200 leading-relaxed">
            The virus must be purged from the servers before the Data Vault can be accessed.
            Return when all servers are green to access.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="bg-black/85 border border-[#00ff00]/60 rounded-xl p-6 md:p-8 text-center shadow-[0_0_30px_rgba(0,255,0,0.3)] max-w-xl">
        <h1 className="text-2xl md:text-3xl font-bold mb-4 text-[#00ff00]">Data Vault Unlocked</h1>
        <p className="text-sm md:text-base text-[#b5ffb5] leading-relaxed mb-6">
          The combination for the server key is <span className="font-mono font-bold text-white">iloveu</span>.
        </p>
        <button
          onClick={handleReset}
          disabled={resetting}
          className="bg-[#00ff00]/20 hover:bg-[#00ff00]/30 border-2 border-[#00ff00] text-[#00ff00] font-bold py-3 px-6 rounded-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed hover:shadow-[0_0_20px_rgba(0,255,0,0.5)]"
        >
          {resetting ? 'Resetting...' : 'I have the goods!'}
        </button>
      </div>
    </div>
  );
};


