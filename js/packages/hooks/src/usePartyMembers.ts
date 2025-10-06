import { getCityFromCoordinates } from '@poltergeist/utils';
import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { User } from '@poltergeist/types';

interface UseCityNameResult {
  partyMembers: User[];
  loading: boolean;
  error: Error | null;
}

export const usePartyMembers = (): UseCityNameResult => {
  const [partyMembers, setPartyMembers] = useState<User[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchPartyMembers = async () => {
      try {
        const members = await apiClient.get<User[]>(
          `/sonar/party/members`
        );
        setPartyMembers(members);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchPartyMembers();
  }, []);

  return {
    partyMembers,
    loading,
    error,
  };
};

export const joinParty = async (inviterID: string): Promise<void> => {
  const { apiClient } = useAPI();

  try {
    const response = await apiClient.post(`/sonar/party/join`, { inviterID });
  } catch (error) {
    console.error(error);
  }
};
