import React, { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Survey, Submission, SubmissionAnswer, User } from '@poltergeist/types';

type SelectedAnswer = {
  id: string;
  phoneNumber: string;
  name: string;
};

export const Answers: React.FC = () => {
  const { apiClient } = useAPI();
  const [submissions, setSubmissions] = useState<Submission[]>([]);
  const [nameFilters, setNameFilters] = useState<string[]>([]);
  const [activityFilters, setActivityFilters] = useState<string[]>([]);
  const [tempNameFilter, setTempNameFilter] = useState<string>('');
  const [tempActivityFilter, setTempActivityFilter] = useState<string>('');

  useEffect(() => {
    const fetchSurveysAndAnswers = async () => {
      try {
        const submissions = await apiClient.get<Submission[]>(
          '/sonar/surveys/submissions'
        );
        setSubmissions(submissions);
      } catch (error) {
        console.error('Failed to fetch surveys and answers', error);
      }
    };

    fetchSurveysAndAnswers();
  }, [apiClient]);

  const handleAddNameFilter = (): void => {
    if (tempNameFilter && !nameFilters.includes(tempNameFilter)) {
      setNameFilters([...nameFilters, tempNameFilter]);
      setTempNameFilter('');
    }
  };

  const handleAddActivityFilter = (): void => {
    if (tempActivityFilter && !activityFilters.includes(tempActivityFilter)) {
      setActivityFilters([...activityFilters, tempActivityFilter]);
      setTempActivityFilter('');
    }
  };

  const handleRemoveNameFilter = (filterToRemove: string): void => {
    setNameFilters(nameFilters.filter((filter) => filter !== filterToRemove));
  };

  const handleRemoveActivityFilter = (filterToRemove: string): void => {
    setActivityFilters(
      activityFilters.filter((filter) => filter !== filterToRemove)
    );
  };

  const answers = {};
  const dupeMap = {};
  submissions.forEach((submission) =>
    submission.answers.forEach((answer) => {
      if (
        nameFilters.some((filter) =>
          submission.user.name.toLowerCase().includes(filter.toLowerCase())
        ) ||
        activityFilters.some((filter) =>
          answer.activity.title.toLowerCase().includes(filter.toLowerCase())
        )
      ) {
        if (
          dupeMap[answer.activityId] &&
          dupeMap[answer.activityId][submission.userId]
        ) {
        } else {
          answers[answer.id] = true;

          if (!dupeMap[answer.activityId]) {
            dupeMap[answer.activityId] = { [submission.userId]: true };
          } else {
            dupeMap[answer.activityId][submission.userId] = true;
          }
        }
      }
    })
  );

  return (
    <div>
      <h1>Your Survey Answers</h1>
      <div>
        <input
          type="text"
          value={tempNameFilter}
          onChange={(e) => setTempNameFilter(e.target.value)}
          placeholder="Filter by name"
        />
        <button onClick={handleAddNameFilter}>Add Name Filter</button>
        {nameFilters.map((filter, index) => (
          <div key={index}>
            {filter}{' '}
            <button onClick={() => handleRemoveNameFilter(filter)}>
              Remove
            </button>
          </div>
        ))}
      </div>
      <div>
        <input
          type="text"
          value={tempActivityFilter}
          onChange={(e) => setTempActivityFilter(e.target.value)}
          placeholder="Filter by activity"
        />
        <button onClick={handleAddActivityFilter}>Add Activity Filter</button>
        {activityFilters.map((filter, index) => (
          <div key={index}>
            {filter}{' '}
            <button onClick={() => handleRemoveActivityFilter(filter)}>
              Remove
            </button>
          </div>
        ))}
      </div>
      {submissions.length > 0 ? (
        <ul>
          {submissions.map((submission, index) =>
            submission.answers
              .filter(
                (answer) =>
                  answer.down &&
                  (![...nameFilters, ...activityFilters].length ||
                    answers[answer.id])
              )
              .map((answer, index) => (
                <li key={index}>
                  <strong>User:</strong> {submission.user.name}{' '}
                  <strong>Activity:</strong> {answer.activity.title}
                </li>
              ))
          )}
        </ul>
      ) : (
        <p>No answers found.</p>
      )}
    </div>
  );
};
