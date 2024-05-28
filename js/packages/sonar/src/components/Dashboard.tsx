import { useAuth } from '@poltergeist/contexts';
import './Dashboard.css';
import React from 'react';
import { Button } from './shared/Button.tsx';
import { useNavigate } from 'react-router-dom';
import { useSurveys } from '../hooks/useSurveys.ts';

type DashboardProps = {};

export function Dashboard(props: DashboardProps) {
  const navigate = useNavigate();
  const { surveys, isLoading } = useSurveys();

  return (
    <div className="Dashboard__background">
      <div className="Dashboard__modal">
        <h3>Welcome back! Choose your adventure:</h3>
        <Button
          title="Summon adventurers"
          onClick={() => navigate('/new-survey')}
        />
        <Button
          title="Assemble crew"
          disabled={isLoading || surveys.length === 0}
          onClick={() => navigate('/assemble-crew')}
        />
      </div>
    </div>
  );
}
