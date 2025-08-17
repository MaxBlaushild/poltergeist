import React from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Button } from './shared/Button.tsx';

export const Welcome = ({ onClose, isOpen, onJourneyBegin }: { onClose: () => void, isOpen: boolean, onJourneyBegin: () => void }) => {
  if (!isOpen) return null;
  
  return <Modal size={ModalSize.SLIGHTLY_LARGER_HERO}>
    <div className="flex flex-col gap-6 text-center w-full items-center relative">
      <XMarkIcon className="w-8 h-8 absolute top-0 right-0 cursor-pointer hover:text-gray-600 transition-colors" onClick={onClose} />
      <div className="flex flex-col gap-4">
        <h2 className="text-2xl font-bold text-gray-900">Welcome to the Unclaimed Streets</h2>
        <p className="text-xl text-gray-700">The world awaits your legend. Explore. Seek quests. Stake your claim. Shape the map with your name.
        </p>
        <Button title="Begin Your Journey" onClick={onJourneyBegin} />
      </div>
    </div>
  </Modal>;
};