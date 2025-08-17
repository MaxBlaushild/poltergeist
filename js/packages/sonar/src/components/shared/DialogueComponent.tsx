import React from 'react';
import { useTutorial } from '../../contexts/TutorialContext.tsx';
export enum Character {
  Gruffbeard = 'gruffbeard',
  Fatbeard = 'fatbeard',
  Meanbeard = 'meanbeard',
}

const characterToUrlMap = {
  [Character.Gruffbeard]: 'https://crew-profile-icons.s3.us-east-1.amazonaws.com/gruff-pirate.png',
  [Character.Fatbeard]: 'https://crew-profile-icons.s3.us-east-1.amazonaws.com/fat-pirate.png',
  [Character.Meanbeard]: 'https://crew-profile-icons.s3.us-east-1.amazonaws.com/mean-pirate.png',
}

const getCharacterUrl = (character: Character) => {
  const url = characterToUrlMap[character];
  if (!url) {
    throw new Error(`Character ${character} not found`);
  }
  return url;
}

interface DialogueComponentProps {
  dialogue: string;
  character: Character;
  onAdvance?: () => void;
}

export const DialogueComponent = ({ dialogue, character, onAdvance }: DialogueComponentProps) => {
  const characterUrl = getCharacterUrl(character);
  const { isInventoryButtonBeingDiscussed, isQuestLogButtonBeingDiscussed } = useTutorial();

  return (
    <>
      <div 
        className="fixed inset-0 z-[9997] cursor-pointer"
        onClick={() => onAdvance?.()}
      />
      <div 
        className="flex justify-end items-end fixed left-0 right-0 z-[9998] flex flex-col gap-2 p-0 bg-transparent" 
        style={{
          bottom: isInventoryButtonBeingDiscussed || isQuestLogButtonBeingDiscussed ? '80px' : '0',
          transition: 'bottom 0.3s ease-in-out'
        }}
        onClick={() => onAdvance?.()}
      >
        <div className={`absolute right-2 bottom-2 z-10 bg-white border-2 z-[9999] border-gray-800 rounded-lg p-4 left-2 shadow-lg`}>
          <h3 className="font-bold text-gray-900 mb-2">{character.charAt(0).toUpperCase() + character.slice(1)}</h3>
          <div className="flex items-center gap-4">
            <p className="text-gray-900 flex-1">{dialogue}</p>
            <svg className="w-6 h-6 text-gray-400 shrink-0 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </div>
        <img 
          src={characterUrl} 
          alt={character} 
          className="w-96 h-96 object-cover z-20" 
        />
      </div>
    </>
  );
}
