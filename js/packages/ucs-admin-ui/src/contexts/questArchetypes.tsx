import React, { useState, useEffect, useContext, useCallback } from 'react';
import {
  DialogueMessage,
  QuestArchetype,
  QuestDifficultyMode,
  LocationArchetype,
  QuestArchetypeChallenge,
  QuestArchetypeNode,
  QuestArchetypeNodeType,
  ZoneQuestArchetype,
} from '@poltergeist/types';
import { useAPI, useAuth } from '@poltergeist/contexts';

export type QuestArchetypeNodeDraft = {
  nodeType?: QuestArchetypeNodeType;
  locationArchetypeId?: string | null;
  locationSelectionMode?: 'random' | 'closest';
  challengeTemplateId?: string | null;
  scenarioTemplateId?: string | null;
  monsterTemplateIds?: string[];
  targetLevel?: number | null;
  encounterProximityMeters?: number | null;
  expositionTitle?: string;
  expositionDescription?: string;
  expositionDialogue?: DialogueMessage[];
  expositionRewardMode?: 'explicit' | 'random';
  expositionRandomRewardSize?: 'small' | 'medium' | 'large';
  expositionRewardExperience?: number | null;
  expositionRewardGold?: number | null;
  expositionMaterialRewards?: { resourceKey: string; amount: number }[];
  expositionItemRewards?: { inventoryItemId: number; quantity: number }[];
  expositionSpellRewards?: { spellId: string }[];
};

export type QuestArchetypeDraft = {
  name: string;
  description: string;
  category?: 'side' | 'main_story';
  questGiverCharacterId?: string | null;
  acceptanceDialogue?: DialogueMessage[];
  imageUrl?: string;
  rootNode: QuestArchetypeNodeDraft;
  difficultyMode?: QuestDifficultyMode;
  difficulty?: number;
  monsterEncounterTargetLevel?: number;
  defaultGold?: number;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience?: number;
  recurrenceFrequency?: string | null;
  materialRewards?: { resourceKey: string; amount: number }[];
  requiredStoryFlags?: string[];
  setStoryFlags?: string[];
  clearStoryFlags?: string[];
  questGiverRelationshipEffects?: {
    trust?: number;
    respect?: number;
    fear?: number;
    debt?: number;
  };
  itemRewards?: { inventoryItemId: number; quantity: number }[];
  spellRewards?: { spellId: string }[];
  characterTags?: string[];
  internalTags?: string[];
};

export type QuestTemplateGeneratorStepDraft = {
  source: 'location_archetype' | 'proximity';
  content: 'challenge' | 'scenario' | 'monster';
  locationArchetypeId?: string | null;
  proximityMeters?: number | null;
};

export type QuestTemplateGeneratorDraft = {
  name?: string;
  themePrompt?: string;
  characterTags?: string[];
  internalTags?: string[];
  steps: QuestTemplateGeneratorStepDraft[];
};

type QuestArchetypesContextType = {
  placeTypes: string[];
  locationArchetypes: LocationArchetype[];
  questArchetypes: QuestArchetype[];
  zoneQuestArchetypes: ZoneQuestArchetype[];
  refreshQuestArchetypes: () => Promise<void>;
  refreshLocationArchetypes: () => Promise<void>;
  createQuestArchetype: (
    draft: QuestArchetypeDraft
  ) => Promise<QuestArchetype | null>;
  generateQuestArchetypeTemplate: (
    draft: QuestTemplateGeneratorDraft
  ) => Promise<QuestArchetype | null>;
  addChallengeToQuestArchetype: (
    questArchetypeId: string,
    proficiency?: string | null,
    unlockedNode?: QuestArchetypeNodeDraft | null,
    challengeTemplateId?: string | null
  ) => void;
  updateQuestArchetypeChallenge: (
    challengeId: string,
    updates: {
      proficiency?: string | null;
      challengeTemplateId?: string | null;
    }
  ) => void;
  deleteQuestArchetypeChallenge: (challengeId: string) => void;
  updateQuestArchetypeNode: (
    nodeId: string,
    updates: QuestArchetypeNodeDraft
  ) => void;
  createLocationArchetype: (locationArchetype: LocationArchetype) => void;
  updateLocationArchetype: (locationArchetype: LocationArchetype) => void;
  updateQuestArchetype: (questArchetype: QuestArchetype) => void;
  deleteQuestArchetype: (questArchetypeId: string) => Promise<void>;
  deleteLocationArchetype: (locationArchetypeId: string) => void;
  createZoneQuestArchetype: (
    zoneId: string,
    questArchetypeId: string,
    numberOfQuests: number,
    characterId?: string | null
  ) => void;
  updateZoneQuestArchetype: (
    zoneQuestArchetypeId: string,
    updates: { characterId?: string | null; numberOfQuests?: number }
  ) => void;
  deleteZoneQuestArchetype: (zoneQuestArchetypeId: string) => void;
};

