var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { useAPI } from '@poltergeist/contexts';
import { useState, useEffect } from 'react';
export const useCandidates = (query) => {
    const [candidates, setCandidates] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const { apiClient } = useAPI();
    useEffect(() => {
        const fetchPlaces = () => __awaiter(void 0, void 0, void 0, function* () {
            try {
                const data = yield apiClient.get(`/sonar/google/places?query=${query}`);
                setCandidates(data);
            }
            catch (err) {
                setError(err);
            }
            finally {
                setLoading(false);
            }
        });
        if (query) {
            fetchPlaces();
        }
    }, [query]);
    return {
        candidates,
        loading,
        error
    };
};
