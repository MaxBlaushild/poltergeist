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
const ArenaContext = createContext(undefined);
export const ArenaProvider = ({ children, arenaId }) => {
    const [arena, setArena] = useState(null);
    const { apiClient } = useAPI();
    const mediaContext = useMediaContext();
    if (!mediaContext) {
        throw new Error('ArenaProvider must be wrapped in a MediaProvider');
    }
    const { uploadMedia, getPresignedUploadURL } = mediaContext;
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
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
    const updateArena = (name, description) => __awaiter(void 0, void 0, void 0, function* () {
        setLoading(true);
        if (!arena) {
            return;
        }
        try {
            const response = yield apiClient.patch(`/sonar/pointsOfInterest/group/${arenaId}`, {
                name,
                description,
            });
            setArena(Object.assign(Object.assign({}, arena), { name,
                description }));
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
            const response = yield apiClient.patch(`/sonar/pointsOfInterest/${id}`, {
                arena,
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
    useEffect(() => {
        if (arenaId) {
            fetchArena(arenaId);
        }
    }, [arenaId]);
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
        } }, { children: children })));
};
export const useArena = () => {
    const context = useContext(ArenaContext);
    if (context === undefined) {
        throw new Error('useArena must be used within a ArenaProvider');
    }
    return context;
};
