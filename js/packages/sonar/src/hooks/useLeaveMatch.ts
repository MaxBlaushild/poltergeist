import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';

const useLeaveMatch = () => {
  const { apiClient } = useAPI();

  const leaveMatch = async (matchID: string) => {
    try {
      await apiClient.post(`/sonar/matches/${matchID}/leave`);
    } catch (error) {
      console.error('Failed to leave match', error);
    }
  };

  return { leaveMatch };
};

export default useLeaveMatch;
