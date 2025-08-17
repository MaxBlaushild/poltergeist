import React, { useState } from 'react';
import { useTutorial } from '../contexts/TutorialContext.tsx';
import { useMap } from '@poltergeist/contexts';
import { DialogueComponent, Character } from './shared/DialogueComponent.tsx';

const tutorialSteps = [
  {
    dialogue: "Arrr, ye did it! Ye've completed yer trainin' and proved yerself worthy! Now get out there and explore the seven seas, find some treasure, and complete them quests, ye scurvy dog!",
    character: Character.Gruffbeard,
  },
];

export const TutorialComplete = () => {
  const { setAreOverlaysVisible } = useMap();
  const { isTutorialOpen, setIsTutorialOpen, isTutorialCompleted, isTrainingCompleted, setIsTutorialCompleted } = useTutorial();
  const [currentStep, setCurrentStep] = useState(0);

  if (isTutorialCompleted || !isTrainingCompleted) {
    return null;
  }

  const onAdvance = () => {
    if (currentStep == tutorialSteps.length - 1) {
      setIsTutorialCompleted(true); 
      setAreOverlaysVisible(true);
    } else {
      setCurrentStep(currentStep + 1);
    }
  }
  return (
    <DialogueComponent 
      dialogue={tutorialSteps[currentStep].dialogue} 
      character={tutorialSteps[currentStep].character} 
      onAdvance={onAdvance} 
    />
  );
};