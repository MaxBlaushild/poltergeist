import { useAuth } from '@poltergeist/contexts';
import './Dashboard.css';
import React, { useEffect } from 'react';
import { Button } from './shared/Button.tsx';
import { useNavigate } from 'react-router-dom';
import { useSurveys } from '../hooks/useSurveys.ts';
import { useMatchContext } from '../contexts/MatchContext.tsx';

type DashboardProps = {};

export function Dashboard(props: DashboardProps) {
  const navigate = useNavigate();
  const { surveys, isLoading } = useSurveys();
  const { match, isCurrentMatchLoading, getCurrentMatch } = useMatchContext();

  useEffect(() => {
    getCurrentMatch();
  }, []);

  return (
    <div className="Dashboard__background">
      <div className="Dashboard__modal">
        <h3>Welcome back! Choose your adventure:</h3>
        {match && (
          <Button
            title="Continue"
            onClick={() => navigate('/match')}
          />
        )}
        <Button
          title="Single player"
          disabled={true}
          onClick={() => navigate('/new-survey')}
        />
        <Button
          title="Co-op"
          disabled={true}
          onClick={() => navigate('/new-survey')}
        />
        {!isCurrentMatchLoading && !match && (
          <Button
            title="Battle mode"
            onClick={() => navigate('/select-battle-arena')}
        />)}
        <Button
          title="Invite new crew members"
          onClick={() => navigate('/new-survey')}
        />
        <Button
          title="View crew manifest"
          disabled={isLoading || surveys.length === 0}
          onClick={() => navigate('/assemble-crew')}
        />
      </div>
    </div>
  );
}
