import React, { createContext, useState, useContext, useEffect } from 'react';
import { useUserProfiles } from './UserProfileContext.tsx';
import { useAPI } from '@poltergeist/contexts';

interface TutorialContextType {
  isTutorialOpen: boolean;
  isTrainingCompleted: boolean;
  isTutorialCompleted: boolean;
  isFilterButtonBeingDiscussed: boolean;
  isSearchButtonBeingDiscussed: boolean;
  isInventoryButtonBeingDiscussed: boolean;
  isQuestLogButtonBeingDiscussed: boolean;
  setIsTutorialOpen: (isOpen: boolean) => void;
  setIsTrainingCompleted: (isCompleted: boolean) => void;
  setIsTutorialCompleted: (isCompleted: boolean) => void;
  setIsFilterButtonBeingDiscussed: (isDiscussed: boolean) => void;
  setIsSearchButtonBeingDiscussed: (isDiscussed: boolean) => void;
  setIsInventoryButtonBeingDiscussed: (isDiscussed: boolean) => void;
  setIsQuestLogButtonBeingDiscussed: (isDiscussed: boolean) => void;
}

const TutorialContext = createContext<TutorialContextType | undefined>(undefined);

export const useTutorial = () => {
  const context = useContext(TutorialContext);
  if (!context) {
    throw new Error('useTutorial must be used within a TutorialProvider');
  }
  return context;
};

export const TutorialProvider = ({ children }: { children: React.ReactNode }) => {
  const { apiClient } = useAPI();
  const { currentUser } = useUserProfiles();
  const [isTutorialOpen, setIsTutorialOpen] = useState(false);
  const [isTutorialCompleted, _setIsTutorialCompleted] = useState(false);
  const [isTrainingCompleted, setIsTrainingCompleted] = useState(false);
  const [isFilterButtonBeingDiscussed, setIsFilterButtonBeingDiscussed] = useState(false);
  const [isSearchButtonBeingDiscussed, setIsSearchButtonBeingDiscussed] = useState(false);
  const [isInventoryButtonBeingDiscussed, setIsInventoryButtonBeingDiscussed] = useState(false);
  const [isQuestLogButtonBeingDiscussed, setIsQuestLogButtonBeingDiscussed] = useState(false);

  useEffect(() => {
    if (currentUser) {
      setIsTutorialOpen(!currentUser.hasSeenTutorial);
      _setIsTutorialCompleted(currentUser.hasSeenTutorial);
    }
  }, [currentUser]);

  const setIsTutorialCompleted = async (isCompleted: boolean) => {
    if (currentUser) {
      await apiClient.patch(`/sonar/users/hasSeenTutorial`, { hasSeenTutorial: isCompleted });
    }
    _setIsTutorialCompleted(isCompleted);
  }

  return (
    <TutorialContext.Provider value={{ 
      isTutorialOpen, 
      setIsTutorialOpen, 
      isTutorialCompleted, 
      setIsTutorialCompleted, 
      isTrainingCompleted, 
      setIsTrainingCompleted, 
      isFilterButtonBeingDiscussed, 
      setIsFilterButtonBeingDiscussed, 
      isSearchButtonBeingDiscussed, 
      setIsSearchButtonBeingDiscussed,
      isInventoryButtonBeingDiscussed,
      setIsInventoryButtonBeingDiscussed,
      isQuestLogButtonBeingDiscussed,
      setIsQuestLogButtonBeingDiscussed,
    }}>
      {children}
    </TutorialContext.Provider>
  );
};
