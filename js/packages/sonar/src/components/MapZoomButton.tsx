import React from 'react';
import { useMap } from '@poltergeist/contexts';
import { useLocation } from '@poltergeist/contexts';
import { MapPinIcon } from '@heroicons/react/24/solid';

export const MapZoomButton = () => {
    const { zoom } = useMap();
    const { location } = useLocation();
    const { map } = useMap();
    
    return (
      <div
        className="absolute top-20 right-4 z-10 bg-white rounded-lg p-2 border-2 border-black opacity-80"
        onClick={() => {
          console.log('clicked');
          if (location?.longitude && location?.latitude) {
            const newCenter = [location.longitude, location.latitude];
            map.current?.flyTo({ center: newCenter, zoom: 15 });
            return;
          }
          if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(
              (position) => {
                const newCenter = [
                  position.coords.longitude,
                  position.coords.latitude,
                ];
                map.current?.flyTo({ center: newCenter, zoom: 15 });
              },
              (error) => {
                console.log('error', error);
              }
            );
          }
        }}
      >
        <MapPinIcon className="w-6 h-6" />
      </div>
    );
  };
  