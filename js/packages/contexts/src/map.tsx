import { MutableRefObject, useEffect, useRef, useState, createContext, useContext, ReactNode } from 'react';
import { useLocation } from './location';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import 'mapbox-gl/dist/mapbox-gl.css';

interface MapContextValue {
  map: MutableRefObject<mapboxgl.Map | undefined>;
  mapContainer: MutableRefObject<HTMLDivElement>;
  zoom: number;
  setZoom: (zoom: number) => void;
}

const MapContext = createContext<MapContextValue | undefined>(undefined);

mapboxgl.accessToken = 'REDACTED';

interface MapProviderProps {
  children: ReactNode;
}

export const MapProvider = ({ children }: MapProviderProps) => {
  const { location } = useLocation();
  const mapContainerRef = useRef<HTMLDivElement>(document.createElement('div'));
  const map = useRef<mapboxgl.Map>();
  const [zoom, setZoom] = useState(16);
  const [lng, setLng] = useState(0);
  const [lat, setLat] = useState(0);
  const [isMapInitialized, setIsMapInitialized] = useState(false);

  useEffect(() => {
    if (!map.current) {
      map.current = new mapboxgl.Map({
        container: mapContainerRef.current,
        style: 'mapbox://styles/maxblaushild/clzq7o8pr00ce01qgey4y0g31',
        center: [location?.longitude ?? 0, location?.latitude ?? 0],
        zoom: 16,
      });
    }
  }, []); // Empty dependency array means this runs once on mount

  useEffect(() => {
    if (isMapInitialized) return;
    if (location?.longitude && location?.latitude && map?.current) {
      setIsMapInitialized(true);
      map.current?.setCenter([location.longitude, location.latitude]);

      map.current?.on('zoom', () => {
        setZoom(map.current?.getZoom() ?? 0);
      });

      map.current?.on('move', () => {
        setLng(map.current?.getCenter().lng ?? 0);
        setLat(map.current?.getCenter().lat ?? 0);
        setZoom(map.current?.getZoom() ?? 0);
      });
    }
  }, [location?.longitude, location?.latitude, map?.current, isMapInitialized]);

  return (
    <MapContext.Provider value={{ map, mapContainer: mapContainerRef, zoom, setZoom }}>
      {children}
    </MapContext.Provider>
  );
};

export const useMap = (): MapContextValue => {
  const context = useContext(MapContext);
  if (context === undefined) {
    throw new Error('useMap must be used within a MapProvider');
  }
  return context;
};
