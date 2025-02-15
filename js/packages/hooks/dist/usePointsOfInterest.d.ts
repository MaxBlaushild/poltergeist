import { PointOfInterest } from '@poltergeist/types';
export interface UsePointsOfInterestResult {
    pointsOfInterest: PointOfInterest[] | null;
    loading: boolean;
    error: Error | null;
}
export declare const usePointsOfInterest: () => UsePointsOfInterestResult;
