import {
  ItemType,
  PointOfInterest,
  Team,
  getHighestFirstCompletedChallenge,
  hasDiscoveredPointOfInterest,
  InventoryItem,
  MatchInventoryItemEffect,
  OwnedInventoryItem,
  User,
  ItemsUsabledOnPointOfInterest,
  Match,
} from '@poltergeist/types';
import React, { useState } from 'react';
import { useMatchContext } from '../../contexts/MatchContext.tsx';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { TabItem, TabNav } from '../shared/TabNav.tsx';
import { SubmitAnswerForChallenge } from './SubmitAnswerForChallenge.tsx';
import { Button } from '../shared/Button.tsx';
import { LockClosedIcon, LockOpenIcon } from '@heroicons/react/20/solid';
import { StatusIndicator } from '../shared/StatusIndicator.tsx';
import { useInventory, useLocation } from '@poltergeist/contexts';
import { scrambleAndObscureWords } from '../../utils/scrambleSentences.ts';
import { toRoman } from '../../utils/toRoman.ts';
import { mapCaptureTiers } from '../../utils/mapCaptureTiers.ts';
import { usePointOfInterestContext } from '../../contexts/PointOfInterestContext.tsx';
import { useDiscoveriesContext } from '../../contexts/DiscoveriesContext.tsx';
import { useSubmissionsContext } from '../../contexts/SubmissionsContext.tsx';
import { useUserProfiles } from '../../contexts/UserProfileContext.tsx';
import { useQuestLogContext } from '../../contexts/QuestLogContext.tsx';
import { useTagContext } from '@poltergeist/contexts';
import tagFilter from '../../utils/tagFilter.ts';
import { Modal, ModalSize } from '../shared/Modal.tsx';
import { PointOfInterestTasks } from './PointOfInterestTasks.tsx';
import { PointOfInterestCompletedTasks } from './PointOfInterestCompletedTasks.tsx';

interface PointOfInterestPanelProps {
  pointOfInterest: PointOfInterest;
  onClose: (immediate: boolean) => void;
  match?: Match | null;
  usersTeam?: Team | undefined;
}

