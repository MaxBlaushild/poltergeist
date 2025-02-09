import fetch from 'node-fetch';
import { OpenStreetMapLocation } from '@poltergeist/types';

const wait = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

const fetchWithRetry = async (url: string, maxRetries = 3, baseDelay = 1000) => {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fetch(url);
    } catch (error) {
      if (attempt === maxRetries - 1) throw error;
      const delay = baseDelay * Math.pow(2, attempt);
      await wait(delay);
    }
  }
  throw new Error('Max retries reached');
};

export const getCityFromCoordinates = async (
  lat: string,
  lon: string
): Promise<string | null> => {
  const url = `https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lon}&format=json`;

  try {
    const response = await fetchWithRetry(url);
    const data: OpenStreetMapLocation = await response.json();

    if (data && data.address && data.address.city) {
      return data.address.city;
    }
    return null;
  } catch (error) {
    console.error('Error fetching data from Nominatim:', error);
    return null;
  }
}
