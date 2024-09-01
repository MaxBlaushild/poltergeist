export const calculateDistance = (poi1, poi2) => {
  const R = 6371e3; // Earth radius in meters
  const lat1 = (poi1.lat * Math.PI) / 180;
  const lat2 = (poi2.lat * Math.PI) / 180;
  const deltaLat = ((poi2.lat - poi1.lat) * Math.PI) / 180;
  const deltaLng = ((poi2.lng - poi1.lng) * Math.PI) / 180;

  const a =
    Math.sin(deltaLat / 2) * Math.sin(deltaLat / 2) +
    Math.cos(lat1) *
      Math.cos(lat2) *
      Math.sin(deltaLng / 2) *
      Math.sin(deltaLng / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

  return R * c; // Distance in meters
};
