import { Activity } from "./activity";

export type Survey = {
    id: string;
    title: string;
    createdAt: string;
    updatedAt: string;
    referrerId: string;
    progenitorId: string;
    activities: Activity[];
};