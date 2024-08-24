import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { Scoreboard } from './Scoreboard.tsx';


export const TeamScore = () => {
  const { id: matchId } = useParams();
  const { match, getMatch, createTeam, addUserToTeam } = useMatchContext();
  const { currentUser } = useUserProfiles();
  const navigate = useNavigate();
  const queryParams = new URLSearchParams(window.location.search);
  const teamId = queryParams.get('teamId');
  const team = match?.teams.find((t) => t.id === teamId);

  return <div className="Team__score">
    <h1>{team?.name}</h1>
    <Scoreboard />
  </div>;
}

