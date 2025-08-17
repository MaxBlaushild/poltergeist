import React, { useState } from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { Logister as LogisterComponent } from '@poltergeist/components';
import { useAuth } from '@poltergeist/contexts';
import { XMarkIcon } from '@heroicons/react/24/outline';

export const Logister = ({
  isOpen,
  onClose,
}: {
  isOpen: boolean;
  onClose: () => void;
}) => {
  const [error, setError] = useState<string | undefined>(undefined);
  const {
    logister,
    getVerificationCode,
    isWaitingForVerificationCode,
    isRegister,
  } = useAuth();

  if (!isOpen) return null;

  return (
    <Modal size={ModalSize.SLIGHTLY_LARGER_HERO}>
      <XMarkIcon className="w-8 h-8 absolute top-2 right-2 cursor-pointer hover:text-gray-600 transition-colors" onClick={onClose} />
      <div className="flex flex-col gap-4 text-center w-full items-center">
      <h2 className="text-2xl font-bold">Sign in or sign up</h2>
      <LogisterComponent
        logister={async (one, two, three, isRegister) => {
          try {
            await logister(one, two, three);
            window.location.reload();
          } catch (e) {
            setError('Something went wrong. Please try again later');
          }
        }}
        getVerificationCode={getVerificationCode}
        isRegister={isRegister}
        isWaitingOnVerificationCode={isWaitingForVerificationCode}
        error={error}
      />
      </div>

    </Modal>
  );
};