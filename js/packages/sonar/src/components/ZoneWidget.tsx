import React, { useEffect, useRef, useState } from 'react';
import { useZoneContext } from '@poltergeist/contexts';
import { useLocation } from '@poltergeist/contexts';
import { isXMetersAway } from '../utils/calculateDistance.ts';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { useUserZoneReputation } from '@poltergeist/hooks';
interface ZoneWidgetProps {
  onWidgetOpen: () => void;
  onWidgetClose: () => void;
}

export const ZoneWidget = ({ onWidgetOpen, onWidgetClose }: ZoneWidgetProps) => {
  const { zones, selectedZone, setSelectedZone, findZoneAtCoordinate } = useZoneContext();
  const { quests, pendingTasks, completedTasks } = useQuestLogContext();
  const { userZoneReputation } = useUserZoneReputation(selectedZone?.id);

  const [isOpen, setIsOpen] = useState(false);
  const [showContent, setShowContent] = useState(false);

  useEffect(() => {
    let timeout;
    if (isOpen) {
      timeout = setTimeout(() => {
        setShowContent(true);
      }, 300); // Match transition duration
    } else {
      setShowContent(false);
    }
    return () => clearTimeout(timeout);
  }, [isOpen]);

  return (
    <div className="absolute top-20 left-1/2 -translate-x-1/2 z-10">
      <div 
        className={`
          bg-white rounded-lg p-2 border-2 border-black opacity-80 
          flex flex-col cursor-pointer
          transition-all duration-300
          ${isOpen ? 'w-64' : 'w-36 h-10'}
        `}
        onClick={() => {
          setIsOpen(!isOpen);
          if (isOpen) {
            onWidgetClose();
          } else {
            onWidgetOpen();
          }
        }}
      >
        <div className="flex items-center gap-2 justify-between">
          <p className="text-sm">{selectedZone?.name ?? 'Hinterlands'}</p>
          <svg 
            xmlns="http://www.w3.org/2000/svg" 
            className={`h-4 w-4 transition-transform duration-300`}
            fill="none" 
            viewBox="0 0 24 24" 
            stroke="currentColor"
          >
            <path 
              strokeLinecap="round" 
              strokeLinejoin="round" 
              strokeWidth={2} 
              d={isOpen ? "M19 14l-7-7-7 7" : "M5 10l7 7 7-7"}
            />
          </svg>
        </div>
        <div className={`overflow-y-auto transition-all duration-300 ${showContent ? 'max-h-24 mt-2' : 'max-h-0'}`}>
          <p className="text-sm">{selectedZone?.description ?? 'This will be a brief description of the zone written in fantasy terms.'}</p>
        </div>
        {showContent && (
          <div className="mt-2 text-sm space-y-1">
            <div className="flex justify-between items-center">
              <span>Active Tasks</span>
              <span>{Object.keys(pendingTasks).length || 0}</span>
            </div>
            <div className="flex justify-between items-center">
              <span>Completed Tasks</span>
              <span>{Object.keys(completedTasks).length || 0}</span>
            </div>
          </div>
        )}
        {showContent && (
          <div className="mt-2 mb-2">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">Reputation: {userZoneReputation?.name}</span>
              <span className="text-xs">{userZoneReputation?.reputationOnLevel} / {userZoneReputation?.reputationToNextLevel}</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className="bg-green-600 h-2 rounded-full transition-all duration-300" 
                style={{ width: `${userZoneReputation ? (userZoneReputation.reputationOnLevel / userZoneReputation.reputationToNextLevel) * 100 : 0}%` }} 
              />
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