export const QuestArchetypesContext =
  React.createContext<QuestArchetypesContextType>({
    questArchetypes: [],
    zoneQuestArchetypes: [],
    refreshQuestArchetypes: async () => {},
    refreshLocationArchetypes: async () => {},
    createQuestArchetype: async () => null,
    generateQuestArchetypeTemplate: async () => null,
    addChallengeToQuestArchetype: () => {},
    updateQuestArchetypeChallenge: () => {},
    deleteQuestArchetypeChallenge: () => {},
    updateQuestArchetypeNode: () => {},
    locationArchetypes: [],
    createLocationArchetype: () => {},
    updateLocationArchetype: () => {},
    placeTypes: [],
    updateQuestArchetype: () => {},
    deleteQuestArchetype: async () => {},
    deleteLocationArchetype: () => {},
    createZoneQuestArchetype: () => {},
    updateZoneQuestArchetype: () => {},
    deleteZoneQuestArchetype: () => {},
  });

export const QuestArchetypesProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [questArchetypes, setQuestArchetypes] = useState<QuestArchetype[]>([]);
  const [locationArchetypes, setLocationArchetypes] = useState<
    LocationArchetype[]
  >([]);
  const [placeTypes, setPlaceTypes] = useState<string[]>([]);
  const [zoneQuestArchetypes, setZoneQuestArchetypes] = useState<
    ZoneQuestArchetype[]
  >([]);
  const populateChallengesForNode = useCallback(
    async function populateQuestArchetypeNode(node: QuestArchetypeNode) {
      const challenges = await apiClient.get<QuestArchetypeChallenge[]>(
        `/sonar/questArchetypes/${node.id}/challenges`
      );
      node.challenges = challenges;
      await Promise.all(
        (node.challenges ?? []).map(async (challenge) => {
          if (challenge.unlockedNode) {
            await populateQuestArchetypeNode(challenge.unlockedNode);
          }
        })
      );

      return node;
    },
    [apiClient]
  );

  const fetchQuestArchetypes = useCallback(async () => {
    const questArchetypes = await apiClient.get<QuestArchetype[]>(
      '/sonar/questArchetypes'
    );
    const populatedQuestArchetypes = await Promise.all(
      questArchetypes.map(async (questArchetype) => {
        questArchetype.root = await populateChallengesForNode(
          questArchetype.root
        );
        return questArchetype;
      })
    );
    setQuestArchetypes(populatedQuestArchetypes);
  }, [apiClient, populateChallengesForNode]);

  const fetchLocationArchetypes = async () => {
    const locationArchetypes = await apiClient.get<LocationArchetype[]>(
      '/sonar/locationArchetypes'
    );
    setLocationArchetypes(locationArchetypes);
  };

  const fetchPlaceTypes = async () => {
    const placeTypes = await apiClient.get<string[]>('/sonar/placeTypes');
    setPlaceTypes(placeTypes);
  };

  const fetchZoneQuestArchetypes = async () => {
    const zoneQuestArchetypes = await apiClient.get<ZoneQuestArchetype[]>(
      '/sonar/zoneQuestArchetypes'
    );
    setZoneQuestArchetypes(zoneQuestArchetypes);
  };

  useEffect(() => {
    if (!user) {
      setQuestArchetypes([]);
      setLocationArchetypes([]);
      setPlaceTypes([]);
      setZoneQuestArchetypes([]);
      return;
    }

    const fetchData = async () => {
      await Promise.all([
        fetchQuestArchetypes(),
        fetchLocationArchetypes(),
        fetchPlaceTypes(),
        fetchZoneQuestArchetypes(),
      ]);
    };
    fetchData();
  }, [user]); // Remove function dependencies since they're defined in component scope

  const createLocationArchetype = async (
    locationArchetype: LocationArchetype
  ) => {
    const newLocationArchetype = await apiClient.post<LocationArchetype>(
      '/sonar/locationArchetypes',
      locationArchetype
    );
    setLocationArchetypes([...locationArchetypes, newLocationArchetype]);
  };

  const createQuestArchetype = async (
    draft: QuestArchetypeDraft
  ): Promise<QuestArchetype | null> => {
    const node = await apiClient.post<QuestArchetypeNode>(
      '/sonar/questArchetypeNodes',
      {
        nodeType: draft.rootNode.nodeType,
        locationArchetypeID: draft.rootNode.locationArchetypeId,
        locationSelectionMode: draft.rootNode.locationSelectionMode,
        challengeTemplateId: draft.rootNode.challengeTemplateId,
        scenarioTemplateId: draft.rootNode.scenarioTemplateId,
        monsterTemplateIds: draft.rootNode.monsterTemplateIds,
        targetLevel: draft.rootNode.targetLevel,
        encounterProximityMeters: draft.rootNode.encounterProximityMeters,
        expositionTitle: draft.rootNode.expositionTitle,
        expositionDescription: draft.rootNode.expositionDescription,
        expositionDialogue: draft.rootNode.expositionDialogue,
        expositionRewardMode: draft.rootNode.expositionRewardMode,
        expositionRandomRewardSize: draft.rootNode.expositionRandomRewardSize,
        expositionRewardExperience: draft.rootNode.expositionRewardExperience,
        expositionRewardGold: draft.rootNode.expositionRewardGold,
        expositionMaterialRewards: draft.rootNode.expositionMaterialRewards,
        expositionItemRewards: draft.rootNode.expositionItemRewards,
        expositionSpellRewards: draft.rootNode.expositionSpellRewards,
      }
    );
    const questArchetype = await apiClient.post<QuestArchetype>(
      '/sonar/questArchetypes',
      {
        name: draft.name,
        description: draft.description,
        category: draft.category,
        questGiverCharacterId: draft.questGiverCharacterId,
        acceptanceDialogue: draft.acceptanceDialogue,
        imageUrl: draft.imageUrl,
        rootId: node.id,
        difficultyMode: draft.difficultyMode,
        difficulty: draft.difficulty,
        monsterEncounterTargetLevel: draft.monsterEncounterTargetLevel,
        defaultGold: draft.defaultGold,
        rewardMode: draft.rewardMode,
        randomRewardSize: draft.randomRewardSize,
        rewardExperience: draft.rewardExperience,
        recurrenceFrequency: draft.recurrenceFrequency,
        materialRewards: draft.materialRewards,
        requiredStoryFlags: draft.requiredStoryFlags,
        setStoryFlags: draft.setStoryFlags,
        clearStoryFlags: draft.clearStoryFlags,
        questGiverRelationshipEffects: draft.questGiverRelationshipEffects,
        itemRewards: draft.itemRewards,
        spellRewards: draft.spellRewards,
        characterTags: draft.characterTags,
        internalTags: draft.internalTags,
      }
    );
    setQuestArchetypes([...questArchetypes, questArchetype]);
    return questArchetype;
  };

  const generateQuestArchetypeTemplate = async (
    draft: QuestTemplateGeneratorDraft
  ): Promise<QuestArchetype | null> => {
    const questArchetype = await apiClient.post<QuestArchetype>(
      '/sonar/questArchetypes/generate-template',
      {
        name: draft.name?.trim() || '',
        themePrompt: draft.themePrompt?.trim() || '',
        characterTags: draft.characterTags ?? [],
        internalTags: draft.internalTags ?? [],
        steps: draft.steps.map((step) => ({
          source: step.source,
          content: step.content,
          locationArchetypeId: step.locationArchetypeId || null,
          proximityMeters:
            step.source === 'proximity'
              ? Math.max(0, Number(step.proximityMeters) || 0)
              : null,
        })),
      }
    );
    if (questArchetype.root) {
      questArchetype.root = await populateChallengesForNode(
        questArchetype.root
      );
    }
    setQuestArchetypes((prev) => [
      ...prev.filter((entry) => entry.id !== questArchetype.id),
      questArchetype,
    ]);
    return questArchetype;
  };

  const updateLocationArchetype = async (
    locationArchetype: LocationArchetype
  ) => {
    const updatedLocationArchetype = await apiClient.patch<LocationArchetype>(
      `/sonar/locationArchetypes/${locationArchetype.id}`,
      locationArchetype
    );
    setLocationArchetypes(
      locationArchetypes.map((locationArchetype) =>
        locationArchetype.id === updatedLocationArchetype.id
          ? updatedLocationArchetype
          : locationArchetype
      )
    );
  };

  const updateQuestArchetype = async (questArchetype: QuestArchetype) => {
    const updatedQuestArchetype = await apiClient.patch<QuestArchetype>(
      `/sonar/questArchetypes/${questArchetype.id}`,
      questArchetype
    );
    setQuestArchetypes(
      questArchetypes.map((questArchetype) =>
        questArchetype.id === updatedQuestArchetype.id
          ? updatedQuestArchetype
          : questArchetype
      )
    );
  };

  const addChallengeToQuestArchetype = async (
    questArchetypeId: string,
    proficiency?: string | null,
    unlockedNode?: QuestArchetypeNodeDraft | null,
    challengeTemplateId?: string | null
  ) => {
    const payload: {
      proficiency?: string;
      challengeTemplateId?: string;
      nodeType?: QuestArchetypeNodeType;
      locationArchetypeID?: string;
      locationSelectionMode?: 'random' | 'closest';
      scenarioTemplateId?: string | null;
      monsterTemplateIds?: string[];
      targetLevel?: number | null;
      encounterProximityMeters?: number | null;
      expositionTitle?: string;
      expositionDescription?: string;
      expositionDialogue?: DialogueMessage[];
      expositionRewardMode?: 'explicit' | 'random';
      expositionRandomRewardSize?: 'small' | 'medium' | 'large';
      expositionRewardExperience?: number | null;
      expositionRewardGold?: number | null;
      expositionMaterialRewards?: { resourceKey: string; amount: number }[];
      expositionItemRewards?: { inventoryItemId: number; quantity: number }[];
      expositionSpellRewards?: { spellId: string }[];
    } = {};

    if (unlockedNode) {
      if (unlockedNode.nodeType) {
        payload.nodeType = unlockedNode.nodeType;
      }
      if (unlockedNode.locationArchetypeId) {
        payload.locationArchetypeID = unlockedNode.locationArchetypeId;
      }
      payload.locationSelectionMode = unlockedNode.locationSelectionMode;
      payload.challengeTemplateId = unlockedNode.challengeTemplateId ?? undefined;
      payload.scenarioTemplateId = unlockedNode.scenarioTemplateId;
      if (
        unlockedNode.monsterTemplateIds &&
        unlockedNode.monsterTemplateIds.length > 0
      ) {
        payload.monsterTemplateIds = unlockedNode.monsterTemplateIds;
      }
      payload.targetLevel = unlockedNode.targetLevel;
      payload.encounterProximityMeters = unlockedNode.encounterProximityMeters;
      payload.expositionTitle = unlockedNode.expositionTitle;
      payload.expositionDescription = unlockedNode.expositionDescription;
      payload.expositionDialogue = unlockedNode.expositionDialogue;
      payload.expositionRewardMode = unlockedNode.expositionRewardMode;
      payload.expositionRandomRewardSize =
        unlockedNode.expositionRandomRewardSize;
      payload.expositionRewardExperience =
        unlockedNode.expositionRewardExperience;
      payload.expositionRewardGold = unlockedNode.expositionRewardGold;
      payload.expositionMaterialRewards =
        unlockedNode.expositionMaterialRewards;
      payload.expositionItemRewards = unlockedNode.expositionItemRewards;
      payload.expositionSpellRewards = unlockedNode.expositionSpellRewards;
    }
    if (proficiency && proficiency.trim().length > 0) {
      payload.proficiency = proficiency.trim();
    }
    if (challengeTemplateId && challengeTemplateId.trim().length > 0) {
      payload.challengeTemplateId = challengeTemplateId.trim();
    }

    await apiClient.post<QuestArchetypeChallenge>(
      `/sonar/questArchetypes/${questArchetypeId}/challenges`,
      payload
    );

    fetchQuestArchetypes();
  };

  const updateQuestArchetypeChallenge = async (
    challengeId: string,
    updates: {
      proficiency?: string | null;
      challengeTemplateId?: string | null;
    }
  ) => {
    await apiClient.patch(
      `/sonar/questArchetypeChallenges/${challengeId}`,
      updates
    );
    fetchQuestArchetypes();
  };

  const deleteQuestArchetypeChallenge = async (challengeId: string) => {
    await apiClient.delete(`/sonar/questArchetypeChallenges/${challengeId}`);
    fetchQuestArchetypes();
  };

  const updateQuestArchetypeNode = async (
    nodeId: string,
    updates: QuestArchetypeNodeDraft
  ) => {
    await apiClient.patch(`/sonar/questArchetypeNodes/${nodeId}`, {
      nodeType: updates.nodeType,
      locationArchetypeID: updates.locationArchetypeId,
      locationSelectionMode: updates.locationSelectionMode,
      challengeTemplateId: updates.challengeTemplateId,
      scenarioTemplateId: updates.scenarioTemplateId,
      monsterTemplateIds: updates.monsterTemplateIds,
      targetLevel: updates.targetLevel,
      encounterProximityMeters: updates.encounterProximityMeters,
      expositionTitle: updates.expositionTitle,
      expositionDescription: updates.expositionDescription,
      expositionDialogue: updates.expositionDialogue,
      expositionRewardMode: updates.expositionRewardMode,
      expositionRandomRewardSize: updates.expositionRandomRewardSize,
      expositionRewardExperience: updates.expositionRewardExperience,
      expositionRewardGold: updates.expositionRewardGold,
      expositionMaterialRewards: updates.expositionMaterialRewards,
      expositionItemRewards: updates.expositionItemRewards,
      expositionSpellRewards: updates.expositionSpellRewards,
    });
    fetchQuestArchetypes();
  };

  const deleteQuestArchetype = async (questArchetypeId: string) => {
    await apiClient.delete<QuestArchetype>(
      `/sonar/questArchetypes/${questArchetypeId}`
    );
    setQuestArchetypes(
      questArchetypes.filter(
        (questArchetype) => questArchetype.id !== questArchetypeId
      )
    );
  };

  const deleteLocationArchetype = async (locationArchetypeId: string) => {
    await apiClient.delete<LocationArchetype>(
      `/sonar/locationArchetypes/${locationArchetypeId}`
    );
    setLocationArchetypes(
      locationArchetypes.filter(
        (locationArchetype) => locationArchetype.id !== locationArchetypeId
      )
    );
  };

  const createZoneQuestArchetype = async (
    zoneId: string,
    questArchetypeId: string,
    numberOfQuests: number,
    characterId?: string | null
  ) => {
    const created = await apiClient.post<ZoneQuestArchetype>(
      '/sonar/zoneQuestArchetypes',
      {
        zoneId,
        questArchetypeId,
        numberOfQuests,
        characterId,
      }
    );
    setZoneQuestArchetypes((prev) => [...prev, created]);
  };

  const updateZoneQuestArchetype = async (
    zoneQuestArchetypeId: string,
    updates: { characterId?: string | null; numberOfQuests?: number }
  ) => {
    const updated = await apiClient.patch<ZoneQuestArchetype>(
      `/sonar/zoneQuestArchetypes/${zoneQuestArchetypeId}`,
      updates
    );
    setZoneQuestArchetypes((prev) =>
      prev.map((zoneQuestArchetype) =>
        zoneQuestArchetype.id === updated.id ? updated : zoneQuestArchetype
      )
    );
  };

  const deleteZoneQuestArchetype = async (zoneQuestArchetypeId: string) => {
    await apiClient.delete<ZoneQuestArchetype>(
      `/sonar/zoneQuestArchetypes/${zoneQuestArchetypeId}`
    );
    setZoneQuestArchetypes(
      zoneQuestArchetypes.filter(
        (zoneQuestArchetype) => zoneQuestArchetype.id !== zoneQuestArchetypeId
      )
    );
  };

  return (
    <QuestArchetypesContext.Provider
      value={{
        questArchetypes,
        locationArchetypes,
        placeTypes,
        zoneQuestArchetypes,
        refreshQuestArchetypes: fetchQuestArchetypes,
        refreshLocationArchetypes: fetchLocationArchetypes,
        createQuestArchetype,
        generateQuestArchetypeTemplate,
        addChallengeToQuestArchetype,
        createLocationArchetype,
        updateLocationArchetype,
        updateQuestArchetype,
        deleteQuestArchetype,
        deleteLocationArchetype,
        createZoneQuestArchetype,
        updateZoneQuestArchetype,
        deleteZoneQuestArchetype,
        updateQuestArchetypeChallenge,
        deleteQuestArchetypeChallenge,
        updateQuestArchetypeNode,
      }}
    >
      {children}
    </QuestArchetypesContext.Provider>
  );
};

export const useQuestArchetypes = () => {
  return useContext(QuestArchetypesContext);
};
