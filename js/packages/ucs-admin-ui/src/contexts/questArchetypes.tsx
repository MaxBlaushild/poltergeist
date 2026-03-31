import React, { useState, useEffect, useContext } from 'react';
import {
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
  scenarioTemplateId?: string | null;
  monsterTemplateIds?: string[];
  targetLevel?: number | null;
  encounterProximityMeters?: number | null;
};

export type QuestArchetypeDraft = {
  name: string;
  description: string;
  acceptanceDialogue?: string[];
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
  deleteQuestArchetype: (questArchetypeId: string) => void;
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
    deleteQuestArchetype: () => {},
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
  const populateChallengesForNode = async (node: QuestArchetypeNode) => {
    const challenges = await apiClient.get<QuestArchetypeChallenge[]>(
      `/sonar/questArchetypes/${node.id}/challenges`
    );
    node.challenges = challenges;
    node.challenges?.forEach(async (challenge) => {
      if (challenge.unlockedNode) {
        await populateChallengesForNode(challenge.unlockedNode);
      }
    });

    return node;
  };

  const fetchQuestArchetypes = async () => {
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
  };

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
        scenarioTemplateId: draft.rootNode.scenarioTemplateId,
        monsterTemplateIds: draft.rootNode.monsterTemplateIds,
        targetLevel: draft.rootNode.targetLevel,
        encounterProximityMeters: draft.rootNode.encounterProximityMeters,
      }
    );
    const questArchetype = await apiClient.post<QuestArchetype>(
      '/sonar/questArchetypes',
      {
        name: draft.name,
        description: draft.description,
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
      scenarioTemplateId?: string | null;
      monsterTemplateIds?: string[];
      targetLevel?: number | null;
      encounterProximityMeters?: number | null;
    } = {};

    if (unlockedNode) {
      if (unlockedNode.nodeType) {
        payload.nodeType = unlockedNode.nodeType;
      }
      if (unlockedNode.locationArchetypeId) {
        payload.locationArchetypeID = unlockedNode.locationArchetypeId;
      }
      payload.scenarioTemplateId = unlockedNode.scenarioTemplateId;
      if (
        unlockedNode.monsterTemplateIds &&
        unlockedNode.monsterTemplateIds.length > 0
      ) {
        payload.monsterTemplateIds = unlockedNode.monsterTemplateIds;
      }
      payload.targetLevel = unlockedNode.targetLevel;
      payload.encounterProximityMeters = unlockedNode.encounterProximityMeters;
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
      scenarioTemplateId: updates.scenarioTemplateId,
      monsterTemplateIds: updates.monsterTemplateIds,
      targetLevel: updates.targetLevel,
      encounterProximityMeters: updates.encounterProximityMeters,
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
