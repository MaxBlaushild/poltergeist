import { InventoryItem } from "./inventoryItem";

export type SubmissionResult = {
	successful: boolean;
	reason: string;
	questCompleted: boolean;
	itemsAwarded: InventoryItem[];
	experienceAwarded: number;
	reputationAwarded: number;
	zoneID: string;
	levelUp: boolean;
	reputationUp: boolean;
};