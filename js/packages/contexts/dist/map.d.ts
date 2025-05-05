import { MutableRefObject, ReactNode } from 'react';
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
interface MapProviderProps {
    children: ReactNode;
}
export declare const MapProvider: ({ children }: MapProviderProps) => import("react/jsx-runtime").JSX.Element;
export declare const useMap: () => MapContextValue;
export {};
