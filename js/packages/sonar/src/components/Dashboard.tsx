import { useAuth } from '@poltergeist/contexts';
import './Dashboard.css';
import React, { useEffect } from 'react';
import { Button } from './shared/Button.tsx';
import { useNavigate } from 'react-router-dom';
import { useSurveys } from '../hooks/useSurveys.ts';
import useHasCurrentMatch from '../hooks/useHasCurrentMatch.ts';

type DashboardProps = {};

export function Dashboard(props: DashboardProps) {
  const navigate = useNavigate();
  const { surveys, isLoading } = useSurveys();
  const { hasCurrentMatch } = useHasCurrentMatch();

  return (
    <div className="Dashboard__background">
      <div className="Dashboard__modal">
        <h3>Welcome back! Choose your adventure:</h3>
        {hasCurrentMatch && (
          <Button
            title="Continue"
            onClick={() => navigate('/match/lobby')}
          />
        )}
        {!hasCurrentMatch && <Button
          title="Single player"
          onClick={() => navigate('/single-player')}
        />}

        {!hasCurrentMatch && (
          <Button
            title="Battle mode"
            onClick={() => navigate('/select-battle-arena')}
        />)}
        {/* <Button
          title="Invite new crew members"
          onClick={() => navigate('/new-survey')}
        />
        <Button
          title="View crew manifest"
          disabled={isLoading || surveys.length === 0}
          onClick={() => navigate('/assemble-crew')}
        /> */}
      </div>
    </div>
  );
}
