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
import { createContext, useState, useEffect, useContext } from 'react';
import { useAPI } from '@poltergeist/contexts';
export const ZoneContext = createContext({
    zones: [],
    selectedZone: null,
    setSelectedZone: () => { },
    createZone: () => { },
    deleteZone: () => { },
});
export const ZoneProvider = ({ children }) => {
    const { apiClient } = useAPI();
    const [zones, setZones] = useState([]);
    const [selectedZone, setSelectedZone] = useState(null);
    const fetchZones = () => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.get('/sonar/zones');
        setZones(response);
        setSelectedZone(response[0]);
    });
    const createZone = (zone) => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.post('/sonar/zones', zone);
        setZones([...zones, response]);
        setSelectedZone(response);
    });
    const deleteZone = (zone) => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.delete(`/sonar/zones/${zone.id}`);
        setZones(zones.filter((z) => z.id !== zone.id));
    });
    useEffect(() => {
        fetchZones();
    }, []);
    return (_jsx(ZoneContext.Provider, Object.assign({ value: {
            zones,
            selectedZone,
            setSelectedZone,
            createZone,
            deleteZone
        } }, { children: children })));
};
export const useZoneContext = () => {
    return useContext(ZoneContext);
};
