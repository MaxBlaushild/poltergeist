import { GooglePlace } from '@poltergeist/types';
interface UsePlaceResult {
    place: GooglePlace | null;
    loading: boolean;
    error: Error | null;
}
export declare const usePlace: (placeId: string) => UsePlaceResult;
export {};
