import { UserZoneReputation } from '@poltergeist/types';
export interface UseUserZoneReputationResult {
    userZoneReputation: UserZoneReputation | null;
    loading: boolean;
    error: Error | null;
}
export declare const useUserZoneReputation: (zoneId: string | undefined) => UseUserZoneReputationResult;
