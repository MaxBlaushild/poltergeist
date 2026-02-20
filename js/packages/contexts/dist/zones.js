var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useState, useEffect, useRef, useCallback } from 'react';
import RBush from 'rbush';
import * as turf from '@turf/turf';
import { useAPI, useLocation, useAuth } from '@poltergeist/contexts';
export const calculateDistance = (poi1, poi2) => {
    const R = 6371e3; // Earth radius in meters
    const lat1 = (poi1.lat * Math.PI) / 180;
    const lat2 = (poi2.lat * Math.PI) / 180;
    const deltaLat = ((poi2.lat - poi1.lat) * Math.PI) / 180;
    const deltaLng = ((poi2.lng - poi1.lng) * Math.PI) / 180;
    const a = Math.sin(deltaLat / 2) * Math.sin(deltaLat / 2) +
        Math.cos(lat1) *
            Math.cos(lat2) *
            Math.sin(deltaLng / 2) *
            Math.sin(deltaLng / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c; // Distance in meters
};
export const isXMetersAway = (poi1, poi2, x) => {
    const distance = calculateDistance(poi1, poi2);
    return distance < x;
};
const ZoneContext = createContext(null);
export const ZoneProvider = ({ children }) => {
    const { apiClient } = useAPI();
    const { user } = useAuth();
    const [zones, setZones] = useState([]);
    const [selectedZone, setSelectedZone] = useState(null);
    const [spatialIndex] = useState(() => new RBush());
    const { location } = useLocation();
    const previousLocation = useRef(location);
    const refreshZones = useCallback(() => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const response = yield apiClient.get('/sonar/zones');
            setZones(response);
        }
        catch (error) {
            console.error('Error fetching zones:', error);
        }
    }), [apiClient]);
    useEffect(() => {
        if (!user) {
            return;
        }
        refreshZones();
    }, [user, refreshZones]);
    // Update spatial index when zones change
    useEffect(() => {
        // Clear existing data
        spatialIndex.clear();
        // Insert new data
        zones.forEach(zone => {
            if (zone.points && zone.points.length > 0) {
                // Convert points to [lng, lat] pairs
                const points = zone.points.map(p => [p.longitude, p.latitude]);
                // Create a polygon from the points
                const polygon = turf.polygon([[
                        ...points,
                        points[0] // Close the polygon
                    ]]);
                // Get the bounding box of the polygon
                const bbox = turf.bbox(polygon);
                // Add to spatial index with bounding box and points
                spatialIndex.insert({
                    minX: bbox[0],
                    minY: bbox[1],
                    maxX: bbox[2],
                    maxY: bbox[3],
                    zoneId: zone.id,
                    points
                });
            }
        });
        if ((location === null || location === void 0 ? void 0 : location.longitude) && (location === null || location === void 0 ? void 0 : location.latitude)) {
            const zone = findZoneAtCoordinate(location.longitude, location.latitude);
            if (zone) {
                setSelectedZone(zone);
            }
        }
    }, [zones, location]);
    const findZoneAtCoordinate = (lng, lat) => {
        if (!lng || !lat)
            return null;
        // First, find all zones whose bounding boxes contain the point
        const candidates = spatialIndex.search({
            minX: lng,
            minY: lat,
            maxX: lng,
            maxY: lat
        });
        // Then check which of these zones actually contain the point
        for (const candidate of candidates) {
            const polygon = turf.polygon([[
                    ...candidate.points,
                    candidate.points[0] // Close the polygon
                ]]);
            const point = turf.point([lng, lat]);
            if (turf.booleanPointInPolygon(point, polygon)) {
                // Find the full zone object from our zones array
                return zones.find(z => z.id === candidate.zoneId) || null;
            }
        }
        return null;
    };
    useEffect(() => {
        if (!(location === null || location === void 0 ? void 0 : location.longitude) || !(location === null || location === void 0 ? void 0 : location.latitude)) {
            return;
        }
        if (!selectedZone || !previousLocation.current || isXMetersAway(previousLocation.current, location, 100)) {
            const zone = findZoneAtCoordinate(location === null || location === void 0 ? void 0 : location.longitude, location === null || location === void 0 ? void 0 : location.latitude);
            if (zone) {
                setSelectedZone(zone);
            }
            previousLocation.current = location;
        }
    }, [location === null || location === void 0 ? void 0 : location.latitude, location === null || location === void 0 ? void 0 : location.longitude]);
    const createZone = (zone) => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.post('/sonar/zones', zone);
        setZones(prev => [...prev, response]);
        setSelectedZone(response);
    });
    const deleteZone = (zone) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            yield apiClient.delete(`/sonar/zones/${zone.id}`);
            setZones(prev => prev.filter(z => z.id !== zone.id));
            if ((selectedZone === null || selectedZone === void 0 ? void 0 : selectedZone.id) === zone.id) {
                setSelectedZone(null);
            }
        }
        catch (error) {
            console.error('Error deleting zone:', error);
        }
    });
    const editZone = (name, description, id) => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.patch(`/sonar/zones/${id}/edit`, { name, description });
        setZones(prev => prev.map(z => z.id === id ? Object.assign(Object.assign({}, z), { name: response.name, description: response.description }) : z));
    });
    return (_jsx(ZoneContext.Provider, Object.assign({ value: {
            zones,
            selectedZone,
            setSelectedZone,
            createZone,
            deleteZone,
            findZoneAtCoordinate,
            editZone,
            refreshZones
        } }, { children: children })));
};
export const useZoneContext = () => {
    const context = useContext(ZoneContext);
    if (!context) {
        throw new Error('useZoneContext must be used within a ZoneProvider');
    }
    return context;
};
