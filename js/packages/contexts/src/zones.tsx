import React, { createContext, useContext, useState, useEffect, useRef } from 'react';
import RBush from 'rbush';
import { Zone } from '@poltergeist/types';
import * as turf from '@turf/turf';
import { useAPI, useLocation, useAuth } from '@poltergeist/contexts';

export const calculateDistance = (poi1, poi2) => {
  const R = 6371e3; // Earth radius in meters
  const lat1 = (poi1.lat * Math.PI) / 180;
  const lat2 = (poi2.lat * Math.PI) / 180;
  const deltaLat = ((poi2.lat - poi1.lat) * Math.PI) / 180;
  const deltaLng = ((poi2.lng - poi1.lng) * Math.PI) / 180;

  const a =
    Math.sin(deltaLat / 2) * Math.sin(deltaLat / 2) +
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

// Define the structure for R-tree nodes
interface ZoneIndexNode {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
  zoneId: string;
  points: Array<[number, number]>; // [longitude, latitude] pairs
}

type ZoneContextType = {
  zones: Zone[];
  selectedZone: Zone | null;
  setSelectedZone: (zone: Zone | null) => void;
  createZone: (zone: Zone) => void;
  deleteZone: (zone: Zone) => void;
  findZoneAtCoordinate: (lng: number, lat: number) => Zone | null;
  editZone: (name: string, description: string, id: string) => void;
};

const ZoneContext = createContext<ZoneContextType | null>(null);

export const ZoneProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [zones, setZones] = useState<Zone[]>([]);
  const [selectedZone, setSelectedZone] = useState<Zone | null>(null);
  const [spatialIndex] = useState(() => new RBush<ZoneIndexNode>());
  const { location } = useLocation();
  const previousLocation = useRef(location);
  useEffect(() => {
    const fetchZones = async () => {
      try {
        const response = await apiClient.get<Zone[]>('/sonar/zones');
        setZones(response);
      } catch (error) {
        console.error('Error fetching zones:', error);
      }
    };
    fetchZones();
  }, [user]);

  // Update spatial index when zones change
  useEffect(() => {
    // Clear existing data
    spatialIndex.clear();
    
    // Insert new data
    zones.forEach(zone => {
      if (zone.points && zone.points.length > 0) {
        // Convert points to [lng, lat] pairs
        const points = zone.points.map(p => [p.longitude, p.latitude] as [number, number]);
        
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

    if (location?.longitude && location?.latitude) {
      const zone = findZoneAtCoordinate(location.longitude, location.latitude);
      if (zone) {
        setSelectedZone(zone);
      }
    }
  }, [zones, location]);

  const findZoneAtCoordinate = (lng: number, lat: number): Zone | null => {
    if (!lng || !lat) return null;

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
    if (!location?.longitude || !location?.latitude) {
      return;
    }
    if (!selectedZone || !previousLocation.current || isXMetersAway(previousLocation.current, location, 100)) {
      const zone = findZoneAtCoordinate(location?.longitude, location?.latitude);
      if (zone) {
        setSelectedZone(zone);
      }
      previousLocation.current = location;
    }
  }, [location?.latitude, location?.longitude]);

  const createZone = async(zone: Zone) => {
    const response = await apiClient.post<Zone>('/sonar/zones', zone);
    setZones(prev => [...prev, response]);
    setSelectedZone(response);
  };

  const deleteZone = (zone: Zone) => {
    setZones(prev => prev.filter(z => z.id !== zone.id));
  };

  const editZone = async (name: string, description: string, id: string) => {
    const response = await apiClient.patch<Zone>(`/sonar/zones/${id}/edit`, { name, description });
    setZones(prev => prev.map(z => z.id === id ? {...z, name: response.name, description: response.description} : z));
  };

  return (
    <ZoneContext.Provider value={{
      zones,
      selectedZone,
      setSelectedZone,
      createZone,
      deleteZone,
      findZoneAtCoordinate,
      editZone
    }}>
      {children}
    </ZoneContext.Provider>
  );
};

export const useZoneContext = () => {
  const context = useContext(ZoneContext);
  if (!context) {
    throw new Error('useZoneContext must be used within a ZoneProvider');
  }
  return context;
};
