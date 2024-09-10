import React, { useEffect, useState } from 'react';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import Divider from './shared/Divider.tsx';
import TextInput from './shared/TextInput.tsx';
import { Button } from './shared/Button.tsx';
import { useAPI } from '@poltergeist/contexts';
import './Admin.css';

const Admin = () => {
  const { match, getCurrentMatch } = useMatchContext();
  const [teamId, setTeamId] = useState<string>('');
  const [pointOfInterestId, setPointOfInterestId] = useState<string>('');
  const [quantity, setQuantity] = useState<string>('');
  const { apiClient } = useAPI();

  useEffect(() => {
    getCurrentMatch();
  }, []);

  return (
    <div className="Admin__background">
      <h1 className="text-2xl font-bold">Admin</h1>
      <h2 className="text-xl font-bold">Teams</h2>
      <Divider />
      {match?.teams.map(team => <div key={team.id}>
        <p className="text-lg font-bold" onClick={() => setTeamId(team.id)}>{team.name}</p>
        <div>
          {team.users.map(user => <div key={user.id}>
            {user.name}
          </div>)}
        </div>
        <Divider />
      </div>)}
      <h2 className="text-xl font-bold">Points of Interest</h2>
      <Divider />
      {match?.pointsOfInterest.map(pointOfInterest => <div key={pointOfInterest.id}>
        <p className="text-lg font-bold" onClick={() => setPointOfInterestId(pointOfInterest.id)}>{pointOfInterest.name}</p>
        <Divider />
      </div>)}
      <h2 className="text-xl font-bold">Unlock point for team</h2>
      <TextInput label="Team ID" value={teamId} onChange={setTeamId} />
      <TextInput label="Point of Interest ID" value={pointOfInterestId} onChange={setPointOfInterestId} />
      <Button title="Unlock" disabled={!teamId || !pointOfInterestId} onClick={async () => {
          const submission = await apiClient.post(
            `/sonar/admin/pointOfInterest/unlock`,
            {
              teamId,
              pointOfInterestId,
            }
          );
          setTeamId('');
          setPointOfInterestId('');
      }} />
      <Divider />
      <h2 className="text-xl font-bold">Capture for team</h2>
      <TextInput label="Team ID" value={teamId} onChange={setTeamId} />
      <TextInput label="Point of Interest ID" value={pointOfInterestId} onChange={setPointOfInterestId} />
      <TextInput label="Quantity" value={quantity} onChange={setQuantity} />
      <Button title="Capture" disabled={!teamId || !pointOfInterestId || !quantity} onClick={async () => {
          const submission = await apiClient.post(
            `/sonar/admin/pointOfInterest/capture`,
            {
              teamId,
              pointOfInterestId,
              tier: JSON.parse(quantity),
            }
          );
          setTeamId('');
          setPointOfInterestId('');
          setQuantity('');
      }} />
    </div>
  );
};

export default Admin;
