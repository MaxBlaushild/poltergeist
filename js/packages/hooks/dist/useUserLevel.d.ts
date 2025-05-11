import { UserLevel } from '@poltergeist/types';
export interface UseUserLevelResult {
    userLevel: UserLevel | null;
    loading: boolean;
    error: Error | null;
}
export declare const useUserLevel: () => UseUserLevelResult;
