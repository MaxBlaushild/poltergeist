import React, { useEffect, useState } from 'react';
import './MatchInProgress.css';
import './Match.css';
import { Match, PointOfInterest, PointOfInterestDiscovery, Team } from '@poltergeist/types';
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
import { useUserLocator } from '../hooks/useUserLocator.tsx';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

export const MatchInProgress = () => {
  const { match } = useMatchContext();
  const { currentUser } = useUserProfiles();
  const usersTeam = match?.teams.find((team) =>
    team.users.some((user) => user.id === currentUser?.id)
  );
  return (
    <MultiplayerMap
      pointsOfInterest={match?.pointsOfInterest ?? []}
      discoveries={[]}
      usersTeam={usersTeam ?? null}
    />
  );
};

interface MultiplayerMapProps {
  pointsOfInterest: PointOfInterest[];
  discoveries: PointOfInterestDiscovery[];
  usersTeam: Team | null;
}

const MultiplayerMap = ({ pointsOfInterest, discoveries, usersTeam }: MultiplayerMapProps) => {
  const { selectedPointOfInterest, setSelectedPointOfInterest } = usePointOfInterestMarkers({
    pointsOfInterest,
    discoveries,
    entityId: usersTeam?.id ?? '',
  });
  const [isPanelVisible, setIsPanelVisible] = useState(false);
  const [isLeaderboardVisible, setIsLeaderboardVisible] = useState(false);
  const [isInventoryVisible, setIsInventoryVisible] = useState(false);
  const [areMapOverlaysVisible, setAreMapOverlaysVisible] = useState(true);

  useUserLocator();

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

  return (
    <Map>
      <MapZoomButton />
      {areMapOverlaysVisible && <MapZoomButton />}
      {areMapOverlaysVisible && (
        <div className="absolute bottom-20 right-0 z-10 w-full p-2">
          <Log />
        </div>
      )}
      <Drawer isVisible={isPanelVisible} onClose={closePanel} peekHeight={0}>
        {selectedPointOfInterest && (
          <PointOfInterestPanel
            pointOfInterest={selectedPointOfInterest}
            onClose={(immediate) => {
              if (immediate) {
                closePanel();
              } else {
                setTimeout(() => {
                  closePanel();
                }, 2000);
              }
            }}
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
          <Inventory onClose={() => setIsInventoryVisible(false)} />
        )}
      </Drawer>
      {areMapOverlaysVisible && <NewItemModal />}
      {areMapOverlaysVisible && <UsedItemModal />}
    </Map>
  );
};
