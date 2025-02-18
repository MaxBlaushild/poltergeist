import React from 'react';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { useInventory } from '@poltergeist/contexts';
import { hasDiscoveredPointOfInterest } from '@poltergeist/types';
import { generateColorFromTeamName } from '../utils/generateColor.ts';
import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/20/solid';
import { useLogContext } from '../contexts/LogContext.tsx';
import { usePointOfInterestContext } from '../contexts/PointOfInterestContext.tsx';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

const Team = 'Team';
const PointOfInterest = 'PointOfInterest';
const InventoryItem = 'InventoryItem';

export const Log = () => {
  const { match, usersTeam } = useMatchContext();
  const { auditItems, fetchAuditItems } = useLogContext();
  const { inventoryItems } = useInventory();
  const [isExpanded, setIsExpanded] = React.useState(false);
  const { pointsOfInterest } = usePointOfInterestContext();
  const { discoveries } = useDiscoveriesContext();
  const { currentUser } = useUserProfiles();

  React.useEffect(() => {
    fetchAuditItems();

    const intervalId = setInterval(() => {
      fetchAuditItems();
    }, 5000);

    return () => clearInterval(intervalId);
  }, [fetchAuditItems]);

  const teamsById = match?.teams?.reduce(
    (acc, team) => ({ ...acc, [team.id]: team }),
    {}
  );
  const pointsOfInterestById = pointsOfInterest.reduce(
    (acc, pointOfInterest) => ({
      ...acc,
      [pointOfInterest.id]: pointOfInterest,
    }),
    {}
  );
  const inventoryItemsById = inventoryItems.reduce(
    (acc, inventoryItem) => ({ ...acc, [inventoryItem.id]: inventoryItem }),
    {}
  );

  const reverseEngineerMessage = (message) => {
    const teamPattern = /\{Team\|([^\}]+)\}/g; // Use global flag to match all occurrences
    const pointOfInterestPattern = /\{PointOfInterest\|([^\}]+)\}/;
    const inventoryItemPattern = /\{InventoryItem\|(\d+)\}/;

    let teamMatches;
    while ((teamMatches = teamPattern.exec(message)) !== null) {
      const teamId = teamMatches[1];
      const teamName = teamsById?.[teamId]?.name || 'Unknown Team';
      const color = generateColorFromTeamName(teamName); // Corrected to use teamName instead of hardcoded string
      message = message.replace(
        `{Team|${teamId}}`,
        `<span style="background-color: white; color: ${color}; border: 2px solid ${color}; padding: 0px 4px; border-radius: 4px; line-height: 1.8;">${teamName}</span>`
      );
    }

    const pointOfInterestMatches = message.match(pointOfInterestPattern);
    if (pointOfInterestMatches) {
      const poiId = pointOfInterestMatches[1];
      const pointOfInterest = pointsOfInterestById?.[poiId];
      const poiName = pointOfInterest?.name || 'Unknown Point of Interest';
      const isDiscovered = hasDiscoveredPointOfInterest(
        pointOfInterest.id,
        usersTeam?.id ?? currentUser?.id ?? '',
        discoveries ?? []
      );
      message = message.replace(
        pointOfInterestPattern,
        `<span style="background-color: #E5E5E5; color: black; border: 2px solid black; padding: 0px 4px; border-radius: 4px; line-height: 1.8; white-space: nowrap;">${isDiscovered ? poiName : `???`}</span>`)
    }

    const inventoryItemMatches = message.match(inventoryItemPattern);
    if (inventoryItemMatches) {
      const inventoryItemId = inventoryItemMatches[1];
      const inventoryItemName =
        inventoryItemsById[inventoryItemId]?.name || 'Unknown Inventory Item';
      message = message.replace(inventoryItemPattern, inventoryItemName);
    }

    return message;
  };

  return (
    auditItems?.length > 0 && (
    <div className="flex flex-col gap-2 w-full rounded-lg bg-black/50 p-3">
      <div className="flex gap-2 justify-around w-full">
        <div className="flex justify-center items-center">
          <h2 className="text-lg font-bold" style={{ color: '#E5E5E5', cursor: 'pointer' }} onClick={() => setIsExpanded(!isExpanded)}>Log</h2>
          {isExpanded ? (
            <ChevronUpIcon
              className="w-4 h-4"
              style={{ color: '#E5E5E5', cursor: 'pointer' }}
              onClick={() => setIsExpanded(false)}
            />
          ) : (
            <ChevronDownIcon
              className="w-4 h-4"
              style={{ color: '#E5E5E5', cursor: 'pointer' }} 
              onClick={() => setIsExpanded(true)}
            />
          )}
        </div>
      </div>
      {isExpanded && (
          <div style={{ maxHeight: '180px', overflowY: 'auto' }} className='w-full'>
          {auditItems.map((item) => (
            <div key={item.id} className="text-sm p-2 font-white">
              <div className='text-md' style={{ color: '#E5E5E5' }}
                dangerouslySetInnerHTML={{
                  __html: reverseEngineerMessage(item.message),
                }}
              />
            </div>
          ))}
        </div>
      )}
    </div>
    )
  );
};
