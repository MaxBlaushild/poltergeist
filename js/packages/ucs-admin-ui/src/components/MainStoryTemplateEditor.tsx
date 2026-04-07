import React, { useMemo, useState } from 'react';
import {
  LocationArchetype,
  MainStoryBeatDraft,
  MainStoryTemplate,
  QuestArchetype,
} from '@poltergeist/types';
import './questArchetypeTheme.css';

const parseList = (value: string) =>
  value
    .split(',')
    .map((entry) => entry.trim())
    .filter(Boolean);

const parseLines = (value: string) =>
  value
    .split('\n')
    .map((entry) => entry.trim())
    .filter(Boolean);

const emptyBeat = (orderIndex: number): MainStoryBeatDraft => ({
  orderIndex,
  act: 1,
  storyRole: '',
  chapterTitle: '',
  chapterSummary: '',
  purpose: '',
  whatChanges: '',
  introducedCharacterKeys: [],
  requiredCharacterKeys: [],
  introducedRevealKeys: [],
  requiredRevealKeys: [],
  requiredZoneTags: [],
  requiredLocationArchetypeIds: [],
  preferredContentMix: [],
  questGiverCharacterKey: '',
  questGiverCharacterId: null,
  questGiverCharacterName: '',
  name: '',
  hook: '',
  description: '',
  acceptanceDialogue: [],
  requiredStoryFlags: [],
  setStoryFlags: [],
  clearStoryFlags: [],
  questGiverRelationshipEffects: {
    trust: 0,
    respect: 0,
    fear: 0,
    debt: 0,
  },
  worldChanges: [],
  unlockedScenarios: [],
  unlockedChallenges: [],
  unlockedMonsterEncounters: [],
  questGiverAfterDescription: '',
  questGiverAfterDialogue: [],
  characterTags: [],
  internalTags: [],
  difficultyMode: 'scale',
  difficulty: 1,
  monsterEncounterTargetLevel: 1,
  whyThisScales: '',
  steps: [],
  challengeTemplateSeeds: [],
  scenarioTemplateSeeds: [],
  monsterTemplateSeeds: [],
  warnings: [],
  questArchetypeId: null,
  questArchetypeName: '',
});

const emptyTemplate = (): MainStoryTemplate => ({
  id: '',
  createdAt: '',
  updatedAt: '',
  name: '',
  premise: '',
  districtFit: '',
  tone: '',
  themeTags: [],
  internalTags: [],
  factionKeys: [],
  characterKeys: [],
  revealKeys: [],
  climaxSummary: '',
  resolutionSummary: '',
  whyItWorks: '',
  beats: [emptyBeat(1)],
});

type Props = {
  initialTemplate?: MainStoryTemplate | null;
  questArchetypes: QuestArchetype[];
  locationArchetypes: LocationArchetype[];
  saving: boolean;
  onCancel: () => void;
  onSave: (template: MainStoryTemplate) => Promise<void> | void;
};

