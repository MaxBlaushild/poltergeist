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
export const usePlace = (placeId) => {
    const [place, setPlace] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const { apiClient } = useAPI();
    useEffect(() => {
        const fetchPlaces = () => __awaiter(void 0, void 0, void 0, function* () {
            try {
                const data = yield apiClient.get(`/sonar/google/place/${placeId}`);
                setPlace(data);
            }
            catch (err) {
                setError(err);
            }
            finally {
                setLoading(false);
            }
        });
        fetchPlaces();
    }, [placeId]);
    return {
        place: place !== null && place !== void 0 ? place : null,
        loading,
        error
    };
};
