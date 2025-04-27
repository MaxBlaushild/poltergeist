import './Modal.css';
import React from 'react';

export enum ModalSize {
  FULLSCREEN = 'FULLSCREEN',
  HERO = 'HERO',
  FREE = 'FREE',
  TOAST = 'TOAST',
  FORM = 'FORM',
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
  } else if (size === ModalSize.TOAST) {
    modalClasses.push('Modal__toast');
  } else if (size === ModalSize.FORM) {
    modalClasses.push('Modal__form');
  }

  return <div className={modalClasses.join(' ')}>{children}</div>;
};
