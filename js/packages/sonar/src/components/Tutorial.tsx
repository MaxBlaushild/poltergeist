import React, { useState } from 'react';
import { DialogueComponent, Character } from './shared/DialogueComponent.tsx';
import { useMap } from '@poltergeist/contexts';
import { useTutorial } from '../contexts/TutorialContext.tsx';



export const Tutorial = () => {
  const { setAreOverlaysVisible } = useMap();
  const { isTutorialOpen, setIsTutorialOpen, setIsFilterButtonBeingDiscussed, setIsSearchButtonBeingDiscussed, setIsInventoryButtonBeingDiscussed, setIsQuestLogButtonBeingDiscussed } = useTutorial();
  const [currentStep, setCurrentStep] = useState(0);

  const tutorialSteps = [
    {
      dialogue: "Arrr, what scurvy dog let YOU wash up on these unclaimed streets? Keep yer head down if ye want to keep it attached to yer shoulders, ye landlubber. Old Gruffbeard'll show ye the ropes!",
      character: Character.Gruffbeard,
    },
    {
      dialogue: "Discovery be the real treasure 'round these parts. And you're as green as a cabin boy on his first voyage. Get out there and explore, ye scallywag!",
      character: Character.Gruffbeard,
    },
    {
      dialogue: "The locations on this 'ere map be shrouded in mystery, like a fog on the morning sea! Ye can unveil their secrets by giving 'em a tap and pressin' that shiny \"discover\" button, arrr!",
      character: Character.Gruffbeard,
    },
    {
      dialogue: "Once ye've discovered a spot, ye'll find more booty than a pirate's chest - quests to complete, treasures to plunder, and secrets that'd make Davy Jones himself blush!",
      character: Character.Gruffbeard,
      onAdvance: () => setIsFilterButtonBeingDiscussed(true),
    },
    {
      dialogue: "If ye want to sort through yer adventures like a proper navigator, just tap that \"filter\" button in the crow's nest - err, top right corner of yer map!",
      character: Character.Gruffbeard,
      onAdvance: () => {
        setIsSearchButtonBeingDiscussed(true);
        setIsFilterButtonBeingDiscussed(false);
      },
    },
    {
      dialogue: "Filters got ye scratching yer head? Just click the \"search\" and speak yer mind plain as day, ye savvy? Like chatting with yer first mate!",
      character: Character.Gruffbeard,
      onAdvance: () => {
        setIsInventoryButtonBeingDiscussed(true);
        setIsSearchButtonBeingDiscussed(false);
      },
    },
    {
      dialogue: "All yer plundered treasures be safely stored in yer inventory - just hit that button to check 'em! ",
      character: Character.Gruffbeard,
      onAdvance: () => {
        setIsQuestLogButtonBeingDiscussed(true);
        setIsInventoryButtonBeingDiscussed(false);
      },
    },
    {
      dialogue: "And keep an eye on yer quests in the log, like any good captain would!",
      character: Character.Gruffbeard,
      onAdvance: () => {
        setIsQuestLogButtonBeingDiscussed(false);
      },
    },
    {
      dialogue: "Shiver me timbers, that's all ye need to know! I've marked a special spot on yer map to get ye started, right proper! Now get out there and make old Gruffbeard proud, ye swab!",
      character: Character.Gruffbeard,
      onAdvance: () => {
      },
    },
  ];

  const onAdvance = () => {
    if (currentStep == tutorialSteps.length - 1) {
      setIsTutorialOpen(false);
      setAreOverlaysVisible(true);
    } else {
      setCurrentStep(currentStep + 1);
    }

    if (tutorialSteps[currentStep].onAdvance) {
      tutorialSteps[currentStep].onAdvance();
    }
  }

  if (!isTutorialOpen) {
    return null;
  }

  return (
    <DialogueComponent 
      dialogue={tutorialSteps[currentStep].dialogue} 
      character={tutorialSteps[currentStep].character} 
      onAdvance={onAdvance} 
    />
  );
};