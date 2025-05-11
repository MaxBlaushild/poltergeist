import React from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useCompletedTaskContext } from '../contexts/CompletedTaskContext.tsx';

export const LevelUpModal = () => {
  const { levelUp, setLevelUp, completedTask, reputationUp, setReputationUp } = useCompletedTaskContext();

  if (!levelUp) return null;

  return (
    <Modal onClose={() => {
      if (reputationUp) {
        setReputationUp(false);
        setTimeout(() => {
          setReputationUp(true);
        }, 300);
      }
      setLevelUp(false);
    }} size={ModalSize.FREE}>
      <div className="flex flex-col items-center gap-4 p-6">
        <h1 className="text-2xl font-bold text-amber-500">Level Up!</h1>
        <div className="w-full bg-white rounded-lg shadow-md p-4">
          <div className="flex items-center justify-center gap-4">
            <div className="w-16 h-16 rounded-full bg-blue-100 flex items-center justify-center">
              <span className="text-3xl font-bold text-blue-600">+1</span>
            </div>
            <div className="text-center">
              <p className="text-lg font-semibold text-gray-900">
                Congratulations!
              </p>
              <p className="text-sm text-gray-600">
                You gained a level!
              </p>
            </div>
          </div>
        </div>
      </div>
    </Modal>
  );
};

