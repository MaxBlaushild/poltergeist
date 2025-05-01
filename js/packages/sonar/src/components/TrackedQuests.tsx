import React, { useState, useRef, useEffect } from 'react';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import { QuestNodeComponent } from './Quest.tsx';
import { PointOfInterest, QuestNode } from '@poltergeist/types';
import { useMap } from '@poltergeist/contexts';

export const TrackedQuests = ({ openPointOfInterestPanel }: { openPointOfInterestPanel: (pointOfInterest: PointOfInterest) => void }) => {
  const { trackedQuestIds, quests } = useQuestLogContext();
  const [isExpanded, setIsExpanded] = useState(false);
  const [showContent, setShowContent] = useState(false);
  const { flyToLocation } = useMap();
  const trackedQuests = quests.filter((quest) =>
    trackedQuestIds.includes(quest.id)
  );

  useEffect(() => {
    let timeout;
    if (isExpanded) {
      timeout = setTimeout(() => {
        setShowContent(true);
      }, 300); // Match the width transition duration
    } else {
      setShowContent(false);
    }
    return () => clearTimeout(timeout);
  }, [isExpanded]);

  if (trackedQuests.length === 0) {
    return null;
  }

  const onPointOfInterestClick = (
    e: React.MouseEvent<HTMLDivElement>,
    node: QuestNode
  ) => {
    e.stopPropagation();
    setIsExpanded(false);
    flyToLocation(
      parseFloat(node.pointOfInterest.lat),
      parseFloat(node.pointOfInterest.lng)
    );
    openPointOfInterestPanel(node.pointOfInterest);
  };

  return (
    <div className={`flex flex-col ${showContent ? 'gap-2' : 'gap-0'} ${isExpanded ? 'w-72' : 'w-20'} transition-[width] duration-300 rounded-lg bg-black/50 p-2`}>
      <div className="flex justify-between items-center cursor-pointer" onClick={() => setIsExpanded(!isExpanded)}>
        <h1 className="text-sm font-bold text-white">Quests</h1>
        {isExpanded ? (
          <ChevronUpIcon className="w-4 h-4 text-white" />
        ) : (
          <ChevronDownIcon className="w-4 h-4 text-white" />
        )}
      </div>
      <div className={`flex flex-col gap-2 overflow-hidden transition-[max-height] duration-300 ${showContent ? 'max-h-96' : 'max-h-0'}`}> 
        {trackedQuests.map((quest) => (
          <div key={quest.id} className="p-2 rounded bg-black/10">
            <h2 className="font-semibold text-xs text-white mb-2">{quest.name}</h2>
            <QuestNodeComponent
              node={quest.rootNode}
              onPointOfInterestClick={onPointOfInterestClick}
              hasDiscoveredNode={false}
              darkMode
            />
          </div>
        ))}
      </div>
    </div>
  );
};
