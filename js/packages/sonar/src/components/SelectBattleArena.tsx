import React from 'react';
import './SelectBattleArena.css';
import { Button } from './shared/Button.tsx';
import { usePointOfInterestGroups } from '../hooks/usePointOfInterestGroups.ts';
import { useMatchContext } from '../contexts/MatchContext.tsx';

export const SelectBattleArena = () => {
  const { pointOfInterestGroups, loading, error } = usePointOfInterestGroups();
  const { createMatch } = useMatchContext();
  return (
    <div className="SelectBattleArena__background flex flex-col items-center gap-8">
      <h2>Choose your arena</h2>
      <div className="flex flex-col items-center gap-4">
        <h3 className="text-3xl font-bold">Harlem</h3>
        <img src="https://crew-points-of-interest.s3.amazonaws.com/harlem.webp" alt="Harlem" />
        <p>Once a cradle of culture, it now pulses with the en  ergy of pirates who’ve claimed its streets as their own. Here, the air is thick with the scent of soul food and the sound of jazz mixed with the clash of steel. It’s a place where fortunes are made in the shadows, and alliances shift like the tides. In Harlem, your name could be whispered in legend—or lost in the depths of the night. </p>
      </div>
      <Button 
        title="Choose"
        disabled={loading}
        onClick={() => {
          if (pointOfInterestGroups?.[0].pointsOfInterest) {
            createMatch(pointOfInterestGroups[0].pointsOfInterest.map((poi) => poi.id));
          }
        }} />
    </div>
  );
};

