import React, { useEffect, useState } from 'react';
import { Map } from './Map.tsx';
import { useAuth } from '@poltergeist/contexts';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { usePointOfInterestMarkers } from '../hooks/usePointOfInterestMarkers.tsx';
import { useZoneBoundaries } from '../hooks/useZoneBoundaries.ts';
import { PointOfInterest, PointOfInterestDiscovery, Zone } from '@poltergeist/types';
import { MapZoomButton } from './MapZoomButton.tsx';
import { TagFilter } from './TagFilter.tsx';
import { Drawer } from './Drawer.tsx';
import { Button } from './shared/Button.tsx';
import { PointOfInterestPanel } from './PointOfInterestPanel/PointOfInterestPanel.tsx';
import { Inventory } from './Inventory.tsx';
import { QuestLog } from './QuestLog.tsx';
import NewItemModal from './NewItemModal.tsx';
import UsedItemModal from './UsedItemModal.tsx';
import { usePointOfInterestContext } from '../contexts/PointOfInterestContext.tsx';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { Log } from './Log.tsx';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { usePointsOfInterest } from '@poltergeist/hooks';
import { ActivityQuestionnaire } from './ActivityQuestionnaire.tsx';
import { TrackedQuests } from './TrackedQuests.tsx';
import { CelebrationModalManager } from './CelebrationModalManager.tsx';
import { useCompletedTaskContext } from '../contexts/CompletedTaskContext.tsx';
import { ZoneWidget } from './ZoneWidget.tsx';

const MemoizedMap = React.memo(Map);

const MapOverlays = React.memo(({ areMapOverlaysVisible, discoveries, totalPointsOfInterest, openPointOfInterestPanel }: {
  areMapOverlaysVisible: boolean;
  discoveries: PointOfInterestDiscovery[];
  totalPointsOfInterest: PointOfInterest[];
  openPointOfInterestPanel: (pointOfInterest: PointOfInterest) => void;
}) => {
  if (!areMapOverlaysVisible) return null;
  
  return (
    <>
      <MapZoomButton />
      <TagFilter />
      <ActivityQuestionnaire />
      <div className="absolute top-56 right-4 z-10 mt-2">
        <TrackedQuests openPointOfInterestPanel={openPointOfInterestPanel} />
      </div>
      <div className="absolute bottom-20 right-0 z-10 w-full p-2">
        <Log 
          pointsOfInterest={totalPointsOfInterest || []} 
          discoveries={discoveries}
          needsDiscovery={false}
        />
      </div>

    </>
  );
});

const DrawerControls = React.memo(({ 
  isInventoryOpen, 
  isQuestLogOpen, 
  setIsInventoryOpen, 
  setIsQuestLogOpen 
}: {
  isInventoryOpen: boolean;
  isQuestLogOpen: boolean;
  setIsInventoryOpen: (value: boolean) => void;
  setIsQuestLogOpen: (value: boolean) => void;
}) => {
  if (isInventoryOpen || isQuestLogOpen) return null;
  
  return (
    <div className="flex justify-between w-full gap-4">
      <Button
        onClick={() => setIsInventoryOpen(true)}
        title="Inventory"
      />
      <Button
        onClick={() => setIsQuestLogOpen(true)}
        title="Quest Log"
      />
    </div>
  );
});

export const SinglePlayer = () => {
  const { pointsOfInterest, trackedPointOfInterestIds, quests } = useQuestLogContext();
  const { discoveries } = useDiscoveriesContext();
  const { currentUser } = useUserProfiles();
  const { completedTask } = useCompletedTaskContext();

  const { selectedPointOfInterest, setSelectedPointOfInterest } = usePointOfInterestMarkers({
    pointsOfInterest,
    discoveries,
    entityId: currentUser?.id ?? '',
    needsDiscovery: false,
    trackedPointOfInterestIds,
  });

  useZoneBoundaries();

  const openPointOfInterestPanel = (pointOfInterest: PointOfInterest) => {
    if (pointOfInterest) {
      setTimeout(() => {
        setSelectedPointOfInterest(pointOfInterest);
      }, 1000);
    }
  };

  const [levelsGained, setLevelsGained] = useState(0);
  const [zone, setZone] = useState<Zone | undefined>(undefined);
  const [isPanelVisible, setIsPanelVisible] = useState(false);
  const [isInventoryOpen, setIsInventoryOpen] = useState(false);
  const [isQuestLogOpen, setIsQuestLogOpen] = useState(false);
  const [areMapOverlaysVisible, setAreMapOverlaysVisible] = useState(true);
  const { pointsOfInterest: totalPointsOfInterest } = usePointsOfInterest();
  
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
    if (isInventoryOpen || isQuestLogOpen || isPanelVisible || completedTask) {
      setAreMapOverlaysVisible(false);
    }

    if (!isInventoryOpen && !isQuestLogOpen && !isPanelVisible && !completedTask) {
      setTimeout(() => {
        setAreMapOverlaysVisible(true);
      }, 300);
    }
  }, [isInventoryOpen, isQuestLogOpen, isPanelVisible, completedTask]);

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
    <MemoizedMap>
      <ZoneWidget onWidgetOpen={() => setAreMapOverlaysVisible(false)} onWidgetClose={() => setAreMapOverlaysVisible(true)} />
      <MapOverlays 
        areMapOverlaysVisible={areMapOverlaysVisible}
        discoveries={discoveries}
        totalPointsOfInterest={totalPointsOfInterest || []}
        openPointOfInterestPanel={openPointOfInterestPanel}
      />
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
        <DrawerControls 
          isInventoryOpen={isInventoryOpen}
          isQuestLogOpen={isQuestLogOpen}
          setIsInventoryOpen={setIsInventoryOpen}
          setIsQuestLogOpen={setIsQuestLogOpen}
        />
        {isInventoryOpen && <Inventory onClose={() => setIsInventoryOpen(false)} />}
        {isQuestLogOpen && <QuestLog onClose={(pointOfInterest: PointOfInterest | null | undefined) => {
          setIsQuestLogOpen(false);
          if (pointOfInterest) {
            openPointOfInterestPanel(pointOfInterest);
          }
        }} />}
      </Drawer>
      <CelebrationModalManager />
      <NewItemModal />
      <UsedItemModal />
    </MemoizedMap>
  );
};
