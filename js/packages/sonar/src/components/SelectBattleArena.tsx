import React from 'react';
import './SelectBattleArena.css';
import { Button } from './shared/Button.tsx';
import { usePointOfInterestGroups } from '../hooks/usePointOfInterestGroups.ts';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { useNavigate } from 'react-router-dom';

export const SelectBattleArena = () => {
  const { pointOfInterestGroups, loading, error } = usePointOfInterestGroups();
  const { createMatch } = useMatchContext();
  const navigate = useNavigate();

  return (
    <div className="SelectBattleArena__background flex flex-col items-center gap-8">
      <h2>Choose your arena</h2>
      {pointOfInterestGroups &&
        pointOfInterestGroups.map((poiGroup) => (
          <div className="flex flex-col items-center gap-8" key={poiGroup.name}>
            <div>
              <div className="flex flex-col items-center gap-4">
                <h3 className="text-3xl font-bold">{poiGroup.name}</h3>
                <img src={poiGroup.imageUrl} alt={poiGroup.name} />
                <p>{poiGroup.description}</p>
                <Button
                  title="Choose"
                  disabled={loading}
                  onClick={() => {
                    createMatch(poiGroup.pointsOfInterest.map((poi) => poi.ID));
                    navigate('/match');
                  }}
                />
              </div>
            </div>
          </div>
        ))}
    </div>
  );
};
