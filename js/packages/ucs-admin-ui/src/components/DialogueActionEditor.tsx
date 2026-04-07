import React, { useEffect, useState } from 'react';
import { DialogueMessage, CharacterAction } from '@poltergeist/types';
import { DialogueMessageListEditor } from './DialogueMessageListEditor.tsx';

interface DialogueActionEditorProps {
  action: CharacterAction | null;
  onSave: (dialogue: DialogueMessage[]) => void;
  onCancel: () => void;
}

export const DialogueActionEditor: React.FC<DialogueActionEditorProps> = ({
  action,
  onSave,
  onCancel,
}) => {
  const [dialogue, setDialogue] = useState<DialogueMessage[]>([]);

  useEffect(() => {
    if (action?.dialogue) {
      setDialogue([...action.dialogue].sort((a, b) => a.order - b.order));
      return;
    }
    setDialogue([]);
  }, [action]);

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: '20px',
        minHeight: '300px',
        maxHeight: '600px',
      }}
    >
      <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 'bold' }}>
        Edit Dialogue
      </h3>

      <div style={{ flex: 1, overflowY: 'auto' }}>
        <DialogueMessageListEditor
          value={dialogue}
          onChange={setDialogue}
          allowSpeakerToggle
          helperText="Add per-line effects like Angry to control special client-side presentation."
        />
      </div>

      <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
        <button
          onClick={onCancel}
          style={{
            padding: '10px 20px',
            border: '1px solid #ccc',
            borderRadius: '6px',
            backgroundColor: 'white',
            cursor: 'pointer',
            fontSize: '14px',
          }}
        >
          Cancel
        </button>
        <button
          onClick={() => onSave(dialogue)}
          style={{
            padding: '10px 20px',
            border: 'none',
            borderRadius: '6px',
            backgroundColor: '#2196f3',
            color: 'white',
            cursor: 'pointer',
            fontSize: '14px',
            fontWeight: 'bold',
          }}
        >
          Save Dialogue
        </button>
      </div>
    </div>
  );
};