export const MainStoryTemplateEditor = ({
  initialTemplate,
  questArchetypes,
  locationArchetypes,
  saving,
  onCancel,
  onSave,
}: Props) => {
  const [template, setTemplate] = useState<MainStoryTemplate>(
    initialTemplate
      ? {
          ...initialTemplate,
          beats:
            initialTemplate.beats?.length > 0
              ? initialTemplate.beats
              : [emptyBeat(1)],
        }
      : emptyTemplate()
  );
  const [expandedBeatIndexes, setExpandedBeatIndexes] = useState<number[]>([0]);

  const sortedQuestArchetypes = useMemo(
    () => [...questArchetypes].sort((a, b) => a.name.localeCompare(b.name)),
    [questArchetypes]
  );

  const toggleBeatExpanded = (index: number) => {
    setExpandedBeatIndexes((current) =>
      current.includes(index)
        ? current.filter((value) => value !== index)
        : [...current, index]
    );
  };

  const setTemplateField = <K extends keyof MainStoryTemplate>(
    key: K,
    value: MainStoryTemplate[K]
  ) => {
    setTemplate((current) => ({ ...current, [key]: value }));
  };

  const updateBeat = (
    beatIndex: number,
    updater: (beat: MainStoryBeatDraft) => MainStoryBeatDraft
  ) => {
    setTemplate((current) => ({
      ...current,
      beats: current.beats.map((beat, index) =>
        index === beatIndex ? updater(beat) : beat
      ),
    }));
  };

  const resequenceBeats = (beats: MainStoryBeatDraft[]) =>
    beats.map((beat, index) => ({
      ...beat,
      orderIndex: index + 1,
    }));

  const addBeat = () => {
    setTemplate((current) => ({
      ...current,
      beats: resequenceBeats([...current.beats, emptyBeat(current.beats.length + 1)]),
    }));
    setExpandedBeatIndexes((current) => [...current, template.beats.length]);
  };

  const removeBeat = (beatIndex: number) => {
    setTemplate((current) => ({
      ...current,
      beats: resequenceBeats(
        current.beats.filter((_, index) => index !== beatIndex)
      ),
    }));
    setExpandedBeatIndexes((current) =>
      current.filter((index) => index !== beatIndex).map((index) =>
        index > beatIndex ? index - 1 : index
      )
    );
  };

  const moveBeat = (beatIndex: number, direction: -1 | 1) => {
    setTemplate((current) => {
      const nextIndex = beatIndex + direction;
      if (nextIndex < 0 || nextIndex >= current.beats.length) {
        return current;
      }
      const beats = [...current.beats];
      const [moved] = beats.splice(beatIndex, 1);
      beats.splice(nextIndex, 0, moved);
      return { ...current, beats: resequenceBeats(beats) };
    });
    setExpandedBeatIndexes((current) =>
      current.map((index) => {
        if (index === beatIndex) return beatIndex + direction;
        if (index === beatIndex + direction) return beatIndex;
        return index;
      })
    );
  };

  const handleSave = async () => {
    await onSave({
      ...template,
      beats: resequenceBeats(template.beats),
    });
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(4, 8, 11, 0.72)',
        backdropFilter: 'blur(6px)',
        zIndex: 1000,
        overflowY: 'auto',
        padding: 24,
      }}
    >
      <div
        className="qa-card"
        style={{
          maxWidth: 1240,
          margin: '0 auto',
          background:
            'linear-gradient(180deg, rgba(20, 30, 36, 0.98), rgba(10, 17, 22, 0.98))',
        }}
      >
        <div className="qa-card-header">
          <div>
            <div className="qa-kicker">Main Story Authoring</div>
            <h2 className="qa-title" style={{ fontSize: 'clamp(28px, 3vw, 36px)' }}>
              {initialTemplate ? 'Edit Main Story Template' : 'New Main Story Template'}
            </h2>
            <div className="qa-meta">
              Build the campaign shell, then author each beat’s quest flow,
              unlock content, and story-state transitions.
            </div>
          </div>
          <div className="qa-actions">
            <button type="button" className="qa-btn qa-btn-outline" onClick={onCancel}>
              Cancel
            </button>
            <button
              type="button"
              className="qa-btn qa-btn-primary"
              onClick={() => void handleSave()}
              disabled={saving}
            >
              {saving ? 'Saving...' : initialTemplate ? 'Save Template' : 'Create Template'}
            </button>
          </div>
        </div>

        <div className="qa-stat-grid" style={{ marginTop: 20 }}>
          <div className="qa-field">
            <div className="qa-label">Name</div>
            <input
              className="qa-input"
              value={template.name}
              onChange={(event) => setTemplateField('name', event.target.value)}
            />
          </div>
          <div className="qa-field">
            <div className="qa-label">Tone</div>
            <input
              className="qa-input"
              value={template.tone}
              onChange={(event) => setTemplateField('tone', event.target.value)}
            />
          </div>
          <div className="qa-field qa-field--full">
            <div className="qa-label">Premise</div>
            <textarea
              className="qa-textarea"
              rows={3}
              value={template.premise}
              onChange={(event) => setTemplateField('premise', event.target.value)}
            />
          </div>
          <div className="qa-field qa-field--full">
            <div className="qa-label">District Fit</div>
            <textarea
              className="qa-textarea"
              rows={2}
              value={template.districtFit}
              onChange={(event) =>
                setTemplateField('districtFit', event.target.value)
              }
            />
          </div>
          <div className="qa-field">
            <div className="qa-label">Theme Tags</div>
            <input
              className="qa-input"
              value={template.themeTags.join(', ')}
              onChange={(event) =>
                setTemplateField('themeTags', parseList(event.target.value))
              }
            />
          </div>
          <div className="qa-field">
            <div className="qa-label">Internal Tags</div>
            <input
              className="qa-input"
              value={template.internalTags.join(', ')}
              onChange={(event) =>
                setTemplateField('internalTags', parseList(event.target.value))
              }
            />
          </div>
          <div className="qa-field">
            <div className="qa-label">Faction Keys</div>
            <input
              className="qa-input"
              value={template.factionKeys.join(', ')}
              onChange={(event) =>
                setTemplateField('factionKeys', parseList(event.target.value))
              }
            />
          </div>
          <div className="qa-field">
            <div className="qa-label">Character Keys</div>
            <input
              className="qa-input"
              value={template.characterKeys.join(', ')}
              onChange={(event) =>
                setTemplateField('characterKeys', parseList(event.target.value))
              }
            />
          </div>
          <div className="qa-field">
            <div className="qa-label">Reveal Keys</div>
            <input
              className="qa-input"
              value={template.revealKeys.join(', ')}
              onChange={(event) =>
                setTemplateField('revealKeys', parseList(event.target.value))
              }
            />
          </div>
          <div className="qa-field qa-field--full">
            <div className="qa-label">Climax Summary</div>
            <textarea
              className="qa-textarea"
              rows={2}
              value={template.climaxSummary}
              onChange={(event) =>
                setTemplateField('climaxSummary', event.target.value)
              }
            />
          </div>
          <div className="qa-field qa-field--full">
            <div className="qa-label">Resolution Summary</div>
            <textarea
              className="qa-textarea"
              rows={2}
              value={template.resolutionSummary}
              onChange={(event) =>
                setTemplateField('resolutionSummary', event.target.value)
              }
            />
          </div>
          <div className="qa-field qa-field--full">
            <div className="qa-label">Why It Works</div>
            <textarea
              className="qa-textarea"
              rows={2}
              value={template.whyItWorks}
              onChange={(event) =>
                setTemplateField('whyItWorks', event.target.value)
              }
            />
          </div>
        </div>

        <div className="qa-divider" />

        <div className="qa-card-header">
          <div>
            <h3 className="qa-card-title" style={{ fontSize: 20 }}>
              Beats
            </h3>
            <div className="qa-meta">
              Author the quest beats, subbeats, unlock bundles, and world changes.
            </div>
          </div>
          <button type="button" className="qa-btn qa-btn-primary" onClick={addBeat}>
            Add Beat
          </button>
        </div>

        <div className="qa-stack" style={{ marginTop: 16 }}>
          {template.beats.map((beat, beatIndex) => {
            const isExpanded = expandedBeatIndexes.includes(beatIndex);
            return (
              <div
                key={`beat-${beatIndex}`}
                className="qa-panel"
                style={{
                  background:
                    'linear-gradient(180deg, rgba(15, 23, 28, 0.96), rgba(8, 13, 17, 0.96))',
                }}
              >
                <div className="qa-card-header">
                  <div>
                    <div className="qa-node-title">
                      Beat {beatIndex + 1}: {beat.chapterTitle || 'Untitled Beat'}
                    </div>
                    <div className="qa-meta">
                      Act {beat.act} · {beat.storyRole || 'story beat'} ·{' '}
                      {beat.questArchetypeName || 'no attached quest archetype'}
                    </div>
                  </div>
                  <div className="qa-actions">
                    <button
                      type="button"
                      className="qa-btn qa-btn-outline"
                      onClick={() => moveBeat(beatIndex, -1)}
                      disabled={beatIndex === 0}
                    >
                      Move Up
                    </button>
                    <button
                      type="button"
                      className="qa-btn qa-btn-outline"
                      onClick={() => moveBeat(beatIndex, 1)}
                      disabled={beatIndex === template.beats.length - 1}
                    >
                      Move Down
                    </button>
                    <button
                      type="button"
                      className="qa-btn qa-btn-outline"
                      onClick={() => toggleBeatExpanded(beatIndex)}
                    >
                      {isExpanded ? 'Collapse' : 'Expand'}
                    </button>
                    <button
                      type="button"
                      className="qa-btn qa-btn-danger"
                      onClick={() => removeBeat(beatIndex)}
                      disabled={template.beats.length === 1}
                    >
                      Delete Beat
                    </button>
                  </div>
                </div>

                {isExpanded && (
                  <div style={{ marginTop: 18 }}>
                    <div className="qa-stat-grid">
                      <div className="qa-field">
                        <div className="qa-label">Chapter Title</div>
                        <input
                          className="qa-input"
                          value={beat.chapterTitle}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              chapterTitle: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Act</div>
                        <input
                          className="qa-input"
                          type="number"
                          min={1}
                          value={beat.act}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              act: Number(event.target.value) || 1,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Story Role</div>
                        <input
                          className="qa-input"
                          value={beat.storyRole}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              storyRole: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Quest Archetype</div>
                        <select
                          className="qa-input"
                          value={beat.questArchetypeId || ''}
                          onChange={(event) => {
                            const nextId = event.target.value || null;
                            const selectedArchetype =
                              sortedQuestArchetypes.find(
                                (archetype) => archetype.id === nextId
                              ) ?? null;
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              questArchetypeId: nextId,
                              questArchetypeName: selectedArchetype?.name || '',
                            }));
                          }}
                        >
                          <option value="">No archetype attached</option>
                          {sortedQuestArchetypes.map((archetype) => (
                            <option key={archetype.id} value={archetype.id}>
                              {archetype.name}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">Chapter Summary</div>
                        <textarea
                          className="qa-textarea"
                          rows={2}
                          value={beat.chapterSummary}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              chapterSummary: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">Purpose</div>
                        <textarea
                          className="qa-textarea"
                          rows={2}
                          value={beat.purpose}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              purpose: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">What Changes</div>
                        <textarea
                          className="qa-textarea"
                          rows={2}
                          value={beat.whatChanges}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              whatChanges: event.target.value,
                            }))
                          }
                        />
                      </div>
                    </div>

                    <div className="qa-divider" />

                    <div className="qa-stat-grid">
                      <div className="qa-field">
                        <div className="qa-label">Quest Name</div>
                        <input
                          className="qa-input"
                          value={beat.name}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              name: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Hook</div>
                        <input
                          className="qa-input"
                          value={beat.hook}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              hook: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Quest Giver Key</div>
                        <input
                          className="qa-input"
                          value={beat.questGiverCharacterKey}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              questGiverCharacterKey: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Quest Giver Name</div>
                        <input
                          className="qa-input"
                          value={beat.questGiverCharacterName || ''}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              questGiverCharacterName: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">Description</div>
                        <textarea
                          className="qa-textarea"
                          rows={3}
                          value={beat.description}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              description: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">Acceptance Dialogue</div>
                        <textarea
                          className="qa-textarea"
                          rows={3}
                          value={beat.acceptanceDialogue.join('\n')}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              acceptanceDialogue: parseLines(event.target.value),
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">Aftermath Description</div>
                        <textarea
                          className="qa-textarea"
                          rows={2}
                          value={beat.questGiverAfterDescription || ''}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              questGiverAfterDescription: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">Aftermath Dialogue</div>
                        <textarea
                          className="qa-textarea"
                          rows={2}
                          value={beat.questGiverAfterDialogue.join('\n')}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              questGiverAfterDialogue: parseLines(
                                event.target.value
                              ),
                            }))
                          }
                        />
                      </div>
                    </div>

                    <div className="qa-divider" />

                    <div className="qa-stat-grid">
                      {[
                        ['Required Zone Tags', beat.requiredZoneTags, 'requiredZoneTags'],
                        [
                          'Preferred Content Mix',
                          beat.preferredContentMix,
                          'preferredContentMix',
                        ],
                        ['Character Tags', beat.characterTags, 'characterTags'],
                        ['Internal Tags', beat.internalTags, 'internalTags'],
                        ['Required Story Flags', beat.requiredStoryFlags, 'requiredStoryFlags'],
                        ['Set Story Flags', beat.setStoryFlags, 'setStoryFlags'],
                        ['Clear Story Flags', beat.clearStoryFlags, 'clearStoryFlags'],
                        [
                          'Introduced Character Keys',
                          beat.introducedCharacterKeys,
                          'introducedCharacterKeys',
                        ],
                        [
                          'Required Character Keys',
                          beat.requiredCharacterKeys,
                          'requiredCharacterKeys',
                        ],
                        [
                          'Introduced Reveal Keys',
                          beat.introducedRevealKeys,
                          'introducedRevealKeys',
                        ],
                        [
                          'Required Reveal Keys',
                          beat.requiredRevealKeys,
                          'requiredRevealKeys',
                        ],
                        [
                          'Challenge Seeds',
                          beat.challengeTemplateSeeds,
                          'challengeTemplateSeeds',
                        ],
                        [
                          'Scenario Seeds',
                          beat.scenarioTemplateSeeds,
                          'scenarioTemplateSeeds',
                        ],
                        [
                          'Monster Seeds',
                          beat.monsterTemplateSeeds,
                          'monsterTemplateSeeds',
                        ],
                      ].map(([label, value, key]) => (
                        <div className="qa-field" key={`${beatIndex}-${key}`}>
                          <div className="qa-label">{label}</div>
                          <input
                            className="qa-input"
                            value={(value as string[]).join(', ')}
                            onChange={(event) =>
                              updateBeat(beatIndex, (current) => {
                                const nextBeat = {
                                  ...current,
                                } as MainStoryBeatDraft & Record<string, string[]>;
                                nextBeat[key as string] = parseList(
                                  event.target.value
                                );
                                return nextBeat;
                              })
                            }
                          />
                        </div>
                      ))}
                      <div className="qa-field">
                        <div className="qa-label">Required Location Archetypes</div>
                        <select
                          className="qa-input"
                          value=""
                          onChange={(event) => {
                            const value = event.target.value;
                            if (!value) return;
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              requiredLocationArchetypeIds: current.requiredLocationArchetypeIds.includes(
                                value
                              )
                                ? current.requiredLocationArchetypeIds
                                : [...current.requiredLocationArchetypeIds, value],
                            }));
                            event.target.value = '';
                          }}
                        >
                          <option value="">Add location archetype...</option>
                          {locationArchetypes.map((archetype) => (
                            <option key={archetype.id} value={archetype.id}>
                              {archetype.name}
                            </option>
                          ))}
                        </select>
                        <div className="qa-tag-row" style={{ marginTop: 10 }}>
                          {beat.requiredLocationArchetypeIds.map((id) => {
                            const archetype =
                              locationArchetypes.find((entry) => entry.id === id) ??
                              null;
                            return (
                              <button
                                key={id}
                                type="button"
                                className="qa-chip"
                                onClick={() =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    requiredLocationArchetypeIds:
                                      current.requiredLocationArchetypeIds.filter(
                                        (entry) => entry !== id
                                      ),
                                  }))
                                }
                              >
                                {archetype?.name || id} x
                              </button>
                            );
                          })}
                        </div>
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Difficulty Mode</div>
                        <select
                          className="qa-input"
                          value={beat.difficultyMode}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              difficultyMode: event.target
                                .value as MainStoryBeatDraft['difficultyMode'],
                            }))
                          }
                        >
                          <option value="scale">Scale</option>
                          <option value="fixed">Fixed</option>
                        </select>
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Difficulty</div>
                        <input
                          className="qa-input"
                          type="number"
                          min={1}
                          value={beat.difficulty}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              difficulty: Number(event.target.value) || 1,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field">
                        <div className="qa-label">Monster Target Level</div>
                        <input
                          className="qa-input"
                          type="number"
                          min={1}
                          value={beat.monsterEncounterTargetLevel}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              monsterEncounterTargetLevel:
                                Number(event.target.value) || 1,
                            }))
                          }
                        />
                      </div>
                      <div className="qa-field qa-field--full">
                        <div className="qa-label">Why This Scales</div>
                        <textarea
                          className="qa-textarea"
                          rows={2}
                          value={beat.whyThisScales}
                          onChange={(event) =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              whyThisScales: event.target.value,
                            }))
                          }
                        />
                      </div>
                    </div>

                    <div className="qa-divider" />

                    <div className="qa-card-header">
                      <div>
                        <h4 className="qa-card-title" style={{ fontSize: 16 }}>
                          Relationship Effects
                        </h4>
                      </div>
                    </div>
                    <div className="qa-stat-grid">
                      {(['trust', 'respect', 'fear', 'debt'] as const).map(
                        (key) => (
                          <div className="qa-field" key={`${beatIndex}-${key}`}>
                            <div className="qa-label">{key}</div>
                            <input
                              className="qa-input"
                              type="number"
                              value={beat.questGiverRelationshipEffects?.[key] || 0}
                              onChange={(event) =>
                                updateBeat(beatIndex, (current) => ({
                                  ...current,
                                  questGiverRelationshipEffects: {
                                    ...(current.questGiverRelationshipEffects || {}),
                                    [key]: Number(event.target.value) || 0,
                                  },
                                }))
                              }
                            />
                          </div>
                        )
                      )}
                    </div>

                    <div className="qa-divider" />

                    <div className="qa-card-header">
                      <div>
                        <h4 className="qa-card-title" style={{ fontSize: 16 }}>
                          Steps
                        </h4>
                      </div>
                      <button
                        type="button"
                        className="qa-btn qa-btn-outline"
                        onClick={() =>
                          updateBeat(beatIndex, (current) => ({
                            ...current,
                            steps: [
                              ...current.steps,
                              {
                                source: 'location',
                                content: 'challenge',
                                locationConcept: '',
                                locationMetadataTags: [],
                                templateConcept: '',
                                potentialContent: [],
                                challengeStatTags: [],
                                scenarioBeats: [],
                                monsterTemplateNames: [],
                                monsterTemplateIds: [],
                                encounterTone: [],
                              },
                            ],
                          }))
                        }
                      >
                        Add Step
                      </button>
                    </div>
                    <div className="qa-stack">
                      {beat.steps.map((step, stepIndex) => (
                        <div key={`${beatIndex}-step-${stepIndex}`} className="qa-route-branch-card">
                          <div className="qa-card-header">
                            <div className="qa-node-title">Step {stepIndex + 1}</div>
                            <button
                              type="button"
                              className="qa-btn qa-btn-danger"
                              onClick={() =>
                                updateBeat(beatIndex, (current) => ({
                                  ...current,
                                  steps: current.steps.filter(
                                    (_, index) => index !== stepIndex
                                  ),
                                }))
                              }
                            >
                              Remove Step
                            </button>
                          </div>
                          <div className="qa-stat-grid">
                            <div className="qa-field">
                              <div className="qa-label">Source</div>
                              <select
                                className="qa-input"
                                value={step.source}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? { ...entry, source: event.target.value }
                                        : entry
                                    ),
                                  }))
                                }
                              >
                                <option value="location">Location</option>
                                <option value="proximity">Proximity</option>
                              </select>
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Content</div>
                              <select
                                className="qa-input"
                                value={step.content}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? { ...entry, content: event.target.value }
                                        : entry
                                    ),
                                  }))
                                }
                              >
                                <option value="challenge">Challenge</option>
                                <option value="scenario">Scenario</option>
                                <option value="monster">Monster</option>
                              </select>
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Location Concept</div>
                              <input
                                className="qa-input"
                                value={step.locationConcept}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? { ...entry, locationConcept: event.target.value }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Location Archetype</div>
                              <select
                                className="qa-input"
                                value={step.locationArchetypeId || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            locationArchetypeId:
                                              event.target.value || null,
                                            locationArchetypeName:
                                              locationArchetypes.find(
                                                (archetype) =>
                                                  archetype.id === event.target.value
                                              )?.name || '',
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              >
                                <option value="">No archetype</option>
                                {locationArchetypes.map((archetype) => (
                                  <option key={archetype.id} value={archetype.id}>
                                    {archetype.name}
                                  </option>
                                ))}
                              </select>
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Distance Meters</div>
                              <input
                                className="qa-input"
                                type="number"
                                min={0}
                                value={step.distanceMeters ?? ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            distanceMeters:
                                              event.target.value === ''
                                                ? null
                                                : Number(event.target.value) || 0,
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field qa-field--full">
                              <div className="qa-label">Template Concept</div>
                              <textarea
                                className="qa-textarea"
                                rows={2}
                                value={step.templateConcept}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? { ...entry, templateConcept: event.target.value }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Location Metadata Tags</div>
                              <input
                                className="qa-input"
                                value={step.locationMetadataTags.join(', ')}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            locationMetadataTags: parseList(
                                              event.target.value
                                            ),
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Potential Content</div>
                              <input
                                className="qa-input"
                                value={step.potentialContent.join(', ')}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            potentialContent: parseList(
                                              event.target.value
                                            ),
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Challenge Question</div>
                              <input
                                className="qa-input"
                                value={step.challengeQuestion || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            challengeQuestion: event.target.value,
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Challenge Description</div>
                              <input
                                className="qa-input"
                                value={step.challengeDescription || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            challengeDescription:
                                              event.target.value,
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Scenario Prompt</div>
                              <input
                                className="qa-input"
                                value={step.scenarioPrompt || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            scenarioPrompt: event.target.value,
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Monster Template Names</div>
                              <input
                                className="qa-input"
                                value={(step.monsterTemplateNames || []).join(', ')}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    steps: current.steps.map((entry, index) =>
                                      index === stepIndex
                                        ? {
                                            ...entry,
                                            monsterTemplateNames: parseList(
                                              event.target.value
                                            ),
                                          }
                                        : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>

                    <div className="qa-divider" />

                    <div className="qa-card-header">
                      <div>
                        <h4 className="qa-card-title" style={{ fontSize: 16 }}>
                          World Changes
                        </h4>
                      </div>
                      <button
                        type="button"
                        className="qa-btn qa-btn-outline"
                        onClick={() =>
                          updateBeat(beatIndex, (current) => ({
                            ...current,
                            worldChanges: [
                              ...current.worldChanges,
                              {
                                type: 'move_character',
                                targetKey: '',
                                characterKey: '',
                                pointOfInterestHint: '',
                                destinationHint: '',
                                zoneTags: [],
                                description: '',
                                clue: '',
                              },
                            ],
                          }))
                        }
                      >
                        Add World Change
                      </button>
                    </div>
                    <div className="qa-stack">
                      {beat.worldChanges.map((change, changeIndex) => (
                        <div key={`${beatIndex}-world-${changeIndex}`} className="qa-route-branch-card">
                          <div className="qa-card-header">
                            <div className="qa-node-title">
                              World Change {changeIndex + 1}
                            </div>
                            <button
                              type="button"
                              className="qa-btn qa-btn-danger"
                              onClick={() =>
                                updateBeat(beatIndex, (current) => ({
                                  ...current,
                                  worldChanges: current.worldChanges.filter(
                                    (_, index) => index !== changeIndex
                                  ),
                                }))
                              }
                            >
                              Remove
                            </button>
                          </div>
                          <div className="qa-stat-grid">
                            <div className="qa-field">
                              <div className="qa-label">Type</div>
                              <select
                                className="qa-input"
                                value={change.type}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    worldChanges: current.worldChanges.map(
                                      (entry, index) =>
                                        index === changeIndex
                                          ? { ...entry, type: event.target.value as typeof entry.type }
                                          : entry
                                    ),
                                  }))
                                }
                              >
                                <option value="move_character">Move Character</option>
                                <option value="show_poi_text">Show POI Text</option>
                              </select>
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Target Key</div>
                              <input
                                className="qa-input"
                                value={change.targetKey}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    worldChanges: current.worldChanges.map(
                                      (entry, index) =>
                                        index === changeIndex
                                          ? { ...entry, targetKey: event.target.value }
                                          : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Character Key</div>
                              <input
                                className="qa-input"
                                value={change.characterKey || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    worldChanges: current.worldChanges.map(
                                      (entry, index) =>
                                        index === changeIndex
                                          ? { ...entry, characterKey: event.target.value }
                                          : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Destination Hint</div>
                              <input
                                className="qa-input"
                                value={change.destinationHint || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    worldChanges: current.worldChanges.map(
                                      (entry, index) =>
                                        index === changeIndex
                                          ? { ...entry, destinationHint: event.target.value }
                                          : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">POI Hint</div>
                              <input
                                className="qa-input"
                                value={change.pointOfInterestHint || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    worldChanges: current.worldChanges.map(
                                      (entry, index) =>
                                        index === changeIndex
                                          ? {
                                              ...entry,
                                              pointOfInterestHint: event.target.value,
                                            }
                                          : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Zone Tags</div>
                              <input
                                className="qa-input"
                                value={(change.zoneTags || []).join(', ')}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    worldChanges: current.worldChanges.map(
                                      (entry, index) =>
                                        index === changeIndex
                                          ? { ...entry, zoneTags: parseList(event.target.value) }
                                          : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field qa-field--full">
                              <div className="qa-label">Description</div>
                              <textarea
                                className="qa-textarea"
                                rows={2}
                                value={change.description || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    worldChanges: current.worldChanges.map(
                                      (entry, index) =>
                                        index === changeIndex
                                          ? { ...entry, description: event.target.value }
                                          : entry
                                    ),
                                  }))
                                }
                              />
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>

                    <div className="qa-divider" />

                    <div className="qa-card-header">
                      <div>
                        <h4 className="qa-card-title" style={{ fontSize: 16 }}>
                          Unlocked Content
                        </h4>
                      </div>
                      <div className="qa-actions">
                        <button
                          type="button"
                          className="qa-btn qa-btn-outline"
                          onClick={() =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              unlockedScenarios: [
                                ...current.unlockedScenarios,
                                {
                                  name: '',
                                  prompt: '',
                                  pointOfInterestHint: '',
                                  internalTags: [],
                                  difficulty: current.difficulty,
                                },
                              ],
                            }))
                          }
                        >
                          Add Scenario
                        </button>
                        <button
                          type="button"
                          className="qa-btn qa-btn-outline"
                          onClick={() =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              unlockedChallenges: [
                                ...current.unlockedChallenges,
                                {
                                  question: '',
                                  description: '',
                                  pointOfInterestHint: '',
                                  submissionType: 'text',
                                  proficiency: null,
                                  statTags: [],
                                  difficulty: current.difficulty,
                                },
                              ],
                            }))
                          }
                        >
                          Add Challenge
                        </button>
                        <button
                          type="button"
                          className="qa-btn qa-btn-outline"
                          onClick={() =>
                            updateBeat(beatIndex, (current) => ({
                              ...current,
                              unlockedMonsterEncounters: [
                                ...current.unlockedMonsterEncounters,
                                {
                                  name: '',
                                  description: '',
                                  pointOfInterestHint: '',
                                  encounterType: 'monster',
                                  monsterCount: 1,
                                  encounterTone: [],
                                  monsterTemplateHints: [],
                                  targetLevel: current.monsterEncounterTargetLevel,
                                },
                              ],
                            }))
                          }
                        >
                          Add Encounter
                        </button>
                      </div>
                    </div>

                    <div className="qa-stack">
                      {beat.unlockedScenarios.map((scenario, scenarioIndex) => (
                        <div key={`${beatIndex}-scenario-${scenarioIndex}`} className="qa-route-branch-card">
                          <div className="qa-card-header">
                            <div className="qa-node-title">
                              Unlocked Scenario {scenarioIndex + 1}
                            </div>
                            <button
                              type="button"
                              className="qa-btn qa-btn-danger"
                              onClick={() =>
                                updateBeat(beatIndex, (current) => ({
                                  ...current,
                                  unlockedScenarios:
                                    current.unlockedScenarios.filter(
                                      (_, index) => index !== scenarioIndex
                                    ),
                                }))
                              }
                            >
                              Remove
                            </button>
                          </div>
                          <div className="qa-stat-grid">
                            <div className="qa-field">
                              <div className="qa-label">Name</div>
                              <input
                                className="qa-input"
                                value={scenario.name}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedScenarios:
                                      current.unlockedScenarios.map((entry, index) =>
                                        index === scenarioIndex
                                          ? { ...entry, name: event.target.value }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Difficulty</div>
                              <input
                                className="qa-input"
                                type="number"
                                min={1}
                                value={scenario.difficulty || 1}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedScenarios:
                                      current.unlockedScenarios.map((entry, index) =>
                                        index === scenarioIndex
                                          ? {
                                              ...entry,
                                              difficulty:
                                                Number(event.target.value) || 1,
                                            }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field qa-field--full">
                              <div className="qa-label">Prompt</div>
                              <textarea
                                className="qa-textarea"
                                rows={2}
                                value={scenario.prompt}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedScenarios:
                                      current.unlockedScenarios.map((entry, index) =>
                                        index === scenarioIndex
                                          ? { ...entry, prompt: event.target.value }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                          </div>
                        </div>
                      ))}

                      {beat.unlockedChallenges.map((challenge, challengeIndex) => (
                        <div key={`${beatIndex}-challenge-${challengeIndex}`} className="qa-route-branch-card">
                          <div className="qa-card-header">
                            <div className="qa-node-title">
                              Unlocked Challenge {challengeIndex + 1}
                            </div>
                            <button
                              type="button"
                              className="qa-btn qa-btn-danger"
                              onClick={() =>
                                updateBeat(beatIndex, (current) => ({
                                  ...current,
                                  unlockedChallenges:
                                    current.unlockedChallenges.filter(
                                      (_, index) => index !== challengeIndex
                                    ),
                                }))
                              }
                            >
                              Remove
                            </button>
                          </div>
                          <div className="qa-stat-grid">
                            <div className="qa-field">
                              <div className="qa-label">Question</div>
                              <input
                                className="qa-input"
                                value={challenge.question}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedChallenges:
                                      current.unlockedChallenges.map((entry, index) =>
                                        index === challengeIndex
                                          ? { ...entry, question: event.target.value }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Submission Type</div>
                              <input
                                className="qa-input"
                                value={challenge.submissionType || 'text'}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedChallenges:
                                      current.unlockedChallenges.map((entry, index) =>
                                        index === challengeIndex
                                          ? {
                                              ...entry,
                                              submissionType: event.target.value as typeof entry.submissionType,
                                            }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Proficiency</div>
                              <input
                                className="qa-input"
                                value={challenge.proficiency || ''}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedChallenges:
                                      current.unlockedChallenges.map((entry, index) =>
                                        index === challengeIndex
                                          ? {
                                              ...entry,
                                              proficiency:
                                                event.target.value || null,
                                            }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Stat Tags</div>
                              <input
                                className="qa-input"
                                value={(challenge.statTags || []).join(', ')}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedChallenges:
                                      current.unlockedChallenges.map((entry, index) =>
                                        index === challengeIndex
                                          ? {
                                              ...entry,
                                              statTags: parseList(event.target.value),
                                            }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field qa-field--full">
                              <div className="qa-label">Description</div>
                              <textarea
                                className="qa-textarea"
                                rows={2}
                                value={challenge.description}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedChallenges:
                                      current.unlockedChallenges.map((entry, index) =>
                                        index === challengeIndex
                                          ? { ...entry, description: event.target.value }
                                          : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                          </div>
                        </div>
                      ))}

                      {beat.unlockedMonsterEncounters.map((encounter, encounterIndex) => (
                        <div key={`${beatIndex}-encounter-${encounterIndex}`} className="qa-route-branch-card">
                          <div className="qa-card-header">
                            <div className="qa-node-title">
                              Unlocked Encounter {encounterIndex + 1}
                            </div>
                            <button
                              type="button"
                              className="qa-btn qa-btn-danger"
                              onClick={() =>
                                updateBeat(beatIndex, (current) => ({
                                  ...current,
                                  unlockedMonsterEncounters:
                                    current.unlockedMonsterEncounters.filter(
                                      (_, index) => index !== encounterIndex
                                    ),
                                }))
                              }
                            >
                              Remove
                            </button>
                          </div>
                          <div className="qa-stat-grid">
                            <div className="qa-field">
                              <div className="qa-label">Name</div>
                              <input
                                className="qa-input"
                                value={encounter.name}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedMonsterEncounters:
                                      current.unlockedMonsterEncounters.map(
                                        (entry, index) =>
                                          index === encounterIndex
                                            ? { ...entry, name: event.target.value }
                                            : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Encounter Type</div>
                              <input
                                className="qa-input"
                                value={encounter.encounterType || 'monster'}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedMonsterEncounters:
                                      current.unlockedMonsterEncounters.map(
                                        (entry, index) =>
                                          index === encounterIndex
                                            ? {
                                                ...entry,
                                                encounterType:
                                                  event.target.value as typeof entry.encounterType,
                                              }
                                            : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Monster Count</div>
                              <input
                                className="qa-input"
                                type="number"
                                min={1}
                                value={encounter.monsterCount || 1}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedMonsterEncounters:
                                      current.unlockedMonsterEncounters.map(
                                        (entry, index) =>
                                          index === encounterIndex
                                            ? {
                                                ...entry,
                                                monsterCount:
                                                  Number(event.target.value) || 1,
                                              }
                                            : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Target Level</div>
                              <input
                                className="qa-input"
                                type="number"
                                min={1}
                                value={encounter.targetLevel || 1}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedMonsterEncounters:
                                      current.unlockedMonsterEncounters.map(
                                        (entry, index) =>
                                          index === encounterIndex
                                            ? {
                                                ...entry,
                                                targetLevel:
                                                  Number(event.target.value) || 1,
                                              }
                                            : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Encounter Tone</div>
                              <input
                                className="qa-input"
                                value={(encounter.encounterTone || []).join(', ')}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedMonsterEncounters:
                                      current.unlockedMonsterEncounters.map(
                                        (entry, index) =>
                                          index === encounterIndex
                                            ? {
                                                ...entry,
                                                encounterTone: parseList(
                                                  event.target.value
                                                ),
                                              }
                                            : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field">
                              <div className="qa-label">Monster Hints</div>
                              <input
                                className="qa-input"
                                value={(encounter.monsterTemplateHints || []).join(
                                  ', '
                                )}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedMonsterEncounters:
                                      current.unlockedMonsterEncounters.map(
                                        (entry, index) =>
                                          index === encounterIndex
                                            ? {
                                                ...entry,
                                                monsterTemplateHints: parseList(
                                                  event.target.value
                                                ),
                                              }
                                            : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                            <div className="qa-field qa-field--full">
                              <div className="qa-label">Description</div>
                              <textarea
                                className="qa-textarea"
                                rows={2}
                                value={encounter.description}
                                onChange={(event) =>
                                  updateBeat(beatIndex, (current) => ({
                                    ...current,
                                    unlockedMonsterEncounters:
                                      current.unlockedMonsterEncounters.map(
                                        (entry, index) =>
                                          index === encounterIndex
                                            ? {
                                                ...entry,
                                                description: event.target.value,
                                              }
                                            : entry
                                      ),
                                  }))
                                }
                              />
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default MainStoryTemplateEditor;
