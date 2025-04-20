export interface QuestArchetypeChallenge {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt?: Date;
    reward: number;
    unlockedNodeId?: string;
    unlockedNode?: QuestArchetypeNode;
}
export interface QuestArchetypeNode {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt?: Date;
    locationArchType: string;
    locationArchTypeId: string;
    challenges: QuestArchetypeChallenge[];
}
export interface QuestArchetype {
    id: string;
    name: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt?: Date;
    root: QuestArchetypeNode;
    rootId: string;
}
