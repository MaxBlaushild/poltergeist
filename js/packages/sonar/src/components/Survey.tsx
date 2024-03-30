import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Survey as SurveyType } from '@poltergeist/types';
import axios from 'axios';
import { useAPI } from '@poltergeist/contexts';

interface SurveyParams {
  id: string;
}

export const Survey: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [survey, setSurvey] = useState<SurveyType | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchSurvey = async () => {
      try {
        const survey = await apiClient.get<SurveyType>(`/sonar/surveys/${id}`);
        setSurvey(survey);
      } catch (err) {
        setError('Failed to fetch survey');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchSurvey();
  }, [id]);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>{error}</div>;
  if (!survey) return <div>Survey not found</div>;

  return (
    <div>
      <h1>{survey.title}</h1>
      <ul>
        {survey.activities.map((activity) => (
          <li key={activity.id}>{activity.title}</li>
        ))}
      </ul>
        <button onClick={() => navigator.clipboard.writeText(`${window.location.origin}/submit-answer/${id}`)}>
        Copy Share Link
        </button>
    </div>
  );
};
