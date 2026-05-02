export type RewardProfile = {
    id: string;
    createdAt: string;
    updatedAt: string;
    slug: string;
    name: string;
    description: string;
    active: boolean;
    preferredItemTags: string[];
    preferredMaterialKeys: string[];
    preferredDamageAffinities: string[];
    preferredResourceTypeIds: string[];
    preferEquipment: boolean;
    preferUtility: boolean;
    preferKnowledge: boolean;
    preferNonEquipment: boolean;
};
