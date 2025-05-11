import React from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { Zone } from '@poltergeist/types';
import { useUserZoneReputation } from '@poltergeist/hooks';
import { useZoneContext } from '@poltergeist/contexts';
import { useCompletedTaskContext } from '../contexts/CompletedTaskContext.tsx';

export const ReputationUpModal = () => {
  const { zones } = useZoneContext();
  const { reputationUp, setReputationUp, completedTask, zoneId, levelUp } = useCompletedTaskContext();
  const { userZoneReputation } = useUserZoneReputation(zoneId ?? undefined);
  const zone = zones.find((zone) => zone.id === zoneId);
  console.log('zone', zone);
  console.log('userZoneReputation', userZoneReputation);
  console.log('reputationUp', reputationUp);

  if (!reputationUp || !zone || !userZoneReputation || levelUp) return null;
  console.log('reputation up modal');

  return (
    <Modal onClose={() => {
      setReputationUp(false);
    }} size={ModalSize.FREE}>
      <div className="flex flex-col items-center gap-4 p-6">
        <h1 className="text-2xl font-bold text-amber-500">Reputation Up!</h1>
        <div className="w-full bg-white rounded-lg shadow-md p-4">
          <div className="flex items-center justify-center gap-4">
            <div className="text-center">
              <p className="text-lg font-semibold text-gray-900">
                Congratulations!
              </p>
              <p className="text-sm text-gray-600">
                You have become {userZoneReputation?.name} in {zone.name}.
              </p>
            </div>
          </div>
        </div>
      </div>
    </Modal>
  );
};
