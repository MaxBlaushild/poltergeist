import React, { useState } from 'react';
import { LameActivitySelector } from './shared/LameActivitySelector.tsx';
import { Button } from './shared/Button.tsx';
import './SinglePlayerMenu.css';

export const SinglePlayerMenu = () => {
  const [selectedActivityIds, setSelectedActivityIds] = useState<string[]>([]);
  const [shouldSelectActivities, setShouldSelectActivities] = useState(false);
  const handleActivitySelect = (activityId: string) => {
    setSelectedActivityIds([...selectedActivityIds, activityId]);
  };
  return (
    <div className="SinglePlayerMenu">
      <h2>Select Activities you want to do</h2>
      <label>
        Customize?
        <input type="checkbox" checked={shouldSelectActivities} onChange={() => setShouldSelectActivities(!shouldSelectActivities)}  />
      </label>
      {shouldSelectActivities && <LameActivitySelector
        selectedActivityIds={selectedActivityIds}
        onSelect={handleActivitySelect}
      />}
      <Button title="Start" onClick={() => {}}   />
    </div>
  );
};
