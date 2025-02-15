import React, { useEffect, useState } from 'react';
import { useMap } from '@poltergeist/contexts';

interface MapProps {
  children: React.ReactNode;
}

export const Map = ({ children }: MapProps) => {
  const { map, mapContainer } = useMap();
  const [isLoaded, setIsLoaded] = useState<boolean>(false);

  useEffect(() => {
    if (!mapContainer?.current || !map.current?.isStyleLoaded() || isLoaded) return;
    const parentElement = document.getElementById('map-parent');
    if (parentElement) {
      parentElement.appendChild(mapContainer.current);
      setIsLoaded(true);
    }

    return () => {
      if (mapContainer.current) {
        mapContainer.current.remove();
      }
    };
  }, [mapContainer?.current, map?.current, isLoaded]);

  return (
      <div
        id="map-parent"
        ref={mapContainer}
        style={{
          top: -70,
          left: 0,
          width: '100vw',
          height: '100vh',
          zIndex: 1,
        }}
      >
        {children}
      </div>
  );
};
