import React, { useState, useEffect, useContext } from "react";
import { QuestArchetype, LocationArchetype, QuestArchetypeChallenge, QuestArchetypeNode, ZoneQuestArchetype } from "@poltergeist/types";
import { useAPI, useAuth } from "@poltergeist/contexts";

type QuestArchetypesContextType = {
  placeTypes: string[];
  locationArchetypes: LocationArchetype[];
  questArchetypes: QuestArchetype[];
  zoneQuestArchetypes: ZoneQuestArchetype[];
  createQuestArchetype: (name: string, locationArchetypeId: string, defaultGold?: number) => void;
  addChallengeToQuestArchetype: (questArchetypeId: string, rewardPoints: number, unlockedLocationArchetypeId?: string | null) => void;
  createLocationArchetype: (locationArchetype: LocationArchetype) => void;
  updateLocationArchetype: (locationArchetype: LocationArchetype) => void;
  updateQuestArchetype: (questArchetype: QuestArchetype) => void;
  deleteQuestArchetype: (questArchetypeId: string) => void;
  deleteLocationArchetype: (locationArchetypeId: string) => void;
  createZoneQuestArchetype: (zoneId: string, questArchetypeId: string, numberOfQuests: number) => void;
  deleteZoneQuestArchetype: (zoneQuestArchetypeId: string) => void;
};

export const QuestArchetypesContext = React.createContext<QuestArchetypesContextType>({
  questArchetypes: [],
  zoneQuestArchetypes: [],
  createQuestArchetype: () => {},
  addChallengeToQuestArchetype: () => {},
  locationArchetypes: [],
  createLocationArchetype: () => {},
  updateLocationArchetype: () => {},
  placeTypes: [],
  updateQuestArchetype: () => {},
  deleteQuestArchetype: () => {},
  deleteLocationArchetype: () => {},
  createZoneQuestArchetype: () => {},
  deleteZoneQuestArchetype: () => {},
});

