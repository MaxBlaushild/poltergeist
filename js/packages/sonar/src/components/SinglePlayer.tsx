import React, { useEffect, useState } from 'react';
import { Map } from './Map.tsx';
import { usePointsOfInterest } from '@poltergeist/hooks';
import { useAuth } from '@poltergeist/contexts';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useUserLocator } from '../hooks/useUserLocator.tsx';
import { usePointOfInterestMarkers } from '../hooks/usePointOfInterestMarkers.tsx';
import { PointOfInterest, PointOfInterestDiscovery } from '@poltergeist/types';
import { MapZoomButton } from './MapZoomButton.tsx';
import { Drawer } from './Drawer.tsx';
import { Button } from './shared/Button.tsx';
import { PointOfInterestPanel } from './PointOfInterestPanel.tsx';
import { Inventory } from './Inventory.tsx';
import { QuestLog } from './QuestLog.tsx';
import NewItemModal from './NewItemModal.tsx';
import UsedItemModal from './UsedItemModal.tsx';

export const SinglePlayer = () => {
  const { pointsOfInterest, loading, error } = usePointsOfInterest();
  const { currentUser } = useUserProfiles();

  return (
    <div>
      <SinglePlayerMap
        pointsOfInterest={pointsOfInterest || []}
        discoveries={[]}
        entityId={currentUser?.id ?? ''}
      />
    </div>
  );
};

interface SinglePlayerMapProps {
  pointsOfInterest: PointOfInterest[];
  discoveries: PointOfInterestDiscovery[];
  entityId: string;
}

const SinglePlayerMap = ({ pointsOfInterest, discoveries, entityId }: SinglePlayerMapProps) => {
  const { selectedPointOfInterest, setSelectedPointOfInterest } = usePointOfInterestMarkers({
    pointsOfInterest,
    discoveries,
    entityId,
  });
  const [isPanelVisible, setIsPanelVisible] = useState(false);
  const [isInventoryOpen, setIsInventoryOpen] = useState(false);
  const [isQuestLogOpen, setIsQuestLogOpen] = useState(false);
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
    if (isInventoryOpen || isQuestLogOpen || isPanelVisible) {
      setAreMapOverlaysVisible(false);
    }

    if (!isInventoryOpen && !isQuestLogOpen && !isPanelVisible) {
      setTimeout(() => {
        setAreMapOverlaysVisible(true);
      }, 300);
    }
  }, [isInventoryOpen, isQuestLogOpen, isPanelVisible]);

  const handleMapClick = () => {
    setIsPanelVisible(false);
    setTimeout(() => {
      setSelectedPointOfInterest(null);
    }, 300);
  };

  const handleDrawerClick = (event: React.MouseEvent) => {
    event.stopPropagation();
  };



  return <Map>
    <MapZoomButton />
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
        isVisible={isInventoryOpen || isQuestLogOpen}
        onClose={() => {
          setIsInventoryOpen(false);
          setIsQuestLogOpen(false);
        }}
        peekHeight={isPanelVisible ? 0 : 80}
      >
        <div className="flex justify-between w-full gap-4">
          {!isInventoryOpen && !isQuestLogOpen && (
            <Button
              onClick={() => setIsInventoryOpen(true)}
              title="Inventory"
            ></Button>
          )}
          {!isInventoryOpen && !isQuestLogOpen && (
            <Button
              onClick={() => setIsQuestLogOpen(true)}
              title="Quest Log"
            ></Button>
          )}
        </div>
        {isInventoryOpen && <Inventory onClose={() => setIsInventoryOpen(false)} />}
        {isQuestLogOpen && <QuestLog onClose={() => setIsQuestLogOpen(false)} />}
      </Drawer>
      {areMapOverlaysVisible && <NewItemModal />}
      {areMapOverlaysVisible && <UsedItemModal />}
  </Map>;
};
