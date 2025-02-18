import { jsx as _jsx } from "react/jsx-runtime";
import { useEffect, useRef, useState, createContext, useContext } from 'react';
import { useLocation } from './location';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
const MapContext = createContext(undefined);
mapboxgl.accessToken = 'pk.eyJ1IjoibWF4YmxhdXNoaWxkIiwiYSI6ImNsenE2YWY2bDFmNnQyam9jOXJ4dHFocm4ifQ.tvO7DVEK_OLUyHfwDkUifA';
export const MapProvider = ({ children }) => {
    const { location } = useLocation();
    const mapContainerRef = useRef(document.createElement('div'));
    const map = useRef();
    const [zoom, setZoom] = useState(16);
    const [lng, setLng] = useState(0);
    const [lat, setLat] = useState(0);
    const [isMapInitialized, setIsMapInitialized] = useState(false);
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
    useEffect(() => {
        var _a, _b, _c;
        if (isMapInitialized)
            return;
        if ((location === null || location === void 0 ? void 0 : location.longitude) && (location === null || location === void 0 ? void 0 : location.latitude) && (map === null || map === void 0 ? void 0 : map.current)) {
            setIsMapInitialized(true);
            (_a = map.current) === null || _a === void 0 ? void 0 : _a.setCenter([location.longitude, location.latitude]);
            (_b = map.current) === null || _b === void 0 ? void 0 : _b.on('zoom', () => {
                var _a, _b;
                setZoom((_b = (_a = map.current) === null || _a === void 0 ? void 0 : _a.getZoom()) !== null && _b !== void 0 ? _b : 0);
            });
            (_c = map.current) === null || _c === void 0 ? void 0 : _c.on('move', () => {
                var _a, _b, _c, _d, _e, _f;
                setLng((_b = (_a = map.current) === null || _a === void 0 ? void 0 : _a.getCenter().lng) !== null && _b !== void 0 ? _b : 0);
                setLat((_d = (_c = map.current) === null || _c === void 0 ? void 0 : _c.getCenter().lat) !== null && _d !== void 0 ? _d : 0);
                setZoom((_f = (_e = map.current) === null || _e === void 0 ? void 0 : _e.getZoom()) !== null && _f !== void 0 ? _f : 0);
            });
        }
    }, [location === null || location === void 0 ? void 0 : location.longitude, location === null || location === void 0 ? void 0 : location.latitude, map === null || map === void 0 ? void 0 : map.current, isMapInitialized]);
    return (_jsx(MapContext.Provider, Object.assign({ value: { map, mapContainer: mapContainerRef, zoom, setZoom } }, { children: children })));
};
export const useMap = () => {
    const context = useContext(MapContext);
    if (context === undefined) {
        throw new Error('useMap must be used within a MapProvider');
    }
    return context;
};
