import { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { User } from '@poltergeist/types';


export const useUsers = () => {
  const { apiClient } = useAPI();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await apiClient.get<User[]>('/sonar/users');
        setUsers(response);
        setLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch users'));
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  return { users, loading, error };
};

