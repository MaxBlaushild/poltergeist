import { useState, useEffect } from 'react';
import { getCityFromCoordinates } from '@poltergeist/utils';

interface UseCityNameResult {
  cityName: string | null;
  loading: boolean;
  error: Error | null;
}

export const useCityName = (
  latitude: string,
  longitude: string
): UseCityNameResult => {
  const [cityName, setCityName] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchCityName = async () => {
      try {
        const city = await getCityFromCoordinates(latitude, longitude);
        setCityName(city);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchCityName();
  }, [latitude, longitude]);

  return {
    cityName,
    loading,
    error,
  };
};
