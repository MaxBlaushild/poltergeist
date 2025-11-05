import React from 'react';
import { TreasureChest } from '@poltergeist/types';
import { getMarkerPixelSize } from '../utils/markerSize.ts';

export const TreasureChestMarker = ({
  treasureChest,
  index,
  zoom,
  onClick,
}: {
  treasureChest: TreasureChest;
  index: number;
  zoom: number;
  onClick: (e: React.MouseEvent) => void;
}) => {
  const imageUrl = 'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';

  const pixelSize = getMarkerPixelSize(zoom);

  let opacity = 1;
  if (zoom < 15) {
    opacity = 0.3;
  } else if (zoom < 16) {
    opacity = 0.5;
  } else if (zoom < 17) {
    opacity = 0.7;
  }

  return (
    <button onClick={onClick} className="marker">
      <img
        src={imageUrl}
        alt="Treasure Chest"
        className="rounded-lg border-2 transition-all duration-300"
        style={{
          width: `${pixelSize}px`,
          height: `${pixelSize}px`,
          borderColor: 'black',
          opacity,
          transform: 'scale(1)',
          filter: 'brightness(1)'
        } as React.CSSProperties}
      />
    </button>
  );
};

