import { useState, useEffect } from 'react';

const getStorageKey = (userId: string) => `activeQuest:${userId}`;

export const useActiveQuest = (userId: string) => {
  const [activeQuestId, setActiveQuestIdState] = useState<string | null>(() => {
    if (!userId) return null;
    return localStorage.getItem(getStorageKey(userId));
  });

  const setActiveQuestId = (questId: string | null) => {
    if (!userId) return;
    
    if (questId) {
      localStorage.setItem(getStorageKey(userId), questId);
    } else {
      localStorage.removeItem(getStorageKey(userId));
    }
    setActiveQuestIdState(questId);
  };

  // Clear quest if userId changes
  useEffect(() => {
    setActiveQuestIdState(localStorage.getItem(getStorageKey(userId)));
  }, [userId]);

  return {
    activeQuestId,
    setActiveQuestId
  };
};