export const PointOfInterestPanel = ({
  pointOfInterest,
  onClose,
  match,
  usersTeam,
}: PointOfInterestPanelProps) => {
  const { discoveries, setDiscoveries } = useDiscoveriesContext();
  const { isRootNode } = useQuestLogContext();
  const { submissions } = useSubmissionsContext();
  const { currentUser } = useUserProfiles();
  const { discoverPointOfInterest } = useDiscoveriesContext();
  const completedForTier = mapCaptureTiers(pointOfInterest, submissions);
  const { inventoryItems, consumeItem, setUsedItem, ownedInventoryItems } =
    useInventory();
  const { location } = useLocation();
  const [buttonText, setButtonText] = useState<string>("I'm here!");
  const { tagGroups } = useTagContext();
  const [showAllTags, setShowAllTags] = useState(false);
  const tagGroup = tagGroups?.find((group) =>
    group.tags?.some((tag) =>
      pointOfInterest.tags?.some((t) => t.id === tag.id)
    )
  );
  const isGoldenMonkeyActive = match?.inventoryItemEffects.some(
    (item) =>
      item.inventoryItemId === ItemType.CipherOfTheLaughingMonkey &&
      item.teamId !== usersTeam?.id &&
      new Date(item.expiresAt) > new Date()
  );

  const hasDiscovered = hasDiscoveredPointOfInterest(
    pointOfInterest.id,
    usersTeam ? usersTeam.id : (currentUser?.id ?? ''),
    discoveries ?? []
  );

  const { submission, challenge } = getHighestFirstCompletedChallenge(
    pointOfInterest,
    submissions
  );

  var capturingEntityName = '';
  if (submission?.teamId) {
    capturingEntityName =
      match?.teams.find((team) => team.id === submission.teamId)?.name ??
      'Unknown';
  } else if (submission?.userId) {
    capturingEntityName = currentUser?.name ?? 'Unknown';
  }

  var captureTier: number | null = null;
  if (challenge) {
    captureTier = challenge.tier;
  }

  const onConsumeItem = async (
    ownedInventoryItemId: string,
    itemId: number
  ) => {
    await consumeItem(ownedInventoryItemId, {
      pointOfInterestId: pointOfInterest.id,
    });
    setUsedItem(inventoryItems.find((item) => item.id === itemId) ?? null);
    onClose(true);
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="flex flex-col items-center gap-2">
        <h3 className="text-2xl font-bold">{pointOfInterest.name}</h3>
        {pointOfInterest.originalName && hasDiscovered && (
          <h4 className="text-gray-500 text-xs">
            ({pointOfInterest.originalName})
            {pointOfInterest.googleMapsPlaceId && (
              <a
                href={`https://www.google.com/maps/place/?q=place_id:${pointOfInterest.googleMapsPlaceId}`}
                target="_blank"
                rel="noopener noreferrer"
                className="ml-2 text-blue-500 hover:text-blue-700"
              >
                View on Google Maps
              </a>
            )}
          </h4>
        )}
      </div>
      <img
        src={
          hasDiscovered
            ? pointOfInterest.imageURL
            : `https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp`
        }
        alt={pointOfInterest.name}
      />
      <div className="flex justify-between items-center w-full">
        {usersTeam && (
          <div className="flex-shrink-0">
          <StatusIndicator
            capturingEntityName={capturingEntityName}
            captureTier={captureTier}
            match={match}
            usersTeam={usersTeam}
            />
          </div>
        )}
        <div className={`flex flex-wrap gap-2 justify-end ${usersTeam ? '' : 'w-full'}`}>
          {tagFilter(pointOfInterest.tags ?? [])
            .slice(0, usersTeam ? 1 : 2)
            .map((tag) => (
              <div
                key={tag.id}
                className="px-3 py-1 bg-gray-200 rounded-full text-sm"
              >
                {tag.name
                  .split('_')
                  .map((word, i) =>
                    i === 0
                      ? word.charAt(0).toUpperCase() +
                        word.slice(1).toLowerCase()
                      : word.toLowerCase()
                  )
                  .join(' ')}
              </div>
            ))}
          {tagFilter(pointOfInterest.tags ?? []).length > (usersTeam ? 1 : 2) && (
            <div
              className="px-3 py-1 bg-gray-200 rounded-full text-sm cursor-pointer hover:bg-gray-300"
              onClick={() => setShowAllTags(true)}
            >
              +{tagFilter(pointOfInterest.tags ?? []).length - 1} more
            </div>
          )}
        </div>
      </div>
      {showAllTags && (
        <Modal size={ModalSize.HERO}>
          <h3 className="text-xl font-bold mb-4">All Tags</h3>
          <div className="flex flex-wrap gap-2">
            {tagFilter(pointOfInterest.tags ?? []).map((tag) => (
              <div
                key={tag.id}
                className="px-3 py-1 bg-gray-200 rounded-full text-sm"
              >
                {tag.name
                  .split('_')
                  .map((word, i) =>
                    i === 0
                      ? word.charAt(0).toUpperCase() +
                        word.slice(1).toLowerCase()
                      : word.toLowerCase()
                  )
                  .join(' ')}
              </div>
            ))}
          </div>
          <button
            className="mt-4 px-4 py-2 bg-gray-200 rounded-lg hover:bg-gray-300"
            onClick={() => setShowAllTags(false)}
          >
            Close
          </button>
        </Modal>
      )}
      {hasDiscovered && (
        <TabNav
          tabs={[
            'Info',
            'Tasks',
            'Completed Tasks',
          ]}
        >
          <TabItem key="Info">
            <p className="text-md text-left">{pointOfInterest.description}</p>
          </TabItem>
          <TabItem key="Tasks">
            <PointOfInterestTasks pointOfInterest={pointOfInterest} onClose={onClose} />
          </TabItem>
          <TabItem key="Completed Tasks">
            <PointOfInterestCompletedTasks pointOfInterest={pointOfInterest} />
          </TabItem>
        </TabNav>
      )}
      {!hasDiscovered && (
        <p className="text-xl text-left">
          <span className="font-bold">Clue:</span>{' '}
          {isGoldenMonkeyActive
            ? scrambleAndObscureWords(pointOfInterest.clue, usersTeam?.id ?? '')
            : pointOfInterest.clue}
        </p>
      )}
      {!hasDiscovered && (
        <div className="flex gap-2 w-full">
          <Button
            onClick={async () => {
              try {
                await discoverPointOfInterest(
                  pointOfInterest.id,
                  usersTeam?.id,
                  usersTeam ? undefined : currentUser?.id
                );
              } catch (error) {
                setButtonText('Wrong, dingus');
                setTimeout(() => {
                  setButtonText("I'm here!");
                }, 1000);
              }
            }}
            title={buttonText}
          />

          {ownedInventoryItems
            .filter((item) =>
              ItemsUsabledOnPointOfInterest.includes(item.inventoryItemId)
            )
            .map((item) => {
              const inventoryItem = inventoryItems.find(
                (i) => i.id === item.inventoryItemId
              );
              return (
                <img
                  src={inventoryItem?.imageUrl}
                  alt={inventoryItem?.name}
                  className="rounded-lg border-black border-2 h-12 w-12"
                  onClick={async () => {
                    await onConsumeItem(item.id, inventoryItem?.id ?? 0);

                    if (inventoryItem?.id === ItemType.GoldenTelescope) {
                      setDiscoveries([
                        ...discoveries,
                        {
                          id: '',
                          createdAt: new Date(),
                          updatedAt: new Date(),
                          teamId: usersTeam?.id,
                          userId: usersTeam ? undefined : currentUser?.id,
                          pointOfInterestId: pointOfInterest.id,
                        },
                      ]);
                    }
                  }}
                />
              );
            })}
        </div>
      )}
    </div>
  );
};
