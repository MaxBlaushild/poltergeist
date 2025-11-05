/**
 * Returns the base pin size for a marker based on zoom level.
 * This is the internal size value used by markers.
 * 
 * @param zoom - The map zoom level
 * @returns The base pin size (4, 5, 6, 8, or 16)
 */
export const getMarkerPinSize = (zoom: number): number => {
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
      return 4;
    case 9:
    case 10:
      return 5;
    case 11:
      return 6;
    case 12:
    case 13:
    case 14:
      return 8;
    case 15:
    case 16:
    case 17:
    case 18:
    case 19:
    default:
      return 16;
  }
};

/**
 * Returns the actual pixel size for a marker based on zoom level.
 * This is the size that should be used for width/height in pixels.
 * 
 * @param zoom - The map zoom level
 * @returns The pixel size (pinSize * 2)
 */
export const getMarkerPixelSize = (zoom: number): number => {
  return getMarkerPinSize(zoom) * 2;
};

