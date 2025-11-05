import React, { useState, useEffect } from 'react';
import { Character, CharacterAction, DialogueMessage } from '@poltergeist/types';
import './RPGDialogue.css';

interface RPGDialogueProps {
  character: Character;
  action: CharacterAction;
  onClose: () => void;
}

export const RPGDialogue: React.FC<RPGDialogueProps> = ({
  character,
  action,
  onClose,
}) => {
  const sortedDialogue = [...action.dialogue].sort((a, b) => a.order - b.order);
  const [currentDialogueIndex, setCurrentDialogueIndex] = useState(0);
  const currentDialogue = sortedDialogue[currentDialogueIndex];
  const hasNextDialogue = currentDialogueIndex < sortedDialogue.length - 1;

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      } else if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        if (hasNextDialogue) {
          setCurrentDialogueIndex(prev => prev + 1);
        } else {
          onClose();
        }
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => {
      window.removeEventListener('keydown', handleKeyPress);
    };
  }, [hasNextDialogue, onClose]);

  const advanceDialogue = () => {
    if (hasNextDialogue) {
      setCurrentDialogueIndex(currentDialogueIndex + 1);
    } else {
      onClose();
    }
  };

  const handleOverlayClick = (event: React.MouseEvent<HTMLDivElement>) => {
    // Only advance if clicking the overlay, not the textbox
    if (event.target === event.currentTarget) {
      advanceDialogue();
    }
  };

  const handleTextboxClick = () => {
    advanceDialogue();
  };

  const characterImageUrl = character.dialogueImageUrl || character.mapIconUrl;

  return (
    <div className="RPGDialogue__overlay" onClick={handleOverlayClick}>
      <div className="RPGDialogue__container">
        {/* Close Button */}
        <button
          className="RPGDialogue__close"
          onClick={onClose}
          aria-label="Close dialogue"
        >
          ✕
        </button>

        {/* Character Image */}
        {characterImageUrl && (
          <div className="RPGDialogue__character-image-container">
            <img
              src={characterImageUrl}
              alt={character.name}
              className="RPGDialogue__character-image"
            />
          </div>
        )}

        {/* Textbox */}
        <div className="RPGDialogue__textbox" onClick={handleTextboxClick}>
          <div className="RPGDialogue__textbox-content">
            {/* Speaker Name */}
            <div className="RPGDialogue__speaker">
              {currentDialogue.speaker === 'character' ? character.name : 'You'}
            </div>
            
            {/* Dialogue Text */}
            <div className="RPGDialogue__text">
              {currentDialogue.text}
            </div>

            {/* Blinking Next Indicator */}
            {hasNextDialogue && (
              <div className="RPGDialogue__next-indicator">
                ▼
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

