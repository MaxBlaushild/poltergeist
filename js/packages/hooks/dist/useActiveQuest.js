import { useState, useEffect } from 'react';
const getStorageKey = (userId) => `activeQuest:${userId}`;
export const useActiveQuest = (userId) => {
    const [activeQuestId, setActiveQuestIdState] = useState(() => {
        if (!userId)
            return null;
        return localStorage.getItem(getStorageKey(userId));
    });
    const setActiveQuestId = (questId) => {
        if (!userId)
            return;
        if (questId) {
            localStorage.setItem(getStorageKey(userId), questId);
        }
        else {
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
