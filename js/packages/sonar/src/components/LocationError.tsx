import React from 'react';
import { useLocation } from '@poltergeist/contexts';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { Button } from './shared/Button.tsx';
import { XMarkIcon } from '@heroicons/react/24/solid';

export const LocationError = () => {
  const { error, acknowledgeError } = useLocation();

  if (!error) {
    return null;
  }

  return (
    <Modal size={ModalSize.HERO}>
      <XMarkIcon className='w-6 h-6 absolute top-2 right-2' onClick={acknowledgeError} />
      <p className='w-full text-center p-2'>{error}</p>
    </Modal>
  );
};
