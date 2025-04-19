export interface QuestArchTypeChallenge {
    reward: number;
    unlockedNode?: QuestArchTypeNode;
}
export interface QuestArchTypeNode {
    locationArchType: string;
    challenges: QuestArchTypeChallenge[];
}
export interface QuestArchType {
    id: string;
    root: QuestArchTypeNode;
}
