var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { useState, useEffect } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
export const useUserLevel = () => {
    const { apiClient } = useAPI();
    const { user } = useAuth();
    const [userLevel, setUserLevel] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    useEffect(() => {
        if (!user) {
            // Clear data when not authenticated
            setUserLevel(null);
            setLoading(false);
            return;
        }
        const fetchUserLevel = () => __awaiter(void 0, void 0, void 0, function* () {
            var _a, _b;
            try {
                const fetchedUserLevel = yield apiClient.get(`/sonar/level`);
                setUserLevel(fetchedUserLevel);
            }
            catch (error) {
                // Silently handle auth errors
                if (((_a = error === null || error === void 0 ? void 0 : error.response) === null || _a === void 0 ? void 0 : _a.status) === 401 || ((_b = error === null || error === void 0 ? void 0 : error.response) === null || _b === void 0 ? void 0 : _b.status) === 403) {
                    setUserLevel(null);
                    return;
                }
                setError(error);
            }
            finally {
                setLoading(false);
            }
        });
        fetchUserLevel();
    }, [apiClient, user]);
    return {
        userLevel,
        loading,
        error,
    };
};
