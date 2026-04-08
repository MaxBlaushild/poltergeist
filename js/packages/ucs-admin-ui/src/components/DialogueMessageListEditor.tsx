import React from 'react';
import { DialogueEffect, DialogueMessage } from '@poltergeist/types';

const dialogueEffectOptions: Array<{ value: '' | DialogueEffect; label: string }> = [
  { value: '', label: 'None' },
  { value: 'angry', label: 'Angry' },
  { value: 'surprised', label: 'Surprised' },
  { value: 'whisper', label: 'Whisper' },
  { value: 'shout', label: 'Shout' },
  { value: 'mysterious', label: 'Mysterious' },
  { value: 'determined', label: 'Determined' },
];

const normalizeDialogueEffect = (effect?: DialogueEffect | ''): DialogueEffect | undefined => {
  switch (effect) {
    case 'angry':
    case 'surprised':
    case 'whisper':
    case 'shout':
    case 'mysterious':
    case 'determined':
      return effect;
    default:
      return undefined;
  }
};

type DialogueMessageListEditorProps = {
  label?: string;
  helperText?: string;
  value: DialogueMessage[];
  onChange: (value: DialogueMessage[]) => void;
  defaultSpeaker?: 'character' | 'user';
  allowSpeakerToggle?: boolean;
  characterOptions?: Array<{ value: string; label: string }>;
  requireCharacterSelection?: boolean;
};

const normalizeDialogueMessages = (messages: DialogueMessage[]) =>
  messages.map((message, index) => ({
    speaker: message.speaker === 'user' ? 'user' : 'character',
    text: message.text ?? '',
    order: index,
    effect: normalizeDialogueEffect(message.effect),
    characterId:
      message.speaker === 'user'
        ? undefined
        : message.characterId?.trim() || undefined,
  }));

