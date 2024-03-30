import React from 'react';
import { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Survey } from '@poltergeist/types';

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
            <div>
                <h2>Surveys</h2>
                <ul>
                    {surveys.map(survey => (
                        <li key={survey.id}>
                            <a href={`/surveys/${survey.id}`}>
                                <h3>{survey.title}</h3>
                                <small>Created at: {new Date(survey.createdAt).toLocaleDateString()}</small>
                            </a>
                        </li>
                    ))}
                </ul>
            </div>
        );
    }