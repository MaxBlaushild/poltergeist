import { PointOfInterestGroup } from '@poltergeist/types';
export interface UsePointOfInterestGroupsResult {
    pointOfInterestGroups: PointOfInterestGroup[] | null;
    loading: boolean;
    error: Error | null;
}
export declare const usePointOfInterestGroups: (type?: number) => UsePointOfInterestGroupsResult;
