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
import { getCityFromCoordinates } from '@poltergeist/utils';
export const useCityName = (latitude, longitude) => {
    const [cityName, setCityName] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    useEffect(() => {
        const fetchCityName = () => __awaiter(void 0, void 0, void 0, function* () {
            try {
                const city = yield getCityFromCoordinates(latitude, longitude);
                setCityName(city);
            }
            catch (error) {
                setError(error);
            }
            finally {
                setLoading(false);
            }
        });
        fetchCityName();
    }, [latitude, longitude]);
    return {
        cityName,
        loading,
        error,
    };
};
