import { useMap } from '@poltergeist/contexts';
import { useZoneContext } from '@poltergeist/contexts/dist/zones';
import { useEffect, useState } from 'react';

export const useZoneBoundaries = () => {
  const { map } = useMap();
  const { zones, selectedZone, setSelectedZone } = useZoneContext();

  const createBoundaries = () => {
    // Remove any existing boundary layers
    if (map.current?.getLayer('zone-boundaries')) {
      map.current.removeLayer('zone-boundaries');
    }
    if (map.current?.getLayer('zone-boundaries-glow')) {
      map.current.removeLayer('zone-boundaries-glow');
    }
    if (map.current?.getLayer('zone-boundaries-fill')) {
      map.current.removeLayer('zone-boundaries-fill');
    }
    if (map.current?.getSource('zone-boundaries')) {
      map.current.removeSource('zone-boundaries');
    }

    const boundaries = zones?.map((zone) => {
      // Create a line feature from the zone's points
      const coordinates = zone.points.map(point => [point.longitude, point.latitude]);
      
      // Close the line by adding first point at end
      if (coordinates.length > 0) {
        coordinates.push(coordinates[0]);
      }

      return {
        type: 'Feature',
        properties: {
          id: zone.id,
          name: zone.name
        },
        geometry: {
          type: 'Polygon',
          coordinates: [coordinates]
        }
      };
    });

    // Add the boundaries to the map
    map.current?.addSource('zone-boundaries', {
      type: 'geojson',
      data: {
        type: 'FeatureCollection',
        features: boundaries || []
      }
    });

    // Add fill layer
    map.current?.addLayer({
      id: 'zone-boundaries-fill',
      type: 'fill',
      source: 'zone-boundaries',
      layout: {},
      paint: {
        'fill-color': '#4A90E2',
        'fill-opacity': ['case',
          ['all', ['has', 'id'], ['==', ['get', 'id'], selectedZone?.id || '']],
          0.1,
          0
        ]
      }
    });

    // Add glow layer
    map.current?.addLayer({
      id: 'zone-boundaries-glow',
      type: 'line',
      source: 'zone-boundaries',
      layout: {},
      paint: {
        'line-color': '#4A90E2',
        'line-width': ['case',
          ['all', ['has', 'id'], ['==', ['get', 'id'], selectedZone?.id || '']],
          8,
          0
        ],
        'line-opacity': 0.6,
        'line-blur': 3
      }
    });

    // Add main boundary layer
    map.current?.addLayer({
      id: 'zone-boundaries',
      type: 'line',
      source: 'zone-boundaries',
      layout: {},
      paint: {
        'line-color': ['case',
          ['all', ['has', 'id'], ['==', ['get', 'id'], selectedZone?.id || '']],
          '#4A90E2',
          '#000'
        ],
        'line-width': ['case',
          ['all', ['has', 'id'], ['==', ['get', 'id'], selectedZone?.id || '']],
          3,
          2
        ]
      }
    });

    // Add click handlers for both boundary line and fill
    const handleZoneClick = (e) => {
      if (e.features?.[0]?.properties?.id) {
        const clickedZone = zones.find(zone => zone.id === e.features[0].properties.id);
        if (clickedZone) {
          setSelectedZone(clickedZone);
        }
      }
    };

    map.current?.on('click', 'zone-boundaries', handleZoneClick);
    map.current?.on('click', 'zone-boundaries-fill', handleZoneClick);
  };

  useEffect(() => {
    if (map.current && zones?.length && map.current?.isStyleLoaded()) {
      createBoundaries();
    } else {
      const timer = setInterval(() => {
        if (map.current && zones?.length && map.current?.isStyleLoaded()) {
          createBoundaries();
          clearInterval(timer);
        }
      }, 100);

      return () => clearInterval(timer);
    }
  }, [zones, map, map?.current, selectedZone]);
};