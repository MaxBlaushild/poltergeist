import './Modal.css';
import React from 'react';

export enum ModalSize {
  FULLSCREEN = 'FULLSCREEN',
  HERO = 'HERO',
  FREE = 'FREE',
}

type ModalProps = {
  children: React.ReactNode;
  size?: ModalSize;
};

export const Modal = ({ children, size = ModalSize.HERO }: ModalProps) => {
  const modalClasses = ['Modal__modal'];

  if (size === ModalSize.FULLSCREEN) {
    modalClasses.push('Modal__fullScreen');
  } else if (size === ModalSize.HERO) {
    modalClasses.push('Modal__hero');
  }

  return <div className={modalClasses.join(' ')}>{children}</div>;
};
