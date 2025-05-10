import { MutableRefObject, useEffect, useRef, useState, createContext, useContext, ReactNode, useCallback } from 'react';
import { useLocation } from './location';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import 'mapbox-gl/dist/mapbox-gl.css';

interface MapContextValue {
  map: MutableRefObject<mapboxgl.Map | undefined>;
  mapContainer: MutableRefObject<HTMLDivElement>;
  zoom: number;
  setZoom: (zoom: number) => void;
  setLocation: (lat: number, lng: number) => void;
  flyToLocation: (lat: number, lng: number, zoom?: number) => void;
  lng: number;
  lat: number;
}

const MapContext = createContext<MapContextValue | undefined>(undefined);

mapboxgl.accessToken = 'pk.eyJ1IjoibWF4YmxhdXNoaWxkIiwiYSI6ImNsenE2YWY2bDFmNnQyam9jOXJ4dHFocm4ifQ.tvO7DVEK_OLUyHfwDkUifA';

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
  const [isMapLoaded, setIsMapLoaded] = useState(false);

  useEffect(() => {
    if (!map.current) {
      map.current = new mapboxgl.Map({
        container: mapContainerRef.current,
        style: 'mapbox://styles/maxblaushild/clzq7o8pr00ce01qgey4y0g31',
        center: [0, 0],
        zoom: 16,
      });

      const geolocateControl = new mapboxgl.GeolocateControl({
        positionOptions: {
          enableHighAccuracy: true
        },
        trackUserLocation: true,
        showUserHeading: true
      });
      
      map.current.addControl(geolocateControl);

      map.current.on('load', () => {
        geolocateControl.trigger();
      });

      map.current.on('zoom', () => {
        setZoom(map.current?.getZoom() ?? 0);
      });

      map.current.on('move', () => {
        setLng(map.current?.getCenter().lng ?? 0);
        setLat(map.current?.getCenter().lat ?? 0);
        setZoom(map.current?.getZoom() ?? 0);
      });
    }
  }, []);

  useEffect(() => {
    console.log('useEffect', location?.longitude, location?.latitude, isMapLoaded, map.current);
    if (map.current && location?.longitude && location?.latitude && !isMapLoaded) {
      console.log('setting center', location.longitude, location.latitude);
      map.current.setCenter([location.longitude, location.latitude]);
      setIsMapLoaded(true);
    }
  }, [location?.longitude, location?.latitude, isMapLoaded, map.current]);

  const setLocation = useCallback((lat: number, lng: number) => {
    if (map.current) {
      map.current.setCenter([lng, lat]);
    }
  }, [map]);

  const flyToLocation = useCallback((lat: number, lng: number, zoom?: number) => {
    if (map.current) {
      map.current.flyTo({
        center: [lng, lat],
        zoom: zoom || map.current.getZoom(),
        essential: true,
        duration: 1000
      });
    }
  }, [map]);

  return (
    <MapContext.Provider value={{ 
      map, 
      mapContainer: mapContainerRef, 
      zoom, 
      setZoom, 
      setLocation, 
      flyToLocation,
      lng, lat }}>
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
