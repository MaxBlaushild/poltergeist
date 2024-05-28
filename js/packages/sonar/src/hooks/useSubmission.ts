import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Submission } from '@poltergeist/types';

const useSubmission = (submissionId: string) => {
  const [submission, setSubmission] = useState<Submission | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchSubmission = async () => {
      try {
        const response = await apiClient.get<Submission>(
          `/sonar/submissions/${submissionId}`
        );
        setSubmission(response);
      } catch (err) {
        setError('Failed to fetch submission');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    if (submissionId) {
      fetchSubmission();
    }
  }, [submissionId, apiClient]);

  return { submission, loading, error };
};

export default useSubmission;
