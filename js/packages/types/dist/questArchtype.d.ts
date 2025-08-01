export interface QuestArchTypeChallenge {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt?: Date;
    reward: number;
    unlockedNodeId?: string;
    unlockedNode?: QuestArchTypeNode;
}
export interface QuestArchTypeNode {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt?: Date;
    locationArchType: string;
    locationArchTypeId: string;
    challenges: QuestArchTypeChallenge[];
}
export interface QuestArchType {
    id: string;
    name: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt?: Date;
    root: QuestArchTypeNode;
    rootId: string;
}
