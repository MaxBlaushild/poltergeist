import React from 'react';

interface ImageBadgeProps {
  imageUrl: string;
  onClick: () => void;
  hasBorder?: boolean;
}
export const ImageBadge = ({ imageUrl, onClick, hasBorder = false }: ImageBadgeProps) => {
  return (
    <div
      className={`flex justify-center items-center w-10 h-10 rounded-full overflow-hidden ${hasBorder ? 'border-2 border-black' : ''}`}
      onClick={onClick}
    >
      <img
        src={imageUrl}
        alt="Profile Icon"
        className="object-cover w-full h-full"
      />
    </div>
  );
};
