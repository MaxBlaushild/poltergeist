import React from 'react';

interface QuestLogProps {
  onClose: () => void;
}

export const QuestLog: React.FC<QuestLogProps> = ({ onClose }) => {
  return <div>Quest Log</div>;
};
