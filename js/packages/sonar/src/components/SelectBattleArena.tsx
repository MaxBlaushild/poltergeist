import React from 'react';
import './SelectBattleArena.css';
import { Button } from './shared/Button.tsx';
import { usePointOfInterestGroups } from '@poltergeist/hooks';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { useNavigate } from 'react-router-dom';

export const SelectBattleArena = () => {
  const { pointOfInterestGroups, loading, error } = usePointOfInterestGroups();
  const { createMatch } = useMatchContext();
  const navigate = useNavigate();

  return (
    <div className="SelectBattleArena__background flex flex-col items-center gap-8">
      <h2>Choose your arena</h2>
      <div className="flex overflow-x-auto snap-x snap-mandatory w-full">
        {pointOfInterestGroups &&
          pointOfInterestGroups.map((poiGroup) => (
            <div 
              className="flex-shrink-0 w-full snap-center px-4 overflow-y-auto overflow-x-hidden" 
              key={poiGroup.name}
            >
              <div className="flex flex-col items-center gap-8">
                <div className="flex flex-col items-center gap-4">
                  <h3 className="text-3xl font-bold">{poiGroup.name}</h3>
                  <img 
                    src={poiGroup.imageUrl} 
                    alt={poiGroup.name}
                    className="max-w-full h-auto" 
                  />
                  <p>{poiGroup.description}</p>
                  <Button
                    title="Choose"
                    disabled={loading}
                    onClick={() => {
                      createMatch(poiGroup.pointsOfInterest.map((poi) => poi.id));
                      navigate('/match');
                    }}
                  />
                </div>
              </div>
            </div>
          ))}
      </div>
    </div>
  );
};
