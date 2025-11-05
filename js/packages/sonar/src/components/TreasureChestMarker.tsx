import React from 'react';
import { TreasureChest } from '@poltergeist/types';

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

  let pinSize = 16;
  switch (Math.floor(zoom)) {
    case 0:
    case 1:
    case 2:
    case 3:
    case 4:
    case 5:
    case 6:
    case 7:
    case 8:
      pinSize = 4;
      break;
    case 9:
    case 10:
      pinSize = 5;
      break;
    case 11:
      pinSize = 6;
      break;
    case 12:
    case 13:
    case 14:
      pinSize = 8;
      break;
    case 15:
    case 16:
    case 17:
    case 18:
    case 19:
    default:
      pinSize = 16;
      break;
  }

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
          width: `${pinSize * 4}px`,
          height: `${pinSize * 4}px`,
          borderColor: 'black',
          opacity,
          transform: 'scale(1)',
          filter: 'brightness(1)'
        } as React.CSSProperties}
      />
    </button>
  );
};

