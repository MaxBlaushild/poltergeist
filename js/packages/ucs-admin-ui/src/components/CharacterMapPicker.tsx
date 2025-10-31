import React, { useEffect, useRef, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

mapboxgl.accessToken = 'REDACTED';

interface CharacterMapPickerProps {
  latitude: number;
  longitude: number;
  onChange: (lat: number, lng: number) => void;
}

export const CharacterMapPicker: React.FC<CharacterMapPickerProps> = ({ latitude, longitude, onChange }) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const marker = useRef<mapboxgl.Marker | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);

  // Default to NYC if no coordinates provided
  const defaultLat = 40.7128;
  const defaultLng = -74.0060;
  const initialLat = latitude || defaultLat;
  const initialLng = longitude || defaultLng;

  useEffect(() => {
    if (!mapContainer.current || map.current) return;

    // Initialize map
    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/maxblaushild/clzq7o8pr00ce01qgey4y0g31',
      center: [initialLng, initialLat],
      zoom: 16,
    });

    // Add navigation controls
    map.current.addControl(new mapboxgl.NavigationControl());

    // Create draggable marker
    const el = document.createElement('div');
    el.className = 'custom-marker';
    el.style.width = '32px';
    el.style.height = '32px';
    el.style.backgroundImage = 'url(https://docs.mapbox.com/mapbox-gl-js/assets/custom_marker.png)';
    el.style.backgroundSize = 'cover';
    el.style.cursor = 'grab';

    marker.current = new mapboxgl.Marker({ element: el, draggable: true })
      .setLngLat([initialLng, initialLat])
      .addTo(map.current);

    // Handle marker drag end
    marker.current.on('dragend', () => {
      const lngLat = marker.current!.getLngLat();
      onChange(lngLat.lat, lngLat.lng);
    });

    // Handle map click to move marker
    map.current.on('click', (e) => {
      if (marker.current) {
        marker.current.setLngLat([e.lngLat.lng, e.lngLat.lat]);
        onChange(e.lngLat.lat, e.lngLat.lng);
      }
    });

    // Track when map is loaded
    map.current.on('load', () => {
      setIsLoaded(true);
    });

    // Cleanup
    return () => {
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, [initialLat, initialLng, onChange]);

  // Update marker position when props change
  useEffect(() => {
    if (map.current && isLoaded && marker.current) {
      const currentLngLat = marker.current.getLngLat();
      // Only update if the position actually changed
      if (Math.abs(currentLngLat.lat - initialLat) > 0.0001 || 
          Math.abs(currentLngLat.lng - initialLng) > 0.0001) {
        marker.current.setLngLat([initialLng, initialLat]);
      }
    }
  }, [latitude, longitude, isLoaded, initialLat, initialLng]);

  return (
    <div style={{ position: 'relative' }}>
      <div
        ref={mapContainer}
        style={{
          width: '100%',
          height: '400px',
          borderRadius: '8px',
          border: '1px solid #ccc',
          overflow: 'hidden',
        }}
      />
      <div style={{ 
        marginTop: '8px', 
        fontSize: '14px', 
        color: '#666',
        display: 'flex',
        justifyContent: 'space-between'
      }}>
        <span>Latitude: {latitude ? latitude.toFixed(6) : 'Not set'}</span>
        <span>Longitude: {longitude ? longitude.toFixed(6) : 'Not set'}</span>
      </div>
      <p style={{ 
        marginTop: '4px', 
        fontSize: '12px', 
        color: '#999',
        fontStyle: 'italic'
      }}>
        Click on the map or drag the marker to set the character's position
      </p>
    </div>
  );
};
