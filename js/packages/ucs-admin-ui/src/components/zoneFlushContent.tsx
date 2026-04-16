import React from 'react';

export type ZoneFlushContentType =
  | 'pointsOfInterest'
  | 'quests'
  | 'challenges'
  | 'scenarios'
  | 'expositions'
  | 'monsters'
  | 'treasureChests'
  | 'healingFountains'
  | 'resources'
  | 'movementPatterns'
  | 'jobs';

export type ZoneContentFlushSummary = {
  zoneCount: number;
  deletedPointOfInterestCount: number;
  detachedSharedPointOfInterestCount: number;
  deletedCharacterCount: number;
  deletedQuestCount: number;
  deletedZoneQuestArchetypeCount: number;
  deletedChallengeCount: number;
  deletedScenarioCount: number;
  deletedExpositionCount: number;
  deletedMonsterEncounterCount: number;
  deletedMonsterCount: number;
  deletedTreasureChestCount: number;
  deletedHealingFountainCount: number;
  deletedResourceCount: number;
  deletedMovementPatternCount: number;
  deletedZoneSeedJobCount: number;
  deletedQuestGenerationJobCount: number;
  deletedScenarioGenerationJobCount: number;
  deletedChallengeGenerationJobCount: number;
};

type ZoneFlushContentOption = {
  value: ZoneFlushContentType;
  label: string;
  description: string;
};

export const zoneFlushContentOptions: ZoneFlushContentOption[] = [
  {
    value: 'pointsOfInterest',
    label: 'POIs & NPCs',
    description:
      'Removes zone-linked POIs, deletes matching zone NPCs, and detaches shared POIs from these zones.',
  },
  {
    value: 'quests',
    label: 'Quests & archetypes',
    description:
      'Deletes live quests plus zone quest archetype assignments so generation can start fresh.',
  },
  {
    value: 'challenges',
    label: 'Challenges',
    description: 'Deletes all zone challenges.',
  },
  {
    value: 'scenarios',
    label: 'Scenarios',
    description: 'Deletes all zone scenarios.',
  },
  {
    value: 'expositions',
    label: 'Expositions',
    description: 'Deletes all zone expositions.',
  },
  {
    value: 'monsters',
    label: 'Monsters & encounters',
    description:
      'Deletes monster encounters plus the monsters authored directly in these zones.',
  },
  {
    value: 'treasureChests',
    label: 'Treasure chests',
    description: 'Deletes all zone treasure chests.',
  },
  {
    value: 'healingFountains',
    label: 'Healing fountains',
    description: 'Deletes all zone healing fountains.',
  },
  {
    value: 'resources',
    label: 'Resources',
    description: 'Deletes all zone resources.',
  },
  {
    value: 'movementPatterns',
    label: 'Movement patterns',
    description: 'Deletes all zone movement patterns.',
  },
  {
    value: 'jobs',
    label: 'Seed & generation jobs',
    description:
      'Deletes zone seed jobs plus quest, scenario, and challenge generation jobs.',
  },
];

export const defaultZoneFlushContentTypes: ZoneFlushContentType[] =
  zoneFlushContentOptions.map((option) => option.value);

export const formatZoneContentFlushSummary = (
  summary: ZoneContentFlushSummary
) => {
  const deletedJobCount =
    summary.deletedZoneSeedJobCount +
    summary.deletedQuestGenerationJobCount +
    summary.deletedScenarioGenerationJobCount +
    summary.deletedChallengeGenerationJobCount;

  const entries = [
    ['POIs', summary.deletedPointOfInterestCount],
    ['shared POI links', summary.detachedSharedPointOfInterestCount],
    ['characters', summary.deletedCharacterCount],
    ['quests', summary.deletedQuestCount],
    ['archetypes', summary.deletedZoneQuestArchetypeCount],
    ['challenges', summary.deletedChallengeCount],
    ['scenarios', summary.deletedScenarioCount],
    ['expositions', summary.deletedExpositionCount],
    ['encounters', summary.deletedMonsterEncounterCount],
    ['monsters', summary.deletedMonsterCount],
    ['chests', summary.deletedTreasureChestCount],
    ['fountains', summary.deletedHealingFountainCount],
    ['resources', summary.deletedResourceCount],
    ['movement patterns', summary.deletedMovementPatternCount],
    ['jobs', deletedJobCount],
  ].filter((entry) => entry[1] > 0);

  if (entries.length === 0) {
    return 'No removable zone content was found.';
  }

  return entries
    .slice(0, 6)
    .map(([label, count]) => `${count} ${label}`)
    .join(', ');
};

