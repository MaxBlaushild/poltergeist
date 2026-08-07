import type { GMGame } from '../../gmApi';
import { formatClock } from '../../theme';

// Shared between the Setup tab's schedule editor and the Scoring tab's
// result recorder — both show a game's time/location/notes as context.
export const medal = ['🥇', '🥈', '🥉'];

export const ScheduleLine = ({ game }: { game: GMGame }) => {
  const scheduled = game.startMinutes != null && game.endMinutes != null;
  if (!scheduled && !game.assignedGm && !game.runNotes) return null;
  return (
    <div className="mb-2">
      {(scheduled || game.location || game.assignedGm) && (
        <p className="text-xs text-bone/60">
          {scheduled && (
            <span>
              🕒 {formatClock(game.startMinutes!)}–{formatClock(game.endMinutes!)}{' '}
            </span>
          )}
          {game.location && <span className="text-gold/80">· 📍 {game.location} </span>}
          {game.assignedGm && <span className="text-bone/50">· 👤 {game.assignedGm}</span>}
        </p>
      )}
      {game.runNotes && (
        <details className="mt-1">
          <summary className="text-xs text-gold/80 cursor-pointer">▸ How to run</summary>
          <p className="mt-1 text-sm text-bone/80 whitespace-pre-wrap rounded-md bg-black/30 p-2">
            {game.runNotes}
          </p>
        </details>
      )}
    </div>
  );
};
