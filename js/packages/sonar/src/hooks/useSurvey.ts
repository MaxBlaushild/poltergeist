import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Survey, Submission } from '@poltergeist/types';

export interface UseSurveyResult {
  survey: Survey | null;
  submission: Submission | null;
  surveyLoading: boolean;
  surveyError: Error | null;
  submissionLoading: boolean;
  submissionError: Error | null;
}

export const useSurvey = (id: string): UseSurveyResult => {
  const { apiClient } = useAPI();
  const [survey, setSurvey] = useState<Survey | null>(null);
  const [submission, setSubmission] = useState<Submission | null>(null);
  const [surveyLoading, setSurveyLoading] = useState<boolean>(true);
  const [surveyError, setSurveyError] = useState<Error | null>(null);
  const [submissionLoading, setSubmissionLoading] = useState<boolean>(true);
  const [submissionError, setSubmissionError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchSurvey = async () => {
      try {
        const fetchedSurvey = await apiClient.get<Survey>(
          `/sonar/surveys/${id}`
        );
        setSurvey(fetchedSurvey);
      } catch (error) {
        setSurveyError(error);
      } finally {
        setSurveyLoading(false);
      }
    };

    const fetchSubmission = async () => {
      try {
        const fetchedSubmission = await apiClient.get<Submission>(
          `/sonar/surveys/${id}/submissions`
        );
        setSubmission(fetchedSubmission);
      } catch (error) {
        setSubmissionError(error);
      } finally {
        setSubmissionLoading(false);
      }
    };

    fetchSurvey();
    fetchSubmission();
  }, [id, apiClient]);

  return {
    survey,
    submission,
    surveyLoading,
    surveyError,
    submissionLoading,
    submissionError,
  };
};
