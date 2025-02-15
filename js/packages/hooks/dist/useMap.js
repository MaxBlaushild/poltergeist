import { useEffect, useRef } from 'react';
import { useLocation } from '@poltergeist/contexts';
import mapboxgl from 'mapbox-gl';
mapboxgl.accessToken = 'pk.eyJ1IjoibWF4YmxhdXNoaWxkIiwiYSI6ImNsenE2YWY2bDFmNnQyam9jOXJ4dHFocm4ifQ.tvO7DVEK_OLUyHfwDkUifA';
export const useMap = () => {
    const { location } = useLocation();
    const mapContainerRef = useRef(document.createElement('div'));
    const map = useRef();
    useEffect(() => {
        var _a, _b;
        if (!map.current) {
            map.current = new mapboxgl.Map({
                container: mapContainerRef.current,
                style: 'mapbox://styles/maxblaushild/clzq7o8pr00ce01qgey4y0g31',
                center: [(_a = location === null || location === void 0 ? void 0 : location.longitude) !== null && _a !== void 0 ? _a : 0, (_b = location === null || location === void 0 ? void 0 : location.latitude) !== null && _b !== void 0 ? _b : 0],
                zoom: 16,
            });
        }
    }, []); // Empty dependency array means this runs once on mount
    return { map, mapContainer: mapContainerRef };
};
