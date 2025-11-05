import React, { useState, useEffect } from 'react';
import { DialogueMessage, CharacterAction } from '@poltergeist/types';

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
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editingText, setEditingText] = useState<string>('');

  useEffect(() => {
    if (action && action.dialogue) {
      // Sort by order to ensure correct sequence
      const sortedDialogue = [...action.dialogue].sort((a, b) => a.order - b.order);
      setDialogue(sortedDialogue);
    } else {
      setDialogue([]);
    }
  }, [action]);

  const handleDragStart = (e: React.DragEvent, index: number) => {
    e.dataTransfer.setData('dragIndex', index.toString());
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  };

  const handleDrop = (e: React.DragEvent, dropIndex: number) => {
    e.preventDefault();
    const dragIndex = parseInt(e.dataTransfer.getData('dragIndex'), 10);
    
    if (dragIndex === dropIndex) return;

    const newDialogue = [...dialogue];
    const [draggedItem] = newDialogue.splice(dragIndex, 1);
    newDialogue.splice(dropIndex, 0, draggedItem);

    // Update order field for all items
    const updatedDialogue = newDialogue.map((msg, idx) => ({
      ...msg,
      order: idx,
    }));

    setDialogue(updatedDialogue);
    setEditingIndex(null);
  };

  const addMessage = (speaker: 'character' | 'user') => {
    const newMessage: DialogueMessage = {
      speaker,
      text: '',
      order: dialogue.length,
    };
    setDialogue([...dialogue, newMessage]);
    setEditingIndex(dialogue.length);
    setEditingText('');
  };

  const deleteMessage = (index: number) => {
    const newDialogue = dialogue.filter((_, idx) => idx !== index);
    const updatedDialogue = newDialogue.map((msg, idx) => ({
      ...msg,
      order: idx,
    }));
    setDialogue(updatedDialogue);
    setEditingIndex(null);
  };

  const startEditing = (index: number) => {
    setEditingIndex(index);
    setEditingText(dialogue[index].text);
  };

  const saveEdit = () => {
    if (editingIndex !== null) {
      const updatedDialogue = [...dialogue];
      updatedDialogue[editingIndex] = {
        ...updatedDialogue[editingIndex],
        text: editingText,
      };
      setDialogue(updatedDialogue);
      setEditingIndex(null);
      setEditingText('');
    }
  };

  const toggleSpeaker = (index: number) => {
    const updatedDialogue = [...dialogue];
    updatedDialogue[index] = {
      ...updatedDialogue[index],
      speaker: updatedDialogue[index].speaker === 'character' ? 'user' : 'character',
    };
    setDialogue(updatedDialogue);
  };

  const handleSave = () => {
    onSave(dialogue);
  };

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      gap: '20px',
      minHeight: '300px',
      maxHeight: '600px',
      overflow: 'hidden'
    }}>
      <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 'bold' }}>
        Edit Dialogue
      </h3>

      {/* Dialogue List */}
      <div style={{
        flex: 1,
        overflowY: 'auto',
        border: '1px solid #ccc',
        borderRadius: '8px',
        padding: '15px',
        backgroundColor: '#f9f9f9'
      }}>
        {dialogue.length === 0 ? (
          <div style={{
            textAlign: 'center',
            color: '#999',
            padding: '40px',
            fontStyle: 'italic'
          }}>
            No dialogue messages yet. Add one to get started.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {dialogue.map((message, index) => (
              <div
                key={index}
                draggable
                onDragStart={(e) => handleDragStart(e, index)}
                onDragOver={handleDragOver}
                onDrop={(e) => handleDrop(e, index)}
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: '10px',
                  padding: '15px',
                  backgroundColor: message.speaker === 'character' ? '#e3f2fd' : '#f1f8e9',
                  borderRadius: '8px',
                  cursor: 'move',
                  border: '1px solid',
                  borderColor: message.speaker === 'character' ? '#90caf9' : '#c5e1a5',
                  transition: 'all 0.2s'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.transform = 'scale(1.02)';
                  e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.transform = 'scale(1)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                {/* Speaker Badge */}
                <div style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '5px',
                  alignItems: 'center',
                  minWidth: '60px'
                }}>
                  <div style={{
                    padding: '4px 8px',
                    borderRadius: '12px',
                    fontSize: '11px',
                    fontWeight: 'bold',
                    backgroundColor: message.speaker === 'character' ? '#2196f3' : '#8bc34a',
                    color: 'white',
                    whiteSpace: 'nowrap'
                  }}>
                    {message.speaker === 'character' ? 'CHAR' : 'USER'}
                  </div>
                  <button
                    onClick={() => toggleSpeaker(index)}
                    style={{
                      padding: '2px 6px',
                      fontSize: '10px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                      backgroundColor: 'white',
                      cursor: 'pointer'
                    }}
                  >
                    Switch
                  </button>
                </div>

                {/* Message Content */}
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '5px' }}>
                  {editingIndex === index ? (
                    <textarea
                      value={editingText}
                      onChange={(e) => setEditingText(e.target.value)}
                      style={{
                        width: '100%',
                        padding: '8px',
                        border: '2px solid #2196f3',
                        borderRadius: '4px',
                        fontSize: '14px',
                        minHeight: '60px'
                      }}
                      autoFocus
                    />
                  ) : (
                    <div
                      onClick={() => startEditing(index)}
                      style={{
                        cursor: 'text',
                        padding: '8px',
                        borderRadius: '4px',
                        fontSize: '14px',
                        lineHeight: '1.5',
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-word'
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.backgroundColor = 'rgba(255,255,255,0.5)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.backgroundColor = 'transparent';
                      }}
                    >
                      {message.text || <span style={{ color: '#999', fontStyle: 'italic' }}>Click to edit...</span>}
                    </div>
                  )}
                  {editingIndex === index && (
                    <div style={{ display: 'flex', gap: '5px', justifyContent: 'flex-end' }}>
                      <button
                        onClick={saveEdit}
                        style={{
                          padding: '4px 12px',
                          backgroundColor: '#4caf50',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer',
                          fontSize: '12px'
                        }}
                      >
                        Save
                      </button>
                      <button
                        onClick={() => {
                          setEditingIndex(null);
                          setEditingText('');
                        }}
                        style={{
                          padding: '4px 12px',
                          backgroundColor: '#999',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer',
                          fontSize: '12px'
                        }}
                      >
                        Cancel
                      </button>
                    </div>
                  )}
                </div>

                {/* Order Number */}
                <div style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: '5px',
                  minWidth: '40px'
                }}>
                  <div style={{
                    padding: '4px 8px',
                    borderRadius: '50%',
                    fontSize: '12px',
                    fontWeight: 'bold',
                    backgroundColor: 'white',
                    border: '1px solid #ccc',
                    width: '28px',
                    height: '28px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                  }}>
                    {index + 1}
                  </div>
                  <button
                    onClick={() => deleteMessage(index)}
                    style={{
                      padding: '2px 6px',
                      fontSize: '10px',
                      backgroundColor: '#f44336',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer'
                    }}
                  >
                    ✕
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Add Buttons */}
      <div style={{ display: 'flex', gap: '10px', justifyContent: 'center' }}>
        <button
          onClick={() => addMessage('character')}
          style={{
            padding: '8px 20px',
            backgroundColor: '#2196f3',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontSize: '14px',
            fontWeight: 'bold'
          }}
        >
          + Add Character Message
        </button>
        <button
          onClick={() => addMessage('user')}
          style={{
            padding: '8px 20px',
            backgroundColor: '#8bc34a',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontSize: '14px',
            fontWeight: 'bold'
          }}
        >
          + Add User Message
        </button>
      </div>

      {/* Action Buttons */}
      <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
        <button
          onClick={onCancel}
          style={{
            padding: '10px 20px',
            backgroundColor: '#757575',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontSize: '14px'
          }}
        >
          Cancel
        </button>
        <button
          onClick={handleSave}
          disabled={dialogue.length === 0 || dialogue.some(msg => !msg.text.trim())}
          style={{
            padding: '10px 20px',
            backgroundColor: dialogue.length === 0 || dialogue.some(msg => !msg.text.trim()) ? '#ccc' : '#4caf50',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: dialogue.length === 0 || dialogue.some(msg => !msg.text.trim()) ? 'not-allowed' : 'pointer',
            fontSize: '14px',
            fontWeight: 'bold'
          }}
        >
          Save Dialogue
        </button>
      </div>
    </div>
  );
};

