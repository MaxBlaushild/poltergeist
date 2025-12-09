import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAPI } from '@poltergeist/contexts';

export const UnlockRoom = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const { apiClient } = useAPI();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const unlockRoom = async () => {
      if (!roomId) {
        setError('Invalid room ID');
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        await apiClient.post(`/final-fete/rooms/${roomId}/unlock`, {});
        setSuccess(true);
        
        // Redirect to home after 2 seconds
        setTimeout(() => {
          navigate('/');
        }, 2000);
      } catch (err: any) {
        console.error('Error unlocking room:', err);
        const errorMessage = err.response?.data?.error || err.message || 'Failed to unlock room';
        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    unlockRoom();
  }, [roomId, apiClient, navigate]);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-4 md:p-6 text-[#00ff00]">
      {loading && (
        <div className="text-center">
          <div className="text-xl md:text-2xl font-bold mb-4 text-[#00ff00]">Unlocking room...</div>
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#00ff00] mx-auto" style={{ boxShadow: '0 0 20px #00ff00' }}></div>
        </div>
      )}

      {success && (
        <div className="text-center">
          <div className="text-xl md:text-2xl font-bold text-[#00ff00] mb-4" style={{ textShadow: '0 0 20px #00ff00' }}>Room unlocked successfully!</div>
          <p className="text-[#00ff00] opacity-80">Redirecting to home page...</p>
        </div>
      )}

      {error && (
        <div className="text-center max-w-md w-full px-4">
          <div className="text-xl md:text-2xl font-bold text-red-500 mb-4" style={{ textShadow: '0 0 20px #ff0000' }}>Error</div>
          <p className="text-[#00ff00] mb-4 opacity-80">{error}</p>
          <div className="flex flex-col md:flex-row gap-4 justify-center">
            <button
              onClick={() => navigate('/')}
              className="matrix-button matrix-button-secondary w-full md:w-auto min-h-[44px]"
            >
              Go Home
            </button>
            <button
              onClick={() => window.location.reload()}
              className="matrix-button matrix-button-primary w-full md:w-auto min-h-[44px]"
            >
              Try Again
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

