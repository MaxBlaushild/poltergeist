import { Activity } from './activity';
export type Category = {
    id: string;
    title: string;
    createdAt: Date;
    updatedAt: Date;
    activities: Activity[];
};
