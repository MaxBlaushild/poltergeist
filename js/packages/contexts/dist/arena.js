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
import { createContext, useContext, useState, useEffect } from 'react';
import { useMediaContext } from './media';
import { useAPI } from './api';
import { useAuth } from './auth';
const ArenaContext = createContext(undefined);
export const ArenaProvider = ({ children, arenaId }) => {
    const [arena, setArena] = useState(null);
    const { apiClient } = useAPI();
    const { user } = useAuth();
    const mediaContext = useMediaContext();
    if (!mediaContext) {
        throw new Error('ArenaProvider must be wrapped in a MediaProvider');
    }
    const { uploadMedia, getPresignedUploadURL } = mediaContext;
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [pointOfInterestZones, setPointOfInterestZones] = useState({});
    const fetchArena = (arenaId) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.get(`/sonar/pointsOfInterest/group/${arenaId}`);
            setArena(response);
        }
        catch (err) {
            console.error('Error fetching arena', err);
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const updateArena = (name, description, type, gold, inventoryItemId) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        if (!arena) {
            return;
        }
        try {
            const response = yield apiClient.patch(`/sonar/pointsOfInterest/group/${arenaId}`, {
                name,
                description,
                type,
                gold,
                inventoryItemId,
            });
            setArena(Object.assign(Object.assign({}, arena), { name,
                description,
                type, gold: gold !== null && gold !== void 0 ? gold : arena.gold, inventoryItemId: inventoryItemId !== null && inventoryItemId !== void 0 ? inventoryItemId : arena.inventoryItemId }));
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const addTagToPointOfInterest = (tagId, pointOfInterestId) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.post(`/sonar/tags/add`, {
                tagId,
                pointOfInterestId,
            });
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const removeTagFromPointOfInterest = (tagId, pointOfInterestId) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.delete(`/sonar/tags/${tagId}/pointOfInterest/${pointOfInterestId}`);
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const updateArenaImage = (id, image) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        const imageKey = `arenas/${((image === null || image === void 0 ? void 0 : image.name) || 'image.jpg').toLowerCase().replace(/\s+/g, '-')}`;
        let imageUrl = '';
        if (image) {
            const presignedUrl = yield getPresignedUploadURL("crew-points-of-interest", imageKey);
            if (!presignedUrl)
                return;
            yield uploadMedia(presignedUrl, image);
            imageUrl = presignedUrl.split("?")[0];
        }
        try {
            const response = yield apiClient.patch(`/sonar/pointsofInterest/group/imageUrl/${id}`, {
                imageUrl,
            });
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const createPointOfInterest = (name, description, lat, lng, image, clue) => __awaiter(void 0, void 0, void 0, function* () {
        if (!name || !description || !lat || !lng || !image || !clue || !arenaId) {
            return;
        }
        setLoading(true);
        const key = `${encodeURIComponent(name)}${image.name.substring(image.name.lastIndexOf('.'))}`;
        const imageUrl = yield getPresignedUploadURL('crew-points-of-interest', key);
        if (!imageUrl) {
            return;
        }
        const uploadResult = yield uploadMedia(imageUrl, image);
        if (!uploadResult) {
            return;
        }
        try {
            const res = yield apiClient.post(`/sonar/pointsOfInterest/group/${arenaId}`, {
                name,
                description,
                latitude: JSON.stringify(lat),
                longitude: JSON.stringify(lng),
                imageUrl: imageUrl.split("?")[0],
                clue,
                pointOfInterestGroupId: arenaId,
            });
            fetchArena(arenaId);
        }
        catch (error) {
            console.error('Error creating point:', error);
        }
        finally {
            setLoading(false);
        }
    });
    const updatePointOfInterest = (id, arena) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.patch(`/sonar/pointsOfInterest/${id}`, Object.assign(Object.assign({}, arena), { lat: typeof arena.lat === 'string' ? arena.lat : JSON.stringify(arena.lat), lng: typeof arena.lng === 'string' ? arena.lng : JSON.stringify(arena.lng) }));
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const updatePointOfInterestImage = (id, image) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        const imageKey = `arenas/${((image === null || image === void 0 ? void 0 : image.name) || 'image.jpg').toLowerCase().replace(/\s+/g, '-')}`;
        let imageUrl = '';
        if (image) {
            const presignedUrl = yield getPresignedUploadURL("crew-points-of-interest", imageKey);
            if (!presignedUrl)
                return;
            yield uploadMedia(presignedUrl, image);
            imageUrl = presignedUrl.split("?")[0];
        }
        try {
            const response = yield apiClient.patch(`/sonar/pointsofInterest/imageUrl/${id}`, {
                imageUrl,
            });
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const deletePointOfInterest = (id) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.delete(`/sonar/pointsOfInterest/${id}`);
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const createPointOfInterestChallenge = (id, challenge) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.post(`/sonar/pointsOfInterest/challenge`, Object.assign(Object.assign({}, challenge), { pointOfInterestId: id }));
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const updatePointOfInterestChallenge = (id, challenge) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.patch(`/sonar/pointsOfInterest/challenge/${id}`, Object.assign({}, challenge));
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const deletePointOfInterestChallenge = (id) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.delete(`/sonar/pointsOfInterest/challenge/${id}`);
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const createPointOfInterestChildren = (pointOfInterestId, pointOfInterestGroupMemberId, pointOfInterestChallengeId) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.post(`/sonar/pointOfInterest/children`, {
                pointOfInterestId,
                pointOfInterestGroupMemberId,
                pointOfInterestChallengeId,
            });
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const deletePointOfInterestChildren = (id) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            const response = yield apiClient.delete(`/sonar/pointOfInterest/children/${id}`);
            fetchArena(arenaId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
        }
        finally {
            setLoading(false);
        }
    });
    const getZoneForPointOfInterest = (pointOfInterestId) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const response = yield apiClient.get(`/sonar/pointOfInterest/${pointOfInterestId}/zone`);
            setPointOfInterestZones(prev => (Object.assign(Object.assign({}, prev), { [pointOfInterestId]: response })));
            return response;
        }
        catch (err) {
            // POI might not be in any zone, which is fine
            console.log('POI not in any zone or error fetching zone:', err);
            return null;
        }
    });
    const addPointOfInterestToZone = (zoneId, pointOfInterestId) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            yield apiClient.post(`/sonar/zones/${zoneId}/pointOfInterest/${pointOfInterestId}`);
            yield getZoneForPointOfInterest(pointOfInterestId);
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
            throw err;
        }
        finally {
            setLoading(false);
        }
    });
    const removePointOfInterestFromZone = (zoneId, pointOfInterestId) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        try {
            yield apiClient.delete(`/sonar/zones/${zoneId}/pointOfInterest/${pointOfInterestId}`);
            setPointOfInterestZones(prev => {
                const newZones = Object.assign({}, prev);
                delete newZones[pointOfInterestId];
                return newZones;
            });
        }
        catch (err) {
            setError(err instanceof Error ? err : new Error('An error occurred'));
            throw err;
        }
        finally {
            setLoading(false);
        }
    });
    useEffect(() => {
        if (!user || !arenaId) {
            setArena(null);
            return;
        }
        fetchArena(arenaId);
    }, [arenaId, user]);
    return (_jsx(ArenaContext.Provider, Object.assign({ value: {
            arena,
            loading,
            error,
            updateArena,
            updateArenaImage,
            createPointOfInterest,
            updatePointOfInterest,
            updatePointOfInterestImage,
            deletePointOfInterest,
            createPointOfInterestChallenge,
            updatePointOfInterestChallenge,
            deletePointOfInterestChallenge,
            createPointOfInterestChildren,
            deletePointOfInterestChildren,
            addTagToPointOfInterest,
            removeTagFromPointOfInterest,
            getZoneForPointOfInterest,
            addPointOfInterestToZone,
            removePointOfInterestFromZone,
            pointOfInterestZones,
        } }, { children: children })));
};
export const useArena = () => {
    const context = useContext(ArenaContext);
    if (context === undefined) {
        throw new Error('useArena must be used within a ArenaProvider');
    }
    return context;
};
