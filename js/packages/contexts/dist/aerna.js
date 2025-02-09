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
import { createContext, useContext, useState, useCallback } from 'react';
const PointOfInterestGroupContext = createContext(undefined);
export const ArenaProvider = ({ children }) => {
    const [arena, setArena] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const fetchGroups = useCallback(() => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            // TODO: Replace with actual API call
            const response = yield fetch('/api/point-of-interest-groups');
            if (!response.ok)
                throw new Error('Failed to fetch groups');
            const data = yield response.json();
            setGroups(data);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    }), []);
    const createGroup = useCallback((group) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            // TODO: Replace with actual API call
            const response = yield fetch('/api/point-of-interest-groups', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(group),
            });
            if (!response.ok)
                throw new Error('Failed to create group');
            const newGroup = yield response.json();
            setGroups(prev => [...prev, newGroup]);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
            throw err;
        }
        finally {
            setLoading(false);
        }
    }), []);
    const updateGroup = useCallback((id, group) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            // TODO: Replace with actual API call
            const response = yield fetch(`/api/point-of-interest-groups/${id}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(group),
            });
            if (!response.ok)
                throw new Error('Failed to update group');
            const updatedGroup = yield response.json();
            setGroups(prev => prev.map(g => g.id === id ? updatedGroup : g));
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
            throw err;
        }
        finally {
            setLoading(false);
        }
    }), []);
    const deleteGroup = useCallback((id) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            // TODO: Replace with actual API call
            const response = yield fetch(`/api/point-of-interest-groups/${id}`, {
                method: 'DELETE',
            });
            if (!response.ok)
                throw new Error('Failed to delete group');
            setGroups(prev => prev.filter(g => g.id !== id));
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
            throw err;
        }
        finally {
            setLoading(false);
        }
    }), []);
    return (_jsx(PointOfInterestGroupContext.Provider, Object.assign({ value: {
            groups,
            loading,
            error,
            fetchGroups,
            createGroup,
            updateGroup,
            deleteGroup,
        } }, { children: children })));
};
export const usePointOfInterestGroup = () => {
    const context = useContext(PointOfInterestGroupContext);
    if (context === undefined) {
        throw new Error('usePointOfInterestGroup must be used within a PointOfInterestGroupProvider');
    }
    return context;
};
