import { Activity } from './activity';
import { User } from './user';
import { Submission } from './submission';
export type Survey = {
    id: string;
    title: string;
    createdAt: string;
    updatedAt: string;
    referrerId: string;
    progenitorId: string;
    activities: Activity[];
    user: User;
    surveySubmissions: Submission[];
};
