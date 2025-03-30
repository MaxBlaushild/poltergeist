import React, { useEffect, useState } from 'react';
import './MatchInProgress.css';
import './Match.css';
import {
  hasDiscoveredPointOfInterest,
  getHighestFirstCompletedChallenge,
  Match,
  PointOfInterest,
  PointOfInterestDiscovery,
  Team,
} from '@poltergeist/types';
import { Button } from './shared/Button.tsx';
import { PointOfInterestPanel } from './PointOfInterestPanel.tsx';
import { Drawer } from './Drawer.tsx';
import { Scoreboard } from './Scoreboard.tsx';
import { Inventory } from './Inventory.tsx';
import { Log } from './Log.tsx';
import NewItemModal from './NewItemModal.tsx';
import UsedItemModal from './UsedItemModal.tsx';
import { MapZoomButton } from './MapZoomButton.tsx';
import { Map } from './Map.tsx';
import { usePointOfInterestMarkers } from '../hooks/usePointOfInterestMarkers.tsx';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useInventory } from '@poltergeist/contexts';
import { useLocation } from '@poltergeist/contexts';
import { usePointOfInterestContext } from '../contexts/PointOfInterestContext.tsx';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';
import useSubmission from '../hooks/useSubmission.ts';
import { useSubmissionsContext } from '../contexts/SubmissionsContext.tsx';

export const MatchInProgress = () => {
  const { match, usersTeam } = useMatchContext();
  const { discoveries } = useDiscoveriesContext();
  const { currentUser } = useUserProfiles();
  const { pointsOfInterest } = usePointOfInterestContext();
  const { selectedPointOfInterest, setSelectedPointOfInterest } =
    usePointOfInterestMarkers({
      pointsOfInterest,
      discoveries,
      entityId: usersTeam?.id ?? '',
      needsDiscovery: true,
    });
  const [isPanelVisible, setIsPanelVisible] = useState(false);
  const [isLeaderboardVisible, setIsLeaderboardVisible] = useState(false);
  const [isInventoryVisible, setIsInventoryVisible] = useState(false);
  const [areMapOverlaysVisible, setAreMapOverlaysVisible] = useState(true);
  const { inventoryItems, setUsedItem, consumeItem } = useInventory();
  const { location } = useLocation();
  const { submissions } = useSubmissionsContext();
  

  useEffect(() => {
    if (selectedPointOfInterest) {
      setIsPanelVisible(true);
    } else {
      setIsPanelVisible(false);
    }
  }, [selectedPointOfInterest]);

  const closePanel = () => {
    if (isPanelVisible) {
      setIsPanelVisible(false);
      setSelectedPointOfInterest(null);
    }
  };

  useEffect(() => {
    if (isLeaderboardVisible || isInventoryVisible || isPanelVisible) {
      setAreMapOverlaysVisible(false);
    }

    if (!isLeaderboardVisible && !isInventoryVisible && !isPanelVisible) {
      setTimeout(() => {
        setAreMapOverlaysVisible(true);
      }, 300);
    }
  }, [isLeaderboardVisible, isInventoryVisible, isPanelVisible]);

  const handleMapClick = () => {
    setIsPanelVisible(false);
    setTimeout(() => {
      setSelectedPointOfInterest(null);
    }, 300);
  };

  const handleDrawerClick = (event: React.MouseEvent) => {
    event.stopPropagation();
  };

  const onClosePointOfInterestPanel = (immediate: boolean) => {
    if (immediate) {
      closePanel();
    } else {
      setTimeout(() => {
        closePanel();
      }, 2000);
    }
  };

  return (
    <Map>
      {areMapOverlaysVisible && <MapZoomButton />}
      {areMapOverlaysVisible && (
        <div className="absolute bottom-20 right-0 z-10 w-full p-2">
          <Log 
            match={match || undefined} 
            usersTeam={usersTeam} 
            pointsOfInterest={pointsOfInterest} 
            discoveries={discoveries} 
            needsDiscovery={true}
          />
        </div>
      )}
      <Drawer isVisible={isPanelVisible} onClose={closePanel} peekHeight={0}>
        {selectedPointOfInterest && (
          <PointOfInterestPanel
            pointOfInterest={selectedPointOfInterest}
            onClose={onClosePointOfInterestPanel}
            match={match || undefined}
            usersTeam={usersTeam}
            needsDiscovery
          />
        )}
      </Drawer>
      <Drawer
        isVisible={isLeaderboardVisible || isInventoryVisible}
        onClose={() => {
          setIsLeaderboardVisible(false);
          setIsInventoryVisible(false);
        }}
        peekHeight={isPanelVisible ? 0 : 80}
      >
        <div className="flex justify-between w-full gap-4">
          {!isLeaderboardVisible && !isInventoryVisible && (
            <Button
              onClick={() => setIsLeaderboardVisible(true)}
              title="Leaderboard"
            ></Button>
          )}
          {!isLeaderboardVisible && !isInventoryVisible && (
            <Button
              onClick={() => setIsInventoryVisible(true)}
              title="Inventory"
            ></Button>
          )}
        </div>
        {isLeaderboardVisible && <Scoreboard />}
        {isInventoryVisible && (
          <Inventory 
            onClose={() => setIsInventoryVisible(false)} 
            match={match || undefined} 
            usersTeam={usersTeam} 
          />
        )}
      </Drawer>
      {areMapOverlaysVisible && <NewItemModal />}
      {areMapOverlaysVisible && <UsedItemModal />}
    </Map>
  );
};
