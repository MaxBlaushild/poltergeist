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
      <Button title="Free Play" onClick={() => {}}   />
      <Button title="Start" onClick={() => {}}   />
    </div>
  );
};
