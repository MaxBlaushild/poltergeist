import React, { useEffect } from 'react';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { Match } from './Match.tsx';
import { useNavigate, useParams } from 'react-router-dom';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

export const MatchById = () => {
  const { id: matchId } = useParams();
  const { match, getMatch, addUserToMatch } = useMatchContext();
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
    if (!match || !currentUser) {
      return;
    }

    const isAlreadyInMatch = match.teams.some(team => 
      team.users.some(user => user.id === currentUser.id)
    );
    if (isAlreadyInMatch) {
      return;
    }
    
    let isMounted = true;
    
    addUserToMatch(currentUser.id)
      .then(() => {
        if (isMounted) {
          navigate('/match');
        }
      })
      .catch((error) => {
        console.error('Failed to add user to match:', error);
      });

    return () => {
      isMounted = false;
    };
  }, [currentUser, match, navigate, addUserToMatch]);

  return <Match match={match} />;
};

