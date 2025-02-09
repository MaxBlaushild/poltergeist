import React, { useEffect, useState } from 'react';
import { useInventory } from '@poltergeist/contexts';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import {
  InventoryItem,
  ItemsRequiringTeamId,
  ItemsUsabledInMenu,
  TeamInventoryItem,
} from '@poltergeist/types';
import Divider from './shared/Divider.tsx';
import { ArrowLeftIcon } from '@heroicons/react/20/solid';
import { Button, ButtonColor, ButtonSize } from './shared/Button.tsx';
import { generateColorFromTeamName } from '../utils/generateColor.ts';

export const Inventory = ({ onClose }: { onClose: () => void }) => {
  const { inventoryItems, inventoryItemsAreLoading, consumeItem, setUsedItem } =
    useInventory();
  const { usersTeam, match } = useMatchContext();
  const [selectedItem, setSelectedItem] = useState<
    TeamInventoryItem | undefined
  >(undefined);
  const [itemBeingUsed, setItemBeingUsed] = useState<
    TeamInventoryItem | undefined
  >(undefined);
  const selectedInventoryItem = inventoryItems?.find(
    (i) => i.id === selectedItem?.inventoryItemId
  );
  return (
    <div className="flex flex-col gap-4 w-full">
      {selectedItem || itemBeingUsed ? (
        <ArrowLeftIcon
          className="w-8 h-8 absolute left-4"
          onClick={() => {
            setSelectedItem(undefined);
            setItemBeingUsed(undefined);
          }}
        />
      ) : null}
      <h2 className="text-2xl font-bold">
        {selectedInventoryItem ? selectedInventoryItem?.name : 'Inventory'}
      </h2>
      {!selectedItem && !itemBeingUsed ? (
        <div className="grid grid-cols-3 gap-2 mt-4">
          {Array.from({ length: 12 }).map((_, index) => {
            const item = usersTeam?.teamInventoryItems?.[index];
            const inventoryItem = item
              ? inventoryItems.find((i) => i.id === item.inventoryItemId)
              : null;

            return (
              <div
                key={index}
                className="aspect-square rounded-lg border border-black flex justify-center items-center relative overflow-hidden"
              >
                {inventoryItem ? (
                  <>
                    <img
                      onClick={() => setSelectedItem(item)}
                      src={inventoryItem.imageUrl}
                      alt={inventoryItem.name}
                      className="max-h-full max-w-full"
                    />
                    <span className="absolute bottom-1 right-1 text-white bg-black px-1 rounded">
                      {item?.quantity}
                    </span>
                  </>
                ) : null}
              </div>
            );
          })}
        </div>
      ) : selectedItem && !itemBeingUsed ? (
        <div className="flex flex-col gap-4">
          <img
            src={selectedInventoryItem?.imageUrl}
            alt={selectedInventoryItem?.name}
            className="max-h-full max-w-full rounded-lg border border-black flex"
          />
          <div className="flex flex-col gap-2">
            <p className="text-lg font-bold">
              {selectedInventoryItem?.flavorText}
            </p>
            <p className="text-lg">{selectedInventoryItem?.effectText}</p>
            {ItemsUsabledInMenu.includes(selectedInventoryItem!.id) ? (
              <Button
                title="Use"
                disabled={
                  !selectedItem?.quantity || selectedItem?.quantity === 0
                }
                onClick={() => {
                  if (
                    ItemsRequiringTeamId.includes(selectedItem.inventoryItemId)
                  ) {
                    setItemBeingUsed(selectedItem);
                  } else {
                    consumeItem(selectedItem!.id);
                    setUsedItem(selectedInventoryItem!);
                    onClose();
                  }
                }}
              />
            ) : null}
          </div>
        </div>
      ) : (
        <div>
          <table className="w-full mt-4">
            <thead>
              <tr>
                <th className="text-left">Team</th>
                <th className="text-center">Color</th>
                <th className="text-center">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <br />
              </tr>
              {match?.teams
                .filter((t) => t.id !== usersTeam?.id)
                .map((team) => (
                  <tr key={team.id}>
                    <td className="text-left text-lg font-bold">{team.name}</td>
                    <td className="text-center">
                      <div
                        style={{
                          width: '32px',
                          height: '32px',
                          backgroundColor: generateColorFromTeamName(team.name),
                          borderRadius: '50%',
                          margin: 'auto',
                        }}
                      />
                    </td>
                    <td className="text-center text-xl font-bold p-2">
                      <Button
                        buttonSize={ButtonSize.SMALL}
                        title="Choose"
                        onClick={() => {
                          consumeItem(selectedItem!.id, {
                            targetTeamId: team.id,
                          });
                          setUsedItem(selectedInventoryItem!);
                          onClose();
                        }}
                      />
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
