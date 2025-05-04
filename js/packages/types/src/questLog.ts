import { PointOfInterest } from './pointOfInterest';
import { PointOfInterestChallenge } from './pointOfInterestChallenge';
import { PointOfInterestChallengeSubmission } from './pointOfInterestChallengeSubmission';
import { PointOfInterestGroup } from './pointOfInterestGroup';

export interface QuestObjective {
  challenge: PointOfInterestChallenge;
  isCompleted: boolean;
  submissions: PointOfInterestChallengeSubmission[];
  nextNode?: QuestNode | null;
}

export interface QuestNode {
  pointOfInterest: PointOfInterest;
  objectives: QuestObjective[];
  children: Record<string, QuestNode>;
}

export interface Quest {
  isCompleted: boolean;
  rootNode: QuestNode;
  imageUrl: string;
  name: string;
  description: string;
  id: string;
}

export interface QuestLog {
  quests: Quest[];
  pendingTasks: Record<string, Task[]>;
  completedTasks: Record<string, Task[]>;
  trackedQuestIds: string[];
}

export interface Task {
  challenge: PointOfInterestChallenge;
  questId: string;
}

// Helper function to recursively get tags from a quest node
function getTagsFromNode(node: QuestNode): string[] {
  const tags = new Set<string>();
  
  // Add tags from current node's point of interest
  node.pointOfInterest.tags.forEach(tag => tags.add(tag.name));
  
  // Recursively get tags from children
  node.objectives.forEach(objective => {
    if (objective.nextNode) {
      getTagsFromNode(objective.nextNode).forEach(tag => tags.add(tag));
    }
  });
  
  return Array.from(tags);
}

// Implementation of getTags for Quest
export function getQuestTags(quest: Quest): string[] {
  return getTagsFromNode(quest.rootNode);
}
