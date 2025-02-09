import { Place } from '@poltergeist/types';
interface UsePlacesResult {
    places: Place[];
    loading: boolean;
    error: Error | null;
}
export declare const usePlaces: (address: string) => UsePlacesResult;
export {};