export const QuestArchetypesProvider = ({ children }: { children: React.ReactNode }) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [questArchetypes, setQuestArchetypes] = useState<QuestArchetype[]>([]);
  const [locationArchetypes, setLocationArchetypes] = useState<LocationArchetype[]>([]);
  const [placeTypes, setPlaceTypes] = useState<string[]>([]);
  const [zoneQuestArchetypes, setZoneQuestArchetypes] = useState<ZoneQuestArchetype[]>([]);
  const populateChallengesForNode = async (node: QuestArchetypeNode) => {
    const challenges = await apiClient.get<QuestArchetypeChallenge[]>(`/sonar/questArchetypes/${node.id}/challenges`);
    node.challenges = challenges;
    node.challenges?.forEach(async (challenge) => {
      if (challenge.unlockedNode) {
        await populateChallengesForNode(challenge.unlockedNode);
      }
    });

    return node;
  };

  const fetchQuestArchetypes = async () => {
    const questArchetypes = await apiClient.get<QuestArchetype[]>("/sonar/questArchetypes");
    const populatedQuestArchetypes = await Promise.all(questArchetypes.map(async (questArchetype) => {
      questArchetype.root = await populateChallengesForNode(questArchetype.root);
      return questArchetype;
    }));
    setQuestArchetypes(populatedQuestArchetypes);
  };

  const fetchLocationArchetypes = async () => {
    const locationArchetypes = await apiClient.get<LocationArchetype[]>("/sonar/locationArchetypes");
    setLocationArchetypes(locationArchetypes);
  };

  const fetchPlaceTypes = async () => {
    const placeTypes = await apiClient.get<string[]>("/sonar/placeTypes");
    setPlaceTypes(placeTypes);
  };

  const fetchZoneQuestArchetypes = async () => {
    const zoneQuestArchetypes = await apiClient.get<ZoneQuestArchetype[]>("/sonar/zoneQuestArchetypes");
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

  const createLocationArchetype = async (locationArchetype: LocationArchetype) => {
    const newLocationArchetype = await apiClient.post<LocationArchetype>("/sonar/locationArchetypes", locationArchetype);
    setLocationArchetypes([...locationArchetypes, newLocationArchetype]);
  };

  const createQuestArchetype = async (name: string, locationArchetypeID: string, defaultGold?: number) => {
    const node = await apiClient.post<QuestArchetypeNode>("/sonar/questArchetypeNodes", {
      locationArchetypeID,
    });
    const questArchetype = await apiClient.post<QuestArchetype>("/sonar/questArchetypes", {
      name,
      rootId: node.id,
      defaultGold,
    });
    setQuestArchetypes([...questArchetypes, questArchetype]);
  };

  const updateLocationArchetype = async (locationArchetype: LocationArchetype) => {
    const updatedLocationArchetype = await apiClient.patch<LocationArchetype>(`/sonar/locationArchetypes/${locationArchetype.id}`, locationArchetype);
    setLocationArchetypes(locationArchetypes.map(locationArchetype => locationArchetype.id === updatedLocationArchetype.id ? updatedLocationArchetype : locationArchetype));
  };

  const updateQuestArchetype = async (questArchetype: QuestArchetype) => {
    const updatedQuestArchetype = await apiClient.patch<QuestArchetype>(`/sonar/questArchetypes/${questArchetype.id}`, questArchetype);
    setQuestArchetypes(questArchetypes.map(questArchetype => questArchetype.id === updatedQuestArchetype.id ? updatedQuestArchetype : questArchetype));
  };

  const addChallengeToQuestArchetype = async (questArchetypeId: string, reward: number, locationArchetypeID?: string | null) => {
    const payload: { reward: number; locationArchetypeID?: string } = {
      reward,
    };

    if (locationArchetypeID) {
      payload.locationArchetypeID = locationArchetypeID;
    }

    const newChallenge = await apiClient.post<QuestArchetypeChallenge>(
      `/sonar/questArchetypes/${questArchetypeId}/challenges`,
      payload
    );

    fetchQuestArchetypes();
  };

  const deleteQuestArchetype = async (questArchetypeId: string) => {
    await apiClient.delete<QuestArchetype>(`/sonar/questArchetypes/${questArchetypeId}`);
    setQuestArchetypes(questArchetypes.filter(questArchetype => questArchetype.id !== questArchetypeId));
  };

  const deleteLocationArchetype = async (locationArchetypeId: string) => {
    await apiClient.delete<LocationArchetype>(`/sonar/locationArchetypes/${locationArchetypeId}`);
    setLocationArchetypes(locationArchetypes.filter(locationArchetype => locationArchetype.id !== locationArchetypeId));
  };

  const createZoneQuestArchetype = async (zoneId: string, questArchetypeId: string, numberOfQuests: number) => {
    await apiClient.post<ZoneQuestArchetype>("/sonar/zoneQuestArchetypes", {
      zoneId,
      questArchetypeId,
      numberOfQuests,
    });
  };

  const deleteZoneQuestArchetype = async (zoneQuestArchetypeId: string) => {
    await apiClient.delete<ZoneQuestArchetype>(`/sonar/zoneQuestArchetypes/${zoneQuestArchetypeId}`);
    setZoneQuestArchetypes(zoneQuestArchetypes.filter(zoneQuestArchetype => zoneQuestArchetype.id !== zoneQuestArchetypeId));
  };

  return (
    <QuestArchetypesContext.Provider value={{ 
      questArchetypes, 
      locationArchetypes, 
      placeTypes, 
      zoneQuestArchetypes,
      createQuestArchetype,
      addChallengeToQuestArchetype,
      createLocationArchetype,
      updateLocationArchetype,
      updateQuestArchetype,
      deleteQuestArchetype,
      deleteLocationArchetype,
      createZoneQuestArchetype,
      deleteZoneQuestArchetype,
    }}>
      {children}
    </QuestArchetypesContext.Provider>
  );
};

export const useQuestArchetypes = () => {
  return useContext(QuestArchetypesContext);
};