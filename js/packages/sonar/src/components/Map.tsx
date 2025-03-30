import React, { useEffect, useState, useId } from 'react';
import { useMap } from '@poltergeist/contexts';
import './Map.css';
interface MapProps {
  children: React.ReactNode;
}

export const Map = ({ children }: MapProps) => {
  const { map, mapContainer } = useMap();
  const [isLoaded, setIsLoaded] = useState<boolean>(false);

  useEffect(() => {
    if (!mapContainer?.current || !map.current?.isStyleLoaded() || isLoaded) return;
    const mapElement = mapContainer.current;
    
    const parentElement = document.getElementById(`map-parent`);
    if (parentElement) {
      parentElement.appendChild(mapElement);
      setIsLoaded(true);
    }
  }, [map, mapContainer, isLoaded]);

  return (
      <div
        id={`map-parent`}
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
