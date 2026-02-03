import React, { useState, useEffect } from 'react';
import { Character, CharacterAction, DialogueMessage } from '@poltergeist/types';
import { useAPI } from '@poltergeist/contexts';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import './CharacterPanel.css';

interface CharacterPanelProps {
  character: Character;
  onClose: () => void;
  onStartDialogue?: (character: Character, action: CharacterAction) => void;
  onStartShop?: (character: Character, action: CharacterAction) => void;
}

export const CharacterPanel: React.FC<CharacterPanelProps> = ({
  character,
  onClose,
  onStartDialogue,
  onStartShop,
}) => {
  const { apiClient } = useAPI();
  const { refreshQuestLog } = useQuestLogContext();
  const [characterActions, setCharacterActions] = useState<CharacterAction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isAcceptingQuest, setIsAcceptingQuest] = useState(false);

  useEffect(() => {
    const fetchCharacterActions = async () => {
      setIsLoading(true);
      try {
        const response = await apiClient.get<CharacterAction[]>(
          `/sonar/characters/${character.id}/actions`
        );
        const hasTalkAction = response.some((action) => action.actionType === 'talk');
        if (!hasTalkAction) {
          const fallbackTalkAction: CharacterAction = {
            id: `local-talk-${character.id}`,
            createdAt: new Date(),
            updatedAt: new Date(),
            characterId: character.id,
            actionType: 'talk',
            dialogue: [{ speaker: 'character', text: '...', order: 0 }],
          };
          setCharacterActions([fallbackTalkAction, ...response]);
        } else {
          setCharacterActions(response);
        }
      } catch (error) {
        console.error('Error fetching character actions:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchCharacterActions();
  }, [character.id, apiClient]);

  const handleActionClick = async (action: CharacterAction) => {
    if (action.actionType === 'shop' && onStartShop) {
      onStartShop(character, action);
      onClose();
    } else if (action.actionType === 'talk' && onStartDialogue) {
      onStartDialogue(character, action);
      onClose();
    } else if (action.actionType === 'giveQuest') {
      const questId = action.metadata?.pointOfInterestGroupId;
      if (!questId) {
        console.error('Quest ID not found in action metadata');
        return;
      }

      setIsAcceptingQuest(true);
      try {
        await apiClient.post('/sonar/quests/accept', {
          characterId: character.id,
          pointOfInterestGroupId: questId,
        });
        await refreshQuestLog();
        onClose();
      } catch (error) {
        console.error('Error accepting quest:', error);
      } finally {
        setIsAcceptingQuest(false);
      }
    }
  };

  return (
    <div className="CharacterPanel">
      {/* RPG-style Character Portrait Section */}
      <div className="CharacterPanel__portrait-section">
        <div className="CharacterPanel__portrait-frame">
          {character.dialogueImageUrl ? (
            <img
              src={character.dialogueImageUrl}
              alt={character.name}
              className="CharacterPanel__portrait"
            />
          ) : character.mapIconUrl ? (
            <img
              src={character.mapIconUrl}
              alt={character.name}
              className="CharacterPanel__portrait"
            />
          ) : (
            <div className="CharacterPanel__portrait CharacterPanel__portrait-placeholder">
              <span className="CharacterPanel__portrait-emoji">👤</span>
            </div>
          )}
        </div>
        
        {/* Character Info Box */}
        <div className="CharacterPanel__info-box">
          <h2 className="CharacterPanel__name">{character.name}</h2>
          {character.description && (
            <p className="CharacterPanel__description">{character.description}</p>
          )}
        </div>
      </div>

      {/* Actions List (RPG-style ability buttons) */}
      <div className="CharacterPanel__actions-section">
        <h3 className="CharacterPanel__actions-title">Actions</h3>
        {isLoading ? (
          <div className="CharacterPanel__loading">Loading actions...</div>
        ) : characterActions.length === 0 ? (
          <div className="CharacterPanel__empty">
            This character has no actions yet.
          </div>
        ) : (
          <div className="CharacterPanel__actions-grid">
            {characterActions.map((action) => (
              <button
                key={action.id}
                onClick={() => handleActionClick(action)}
                className="CharacterPanel__action-button"
                disabled={action.actionType === 'giveQuest' && isAcceptingQuest}
              >
                <div className="CharacterPanel__action-icon">
                  {action.actionType === 'shop' ? '💰' : action.actionType === 'giveQuest' ? '📜' : '💬'}
                </div>
                <div className="CharacterPanel__action-name">
                  {action.actionType === 'talk' ? 'Talk' : action.actionType === 'shop' ? 'Shop' : action.actionType === 'giveQuest' ? 'Give Quest' : action.actionType}
                </div>
                {action.actionType === 'talk' && action.dialogue && action.dialogue.length > 0 && (
                  <div className="CharacterPanel__action-count">
                    {action.dialogue.length} line{action.dialogue.length !== 1 ? 's' : ''}
                  </div>
                )}
                {action.actionType === 'shop' && action.metadata?.inventory && (
                  <div className="CharacterPanel__action-count">
                    {action.metadata.inventory.length} item{action.metadata.inventory.length !== 1 ? 's' : ''}
                  </div>
                )}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
