import React, { createContext, useState, useEffect, useContext } from 'react';
import { Zone } from '@poltergeist/types';
import { useAPI } from '@poltergeist/contexts';

type ZoneContextType = {
  zones: Zone[];
  selectedZone: Zone | null;
  setSelectedZone: (zone: Zone | null) => void;
  createZone: (zone: Zone) => void;
  deleteZone: (zone: Zone) => void;
};

export const ZoneContext = createContext<ZoneContextType>({
  zones: [],
  selectedZone: null,
  setSelectedZone: () => {},
  createZone: () => {},
  deleteZone: () => {},
});

export const ZoneProvider = ({ children }: { children: React.ReactNode }) => {
  const { apiClient } = useAPI();
  const [zones, setZones] = useState<Zone[]>([]);
  const [selectedZone, setSelectedZone] = useState<Zone | null>(null);

  const fetchZones = async () => {
    const response = await apiClient.get<Zone[]>('/sonar/zones');
    setZones(response);
    setSelectedZone(response[0]);
  };

  const createZone = async (zone: Zone) => {
    const response = await apiClient.post<Zone>('/sonar/zones', zone);
    setZones([...zones, response]);
    setSelectedZone(response);
  };

  const deleteZone = async (zone: Zone) => {
    const response = await apiClient.delete<Zone>(`/sonar/zones/${zone.id}`);
    setZones(zones.filter((z) => z.id !== zone.id));
  };

  useEffect(() => {
    fetchZones();
  }, []);

  return (
    <ZoneContext.Provider value={{ 
      zones, 
      selectedZone, 
      setSelectedZone, 
      createZone, 
      deleteZone
    }}>
      {children}
    </ZoneContext.Provider>
  );
};

export const useZoneContext = () => {
  return useContext(ZoneContext);
};
