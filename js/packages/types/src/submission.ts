import { SubmissionAnswer } from './submissionAnswer';
import { User } from './user';

export type Submission = {
  id: string;
  createdAt: string;
  updatedAt: string;
  surveyId: string;
  userId: string;
  user: User;
  answers: SubmissionAnswer[];
};
