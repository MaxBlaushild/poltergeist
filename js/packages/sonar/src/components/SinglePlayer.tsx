import React, { useEffect, useState } from 'react';
import { Map } from './Map.tsx';
import { useAuth } from '@poltergeist/contexts';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { usePointOfInterestMarkers } from '../hooks/usePointOfInterestMarkers.tsx';
import { useZoneBoundaries } from '../hooks/useZoneBoundaries.ts';
import { useCharacterMarkers } from '../hooks/useCharacterMarkers.tsx';
import { PointOfInterest, PointOfInterestDiscovery, Zone, Character, CharacterAction } from '@poltergeist/types';
import { MapZoomButton } from './MapZoomButton.tsx';
import { TagFilter } from './TagFilter.tsx';
import { Drawer } from './Drawer.tsx';
import { Button } from './shared/Button.tsx';
import { PointOfInterestPanel } from './PointOfInterestPanel/PointOfInterestPanel.tsx';
import { CharacterPanel } from './CharacterPanel.tsx';
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
import { RPGDialogue } from './RPGDialogue.tsx';
import { Shop } from './Shop.tsx';

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
  
  const [isCharacterPanelVisible, setIsCharacterPanelVisible] = useState(false);
  
  const { selectedCharacter, setSelectedCharacter } = useCharacterMarkers((character: Character) => {
    setIsCharacterPanelVisible(true);
  });

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
  const [dialogueCharacter, setDialogueCharacter] = useState<Character | null>(null);
  const [dialogueAction, setDialogueAction] = useState<CharacterAction | null>(null);
  const [isDialogueVisible, setIsDialogueVisible] = useState(false);
  const [shopCharacter, setShopCharacter] = useState<Character | null>(null);
  const [shopAction, setShopAction] = useState<CharacterAction | null>(null);
  const [isShopVisible, setIsShopVisible] = useState(false);
  
  useEffect(() => {
    if (selectedPointOfInterest) {
      setIsPanelVisible(true);
    } else {
      setIsPanelVisible(false);
    }
  }, [selectedPointOfInterest]);

  useEffect(() => {
    if (selectedCharacter) {
      setIsCharacterPanelVisible(true);
    } else {
      setIsCharacterPanelVisible(false);
    }
  }, [selectedCharacter]);

  const closePanel = () => {
    if (isPanelVisible) {
      setIsPanelVisible(false);
      setSelectedPointOfInterest(null);
    }
  };

  const closeCharacterPanel = () => {
    if (isCharacterPanelVisible) {
      setIsCharacterPanelVisible(false);
      setSelectedCharacter(null);
    }
  };

  const handleStartDialogue = (character: Character, action: CharacterAction) => {
    setDialogueCharacter(character);
    setDialogueAction(action);
    setIsDialogueVisible(true);
  };

  const handleCloseDialogue = () => {
    setIsDialogueVisible(false);
    setDialogueCharacter(null);
    setDialogueAction(null);
  };

  const handleStartShop = (character: Character, action: CharacterAction) => {
    setShopCharacter(character);
    setShopAction(action);
    setIsShopVisible(true);
  };

  const handleCloseShop = () => {
    setIsShopVisible(false);
    setShopCharacter(null);
    setShopAction(null);
  };

  useEffect(() => {
    if (isInventoryOpen || isQuestLogOpen || isPanelVisible || isCharacterPanelVisible || completedTask || isDialogueVisible || isShopVisible) {
      setAreMapOverlaysVisible(false);
    }

    if (!isInventoryOpen && !isQuestLogOpen && !isPanelVisible && !isCharacterPanelVisible && !completedTask && !isDialogueVisible && !isShopVisible) {
      setTimeout(() => {
        setAreMapOverlaysVisible(true);
      }, 300);
    }
  }, [isInventoryOpen, isQuestLogOpen, isPanelVisible, isCharacterPanelVisible, completedTask, isDialogueVisible, isShopVisible]);

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
      <Drawer isVisible={isCharacterPanelVisible} onClose={closeCharacterPanel} peekHeight={0}>
        {selectedCharacter && (
          <CharacterPanel
            character={selectedCharacter}
            onClose={closeCharacterPanel}
            onStartDialogue={handleStartDialogue}
            onStartShop={handleStartShop}
          />
        )}
      </Drawer>
      {isDialogueVisible && dialogueCharacter && dialogueAction && (
        <RPGDialogue
          character={dialogueCharacter}
          action={dialogueAction}
          onClose={handleCloseDialogue}
        />
      )}
      {isShopVisible && shopCharacter && shopAction && (
        <Shop
          character={shopCharacter}
          action={shopAction}
          onClose={handleCloseShop}
        />
      )}
      <Drawer
        isVisible={isInventoryOpen || isQuestLogOpen}
        onClose={() => {
          setIsInventoryOpen(false);
          setIsQuestLogOpen(false);
        }}
        peekHeight={isPanelVisible || isCharacterPanelVisible ? 0 : 80}
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
