import { MutableRefObject } from 'react';
interface UseMapReturn {
    map: MutableRefObject<mapboxgl.Map | undefined>;
    mapContainer: MutableRefObject<HTMLDivElement>;
}
export declare const useMap: () => UseMapReturn;
export {};
