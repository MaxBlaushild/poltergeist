import React, { useEffect } from 'react';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { Match } from './Match.tsx';
import { useNavigate, useParams } from 'react-router-dom';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

export const MatchById = () => {
  const { id: matchId } = useParams();
  const { match, getMatch, createTeam, addUserToTeam } = useMatchContext();
  const { currentUser } = useUserProfiles();
  const navigate = useNavigate();
  const queryParams = new URLSearchParams(window.location.search);
  const teamId = queryParams.get('teamId');

  useEffect(() => {
    if (matchId) {
      getMatch(matchId);
    }
  }, [matchId, getMatch]);

  useEffect(() => {
    if (!match) {
      return;
    }

    if (!currentUser) {
      return;
    }

    const isAlreadyInMatch = match.teams.some(team => team.users.some(user => user.id === currentUser.id))
    if (isAlreadyInMatch) {
      return;
    }
    if (teamId) {
      addUserToTeam(teamId);
    } else {
      createTeam();
    }
    navigate('/match');
  }, [teamId, currentUser, match]);

  return <Match match={match} />;
};

