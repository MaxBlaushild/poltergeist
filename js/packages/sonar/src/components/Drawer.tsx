import { XMarkIcon } from '@heroicons/react/20/solid';
import React from 'react';
import './Drawer.css';

type DrawerProps = {
  isVisible: boolean;
  onClose: () => void;
  children: React.ReactNode;
  peekHeight: number; // New prop to control how much the drawer peeks out when not visible
};

export const Drawer = ({ isVisible, onClose, children, peekHeight }: DrawerProps) => {
  return (
    <div
      className="Drawer"
      style={{
        position: 'fixed',
        bottom: 0,
        left: 0,
        width: '100vw',
        height: 'calc(100% - 90px)',
        transition: 'transform 0.3s ease-in-out',
        transform: isVisible ? 'translateY(0)' : `translateY(calc(100% - ${peekHeight}px))`, // Adjust transform to use peekHeight
        zIndex: 2,
        overflowY: 'scroll',
      }}
    >
      {isVisible && <div className="flex justify-start w-full">
        <XMarkIcon className="h-6 w-6" onClick={onClose} />
      </div>}
      {children}
    </div>
  );
};
