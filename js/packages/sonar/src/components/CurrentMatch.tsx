import React, { useEffect } from 'react';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { Match } from './Match.tsx';

export const CurrentMatch = () => {
  const { match, getCurrentMatch } = useMatchContext();

  useEffect(() => {
    getCurrentMatch();
  }, [getCurrentMatch]);

  return <Match match={match} />;
};

