import { useAPI } from '@poltergeist/contexts';
import { Survey, Submission, SubmissionAnswer } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';

interface ActivitySelection {
    [key: string]: boolean;
}

export const AnswerSurvey: React.FC = () => {
    const { id } = useParams();
    const { apiClient }= useAPI();
    const [survey, setSurvey] = useState<Survey | null>(null);
    const [submission, setSubmission] = useState<Submission | null>(null);
    const [answers, setAnswers] = useState<ActivitySelection>({});

    useEffect(() => {
        const fetchSurveyAndSubmission = async () => {
            try {
                const survey = await apiClient.get<Survey>(`/sonar/surveys/${id}`);
                setSurvey(survey);
                const initialAnswers = survey.activities.reduce((acc, activity) => {
                    acc[activity.id] = false;
                    return acc;
                }, {});
                setAnswers(initialAnswers);
                const submission = await apiClient.get<Submission>(`/sonar/surveys/${id}/submissions`);
                setSubmission(submission);
            } catch (error) {
                console.error("Failed to fetch survey or submission", error);
            }
        };

        fetchSurveyAndSubmission();
    }, [id]);

    if (!survey) {
        return <div>Loading...</div>;
    }

    return (
        <div>
            <h1>{survey.title}</h1>
            <div>
                {survey.activities.map((activity) => {
                    let matchingAnswer;

                    if (submission) {
                        matchingAnswer = submission.answers.find((answer) => answer.activityId === activity.id);
                    }
                    return (
                        <div>
                            <span><strong>{activity.title}</strong></span>
                            {matchingAnswer && <span>  {matchingAnswer.down ? "Yes" : "No"}</span>}
                            {!submission && (
                                    <input type="checkbox" onChange={(e) => {
                                        const updatedAnswers: ActivitySelection = { ...answers, [activity.id]: e.target.checked };
                                        setAnswers(updatedAnswers);
                                    }} />
                            )}
                        </div>
                    )
                })}
                {!submission && <div>
                    <button onClick={async () => {
                        const activityIds: string[] = [];
                        const downs: boolean[] = [];

                        Object.keys(answers).forEach((activityId) => {
                            activityIds.push(activityId);
                            downs.push(answers[activityId]);
                        });

                        try {
                            await apiClient.post(`/sonar/surveys/${survey.id}/submissions`, { activityIds, downs });
                            window.location.href = '/thanks';
                        } catch (e) {
                            alert(`Error!: ${e}`)
                        }
                    }}>Submit</button>
                </div>}
            </div>
        </div>
    );
}