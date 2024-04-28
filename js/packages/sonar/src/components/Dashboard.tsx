import { useAuth } from '@poltergeist/contexts';
import './Dashboard.css';
import React from 'react';
import { Button } from './shared/Button.tsx';
import { useNavigate } from 'react-router-dom';

type DashboardProps = {};

export function Dashboard(props: DashboardProps) {
  const navigate = useNavigate();

  return (
    <div className="Dashboard__background">
      <div className="Dashboard__modal">
        <h3>Welcome back! Choose your adventure:</h3>
        <Button title="Assemble crew" onClick={() => navigate('/answers')} />
        <Button
          title="Summon adventurers"
          onClick={() => navigate('/new-survey')}
        />
        <Button title="View roster" onClick={() => navigate('/surveys')} />
      </div>
    </div>
  );
}