export const DialogueMessageListEditor: React.FC<DialogueMessageListEditorProps> = ({
  label = 'Dialogue',
  helperText,
  value,
  onChange,
  defaultSpeaker = 'character',
  allowSpeakerToggle = false,
  characterOptions = [],
  requireCharacterSelection = false,
}) => {
  const messages = normalizeDialogueMessages(value ?? []);
  const resolvedHelperText = helperText
    ? `${helperText} Use {{username}} to insert the viewer's username.`
    : "Use {{username}} to insert the viewer's username.";

  const commit = (next: DialogueMessage[]) => {
    onChange(normalizeDialogueMessages(next));
  };

  const addMessage = () => {
    commit([
      ...messages,
      {
        speaker: defaultSpeaker,
        text: '',
        order: messages.length,
        characterId:
          defaultSpeaker === 'character' && characterOptions.length > 0
            ? characterOptions[0].value
            : undefined,
      },
    ]);
  };

  const updateMessage = (index: number, updates: Partial<DialogueMessage>) => {
    commit(
      messages.map((message, messageIndex) =>
        messageIndex === index ? { ...message, ...updates } : message,
      ),
    );
  };

  const removeMessage = (index: number) => {
    commit(messages.filter((_, messageIndex) => messageIndex !== index));
  };

  const moveMessage = (index: number, direction: -1 | 1) => {
    const nextIndex = index + direction;
    if (nextIndex < 0 || nextIndex >= messages.length) return;
    const next = [...messages];
    const [entry] = next.splice(index, 1);
    next.splice(nextIndex, 0, entry);
    commit(next);
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: '12px', alignItems: 'center' }}>
        <div>
          <label style={{ display: 'block', marginBottom: helperText ? '4px' : '6px', fontWeight: 500 }}>
            {label}
          </label>
          <p style={{ margin: 0, fontSize: '12px', color: '#6b7280' }}>{resolvedHelperText}</p>
        </div>
        <button
          type="button"
          onClick={addMessage}
          className="bg-white text-gray-700 px-3 py-1 rounded-md"
          style={{ border: '1px solid #d1d5db' }}
        >
          Add Line
        </button>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', marginTop: '10px' }}>
        {messages.length === 0 ? (
          <div
            style={{
              border: '1px dashed #d1d5db',
              borderRadius: '8px',
              padding: '12px',
              color: '#6b7280',
              fontSize: '13px',
              backgroundColor: '#f9fafb',
            }}
          >
            No dialogue lines yet.
          </div>
        ) : null}
        {messages.map((message, index) => (
          <div
            key={`${message.order}-${index}`}
            style={{
              border: '1px solid #d1d5db',
              borderRadius: '8px',
              padding: '12px',
              backgroundColor: '#fff',
              display: 'flex',
              flexDirection: 'column',
              gap: '10px',
            }}
          >
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center', flexWrap: 'wrap' }}>
              <div
                style={{
                  fontSize: '12px',
                  fontWeight: 700,
                  color: '#4b5563',
                  minWidth: '52px',
                }}
              >
                Line {index + 1}
              </div>
              {allowSpeakerToggle ? (
                <select
                  value={message.speaker}
                  onChange={(event) => {
                    const nextSpeaker =
                      event.target.value === 'user' ? 'user' : 'character';
                    updateMessage(index, {
                      speaker: nextSpeaker,
                      characterId:
                        nextSpeaker === 'character'
                          ? message.characterId ??
                            (characterOptions[0]?.value || undefined)
                          : undefined,
                    });
                  }}
                  style={{
                    padding: '6px 8px',
                    border: '1px solid #d1d5db',
                    borderRadius: '6px',
                    fontSize: '12px',
                  }}
                >
                  <option value="character">Character</option>
                  <option value="user">User</option>
                </select>
              ) : (
                <div
                  style={{
                    fontSize: '12px',
                    color: '#6b7280',
                    border: '1px solid #e5e7eb',
                    borderRadius: '999px',
                    padding: '4px 8px',
                  }}
                >
                  {message.speaker === 'user' ? 'User' : 'Character'}
                </div>
              )}
              {message.speaker === 'character' && characterOptions.length > 0 ? (
                <select
                  value={message.characterId ?? ''}
                  onChange={(event) =>
                    updateMessage(index, {
                      characterId: event.target.value || undefined,
                    })
                  }
                  style={{
                    padding: '6px 8px',
                    border: '1px solid #d1d5db',
                    borderRadius: '6px',
                    fontSize: '12px',
                    minWidth: '160px',
                  }}
                >
                  {!requireCharacterSelection ? (
                    <option value="">Any character</option>
                  ) : null}
                  {characterOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              ) : null}
              <select
                value={message.effect ?? ''}
                onChange={(event) =>
                  updateMessage(index, {
                    effect: normalizeDialogueEffect(
                      event.target.value as DialogueEffect | '',
                    ),
                  })
                }
                style={{
                  padding: '6px 8px',
                  border: '1px solid #d1d5db',
                  borderRadius: '6px',
                  fontSize: '12px',
                }}
              >
                {dialogueEffectOptions.map((option) => (
                  <option key={option.value || 'none'} value={option.value}>
                    Effect: {option.label}
                  </option>
                ))}
              </select>
              <div style={{ marginLeft: 'auto', display: 'flex', gap: '6px' }}>
                <button
                  type="button"
                  onClick={() => moveMessage(index, -1)}
                  disabled={index === 0}
                  className="bg-white text-gray-700 px-2 py-1 rounded-md disabled:opacity-50"
                  style={{ border: '1px solid #d1d5db' }}
                >
                  Up
                </button>
                <button
                  type="button"
                  onClick={() => moveMessage(index, 1)}
                  disabled={index === messages.length - 1}
                  className="bg-white text-gray-700 px-2 py-1 rounded-md disabled:opacity-50"
                  style={{ border: '1px solid #d1d5db' }}
                >
                  Down
                </button>
                <button
                  type="button"
                  onClick={() => removeMessage(index)}
                  className="bg-red-50 text-red-700 px-2 py-1 rounded-md"
                  style={{ border: '1px solid #fecaca' }}
                >
                  Remove
                </button>
              </div>
            </div>
            <textarea
              value={message.text}
              onChange={(event) => updateMessage(index, { text: event.target.value })}
              placeholder="Dialogue line"
              style={{
                width: '100%',
                padding: '8px',
                border: '1px solid #d1d5db',
                borderRadius: '6px',
                minHeight: '72px',
              }}
            />
          </div>
        ))}
      </div>
    </div>
  );
};
