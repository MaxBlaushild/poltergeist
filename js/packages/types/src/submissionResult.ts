export type SubmissionResult = {
	successful: boolean;
	reason: string;
	questCompleted: boolean;
	experienceAwarded?: number;
	reputationAwarded?: number;
	itemsAwarded?: Array<{ id: number; name: string; imageUrl: string }>;
    goldAwarded?: number;
	zoneID?: string;
	levelUp?: boolean;
	reputationUp?: boolean;
};