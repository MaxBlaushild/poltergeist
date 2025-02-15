import React, { useState, useEffect } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import { ImageBadge } from '../components/shared/ImageBadge.tsx';
import { useMap } from '@poltergeist/contexts';
import { useLocation } from '@poltergeist/contexts';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

export const useUserLocator = () => {
  const { map } = useMap();
  const { location } = useLocation();
  const { currentUser } = useUserProfiles();
  const [userLocator, setUserLocator] = useState<mapboxgl.Marker | null>(null);

  useEffect(() => {
    if (userLocator) {
      userLocator.setLngLat([location?.longitude ?? 0, location?.latitude ?? 0]);
      return;
    }

    if (!map?.current || !location?.longitude || !location?.latitude) {
      return;
    }

    const initializeMarker = () => {
      const locatorDiv = document.createElement('div');
      createRoot(locatorDiv).render(
        <ImageBadge
          imageUrl={currentUser?.profilePictureUrl ?? '/blank-avatar.webp'}
          onClick={() => {}}
          hasBorder={true}
        />
      );

      const newLocator = new mapboxgl.Marker({
        element: locatorDiv,
        anchor: 'center'
      });
      newLocator.setLngLat([location?.longitude ?? 0, location?.latitude ?? 0]).addTo(map.current);
      setUserLocator(newLocator);
    };

    if (!map.current.isStyleLoaded()) {
      map.current.once('load', initializeMarker);
    } else {
      initializeMarker();
    }

    return () => {
      if (userLocator) {
        userLocator.remove();
      }
    };
  }, [location, map, userLocator, currentUser]);

  return userLocator;
};
