import { Submission } from "./submission";
import { Activity } from "./activity";

export type SubmissionAnswer = {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    surveyId: string;
    submissionId: string;
    submission: Submission;
    activityId: string;
    activity: Activity;
    down: boolean;
    notes: string;
};
