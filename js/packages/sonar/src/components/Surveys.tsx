import React from 'react';
import './Surveys.css';
import { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Survey } from '@poltergeist/types';
import { Modal, ModalSize } from './shared/Modal.tsx';

export function Surveys() {
  const [surveys, setSurveys] = useState<Survey[]>([]);
  const context = useAPI();

  useEffect(() => {
    const fetchSurveys = async () => {
      try {
        const surveys = await context.apiClient.get<Survey[]>('/sonar/surveys');
        if (surveys) {
          setSurveys(surveys);
        }
      } catch (error) {
        console.error('Error fetching surveys:', error);
      }
    };
    fetchSurveys();
  }, []);

  return (
    <div className="Surveys__background">
      <Modal size={ModalSize.FULLSCREEN}>
        <div className="flex flex-col gap-4 w-full mt-4">
          <h2 className="text-2xl font-bold text-left">Sent summons</h2>
          <ul className="w-full">
            {surveys
              .sort(
                (a, b) =>
                  new Date(b.createdAt).getTime() -
                  new Date(a.createdAt).getTime()
              )
              .map((survey) => (
                <li
                  key={survey.id}
                  className="relative rounded-md p-3 text-sm/6 transition hover:bg-black/5 text-left"
                >
                  <a
                    href={`/surveys/${survey.id}`}
                    className="font-semibold text-black text-left"
                  >
                    <span className="absolute inset-0" />
                    {survey.title}
                  </a>
                  <ul className="flex gap-2 text-black/50" aria-hidden="true">
                    <li>{new Date(survey.createdAt).toLocaleDateString()}</li>
                    <li aria-hidden="true">&middot;</li>
                    <li>
                      {survey.surveySubmissions.length} submission
                      {survey.surveySubmissions.length === 1 ? '' : 's'}
                    </li>
                  </ul>
                </li>
              ))}
          </ul>
        </div>
      </Modal>
    </div>
  );
}
