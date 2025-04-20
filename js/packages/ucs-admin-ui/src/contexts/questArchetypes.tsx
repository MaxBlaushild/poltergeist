import React, { useState, useEffect, useContext } from "react";
import { QuestArchetype, LocationArchetype, QuestArchetypeChallenge, QuestArchetypeNode } from "@poltergeist/types";
import { useAPI } from "@poltergeist/contexts";

type QuestArchetypesContextType = {
  placeTypes: string[];
  locationArchetypes: LocationArchetype[];
  questArchetypes: QuestArchetype[];
  createQuestArchetype: (questArchetype: QuestArchetype) => void;
  addChallengeToQuestArchetype: (questArchetypeId: string, challenge: QuestArchetypeChallenge) => void;
  createLocationArchetype: (locationArchetype: LocationArchetype) => void;
  updateLocationArchetype: (locationArchetype: LocationArchetype) => void;
  updateQuestArchetype: (questArchetype: QuestArchetype) => void;
  deleteQuestArchetype: (questArchetypeId: string) => void;
  deleteLocationArchetype: (locationArchetypeId: string) => void;
};

export const QuestArchetypesContext = React.createContext<QuestArchetypesContextType>({
  questArchetypes: [],
  createQuestArchetype: () => {},
  addChallengeToQuestArchetype: () => {},
  locationArchetypes: [],
  createLocationArchetype: () => {},
  updateLocationArchetype: () => {},
  placeTypes: [],
  updateQuestArchetype: () => {},
  deleteQuestArchetype: () => {},
  deleteLocationArchetype: () => {},
});

export const QuestArchetypesProvider = ({ children }: { children: React.ReactNode }) => {
  const { apiClient } = useAPI();
  const [questArchetypes, setQuestArchetypes] = useState<QuestArchetype[]>([]);
  const [locationArchetypes, setLocationArchetypes] = useState<LocationArchetype[]>([]);
  const [placeTypes, setPlaceTypes] = useState<string[]>([]);

  const populateChallengesForNode = async (node: QuestArchetypeNode) => {
    const challenges = await apiClient.get<QuestArchetypeChallenge[]>(`/sonar/questArchetypes/${node.id}/challenges`);
    node.challenges = challenges;

    node.challenges.forEach(async (challenge) => {
      if (challenge.unlockedNode) {
        await populateChallengesForNode(challenge.unlockedNode);
      }
    });

    return node;
  };

  useEffect(() => {
    const fetchQuestArchetypes = async () => {
      const questArchetypes = await apiClient.get<QuestArchetype[]>("/sonar/questArchetypes");
      questArchetypes.forEach(async (questArchetype) => {
        questArchetype.root = await populateChallengesForNode(questArchetype.root);
      });
      setQuestArchetypes(questArchetypes);
    };

    const fetchLocationArchetypes = async () => {
      const locationArchetypes = await apiClient.get<LocationArchetype[]>("/sonar/locationArchetypes");
      setLocationArchetypes(locationArchetypes);
    };

    const fetchPlaceTypes = async () => {
      const placeTypes = await apiClient.get<string[]>("/sonar/placeTypes");
      setPlaceTypes(placeTypes);
    };

    fetchQuestArchetypes();
    fetchLocationArchetypes();
    fetchPlaceTypes();
  }, []);

  const createLocationArchetype = async (locationArchetype: LocationArchetype) => {
    const newLocationArchetype = await apiClient.post<LocationArchetype>("/sonar/locationArchetypes", locationArchetype);
    setLocationArchetypes([...locationArchetypes, newLocationArchetype]);
  };

  const createQuestArchetype = async (questArchetype: QuestArchetype) => {
    const node = await apiClient.post<QuestArchetypeNode>("/sonar/questArchetypeNodes", questArchetype.root);
    questArchetype.root.id = node.id;
    const newQuestArchetype = await apiClient.post<QuestArchetype>("/sonar/questArchetypes", questArchetype);
    setQuestArchetypes([...questArchetypes, newQuestArchetype]);
  };

  const updateLocationArchetype = async (locationArchetype: LocationArchetype) => {
    const updatedLocationArchetype = await apiClient.patch<LocationArchetype>(`/sonar/locationArchetypes/${locationArchetype.id}`, locationArchetype);
    setLocationArchetypes(locationArchetypes.map(locationArchetype => locationArchetype.id === updatedLocationArchetype.id ? updatedLocationArchetype : locationArchetype));
  };

  const updateQuestArchetype = async (questArchetype: QuestArchetype) => {
    const updatedQuestArchetype = await apiClient.patch<QuestArchetype>(`/sonar/questArchetypes/${questArchetype.id}`, questArchetype);
    setQuestArchetypes(questArchetypes.map(questArchetype => questArchetype.id === updatedQuestArchetype.id ? updatedQuestArchetype : questArchetype));
  };

  const addChallengeToQuestArchetype = async (questArchetypeId: string, challenge: QuestArchetypeChallenge) => {
    const newChallenge = await apiClient.post<QuestArchetypeChallenge>(`/sonar/questArchetypes/${questArchetypeId}/challenges`, challenge);
    setQuestArchetypes(questArchetypes.map(questArchetype => questArchetype.id === questArchetypeId ? { ...questArchetype, challenges: [...questArchetype.challenges, newChallenge] } : questArchetype));
  };

  const deleteQuestArchetype = async (questArchetypeId: string) => {
    await apiClient.delete<QuestArchetype>(`/sonar/questArchetypes/${questArchetypeId}`);
    setQuestArchetypes(questArchetypes.filter(questArchetype => questArchetype.id !== questArchetypeId));
  };

  const deleteLocationArchetype = async (locationArchetypeId: string) => {
    await apiClient.delete<LocationArchetype>(`/sonar/locationArchetypes/${locationArchetypeId}`);
    setLocationArchetypes(locationArchetypes.filter(locationArchetype => locationArchetype.id !== locationArchetypeId));
  };

  return (
    <QuestArchetypesContext.Provider value={{ 
      questArchetypes, 
      locationArchetypes, 
      placeTypes, 
      createQuestArchetype,
      addChallengeToQuestArchetype,
      createLocationArchetype,
      updateLocationArchetype,
      updateQuestArchetype,
      deleteQuestArchetype,
      deleteLocationArchetype,
    }}>
      {children}
    </QuestArchetypesContext.Provider>
  );
};

export const useQuestArchetypes = () => {
  return useContext(QuestArchetypesContext);
};