import type { ResourceGatherRequirement } from './resource';

export type ResourceType = {
  id: string;
  name: string;
  slug: string;
  description: string;
  mapIconUrl: string;
  mapIconPrompt: string;
  gatherRequirements?: ResourceGatherRequirement[];
  createdAt?: string;
  updatedAt?: string;
};
