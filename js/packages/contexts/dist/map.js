import { jsx as _jsx } from "react/jsx-runtime";
import { useEffect, useRef, useState, createContext, useContext, useCallback } from 'react';
import { useLocation } from './location';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
const MapContext = createContext(undefined);
mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';
export const MapProvider = ({ children }) => {
    const { location } = useLocation();
    const mapContainerRef = useRef(document.createElement('div'));
    const map = useRef();
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
                var _a, _b;
                setZoom((_b = (_a = map.current) === null || _a === void 0 ? void 0 : _a.getZoom()) !== null && _b !== void 0 ? _b : 0);
            });
            map.current.on('move', () => {
                var _a, _b, _c, _d, _e, _f;
                setLng((_b = (_a = map.current) === null || _a === void 0 ? void 0 : _a.getCenter().lng) !== null && _b !== void 0 ? _b : 0);
                setLat((_d = (_c = map.current) === null || _c === void 0 ? void 0 : _c.getCenter().lat) !== null && _d !== void 0 ? _d : 0);
                setZoom((_f = (_e = map.current) === null || _e === void 0 ? void 0 : _e.getZoom()) !== null && _f !== void 0 ? _f : 0);
            });
        }
    }, []);
    useEffect(() => {
        if (map.current && (location === null || location === void 0 ? void 0 : location.longitude) && (location === null || location === void 0 ? void 0 : location.latitude) && !isMapLoaded) {
            map.current.setCenter([location.longitude, location.latitude]);
            setIsMapLoaded(true);
        }
    }, [location === null || location === void 0 ? void 0 : location.longitude, location === null || location === void 0 ? void 0 : location.latitude, isMapLoaded, map.current]);
    const setLocation = useCallback((lat, lng) => {
        if (map.current) {
            map.current.setCenter([lng, lat]);
        }
    }, [map]);
    const flyToLocation = useCallback((lat, lng, zoom) => {
        if (map.current) {
            map.current.flyTo({
                center: [lng, lat],
                zoom: zoom || map.current.getZoom(),
                essential: true,
                duration: 1000
            });
        }
    }, [map]);
    return (_jsx(MapContext.Provider, Object.assign({ value: {
            map,
            mapContainer: mapContainerRef,
            zoom,
            setZoom,
            setLocation,
            flyToLocation,
            lng, lat
        } }, { children: children })));
};
export const useMap = () => {
    const context = useContext(MapContext);
    if (context === undefined) {
        throw new Error('useMap must be used within a MapProvider');
    }
    return context;
};
