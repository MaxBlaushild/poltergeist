import React from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useCompletedTaskContext } from '../contexts/CompletedTaskContext.tsx';

export const ReputationUpModal = () => {
  const { reputationUp, setReputationUp, zoneName, newReputationLevel, levelUp } = useCompletedTaskContext();

  if (!reputationUp || !zoneName || newReputationLevel === null || levelUp) return null;

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
                You reached level {newReputationLevel} in {zoneName}!
              </p>
            </div>
          </div>
        </div>
      </div>
    </Modal>
  );
};