type ZoneFlushContentModalProps = {
  isOpen: boolean;
  zoneCount: number;
  selectedContentTypes: ZoneFlushContentType[];
  onToggleContentType: (contentType: ZoneFlushContentType) => void;
  onSelectAll: () => void;
  onClearAll: () => void;
  onCancel: () => void;
  onConfirm: () => void;
  isSubmitting?: boolean;
  error?: string | null;
};

export const ZoneFlushContentModal: React.FC<ZoneFlushContentModalProps> = ({
  isOpen,
  zoneCount,
  selectedContentTypes,
  onToggleContentType,
  onSelectAll,
  onClearAll,
  onCancel,
  onConfirm,
  isSubmitting = false,
  error = null,
}) => {
  if (!isOpen) {
    return null;
  }

  const selectedTypeSet = new Set(selectedContentTypes);
  const zoneLabel = zoneCount === 1 ? 'this zone' : `${zoneCount} zones`;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4">
      <div className="w-full max-w-3xl rounded-2xl bg-white shadow-2xl">
        <div className="border-b border-slate-200 px-6 py-5">
          <div className="text-xs font-semibold uppercase tracking-[0.2em] text-amber-700">
            Flush Zone Content
          </div>
          <h2 className="mt-2 text-xl font-semibold text-slate-900">
            Choose what to clear from {zoneLabel}
          </h2>
          <p className="mt-2 text-sm leading-6 text-slate-600">
            The zone records, boundaries, descriptions, and tags stay intact.
            Only the selected zone-scoped content will be removed.
          </p>
        </div>

        <div className="px-6 py-5">
          <div className="mb-4 flex items-center justify-between gap-3">
            <div className="text-sm text-slate-500">
              {selectedContentTypes.length} of {zoneFlushContentOptions.length}{' '}
              content types selected
            </div>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={onSelectAll}
                className="rounded-md border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50"
              >
                Select All
              </button>
              <button
                type="button"
                onClick={onClearAll}
                className="rounded-md border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50"
              >
                Clear All
              </button>
            </div>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            {zoneFlushContentOptions.map((option) => {
              const isChecked = selectedTypeSet.has(option.value);
              return (
                <label
                  key={option.value}
                  className={`flex cursor-pointer gap-3 rounded-xl border px-4 py-4 transition ${
                    isChecked
                      ? 'border-amber-300 bg-amber-50/70'
                      : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50'
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={isChecked}
                    onChange={() => onToggleContentType(option.value)}
                    className="mt-1 h-4 w-4 rounded border-slate-300 text-amber-600 focus:ring-amber-500"
                  />
                  <div>
                    <div className="text-sm font-semibold text-slate-900">
                      {option.label}
                    </div>
                    <div className="mt-1 text-sm leading-6 text-slate-600">
                      {option.description}
                    </div>
                  </div>
                </label>
              );
            })}
          </div>

          {error && (
            <div className="mt-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {error}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between gap-3 border-t border-slate-200 px-6 py-4">
          <div className="text-sm text-slate-500">
            Shared POIs are detached from the selected zones instead of being
            deleted outright.
          </div>
          <div className="flex gap-3">
            <button
              type="button"
              onClick={onCancel}
              className="rounded-md border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={onConfirm}
              disabled={isSubmitting || selectedContentTypes.length === 0}
              className="rounded-md bg-amber-600 px-4 py-2 text-sm font-semibold text-white hover:bg-amber-700 disabled:bg-amber-300"
            >
              {isSubmitting
                ? 'Flushing...'
                : zoneCount === 1
                  ? 'Flush Selected Content'
                  : `Flush ${zoneCount} Zones`}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
