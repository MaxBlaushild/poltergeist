import { XMarkIcon } from '@heroicons/react/20/solid';
import './Chip.css';
import React from 'react';

export enum ChipType {
  ACTIVITY = 'activity',
  PERSON = 'person',
}

interface ChipProps {
  label: string;
  onDelete: () => void;
  type: ChipType;
}

export const Chip = ({ label, onDelete, type }: ChipProps) => {
  const chipClass =
    type === ChipType.ACTIVITY ? 'Chip__activity' : 'Chip__person';
  return (
    <div className={`Chip__chip ${chipClass}`}>
      <span>{label}</span>
      <button>
        <XMarkIcon className="h-5 w-5 text-black" onClick={onDelete} />
      </button>
    </div>
  );
};
