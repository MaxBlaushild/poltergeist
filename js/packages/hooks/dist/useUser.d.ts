import { User } from '@poltergeist/types';
export interface UseUserResult {
    user: User | null;
    loading: boolean;
    error: Error | null;
}
export declare const useUser: (username: string | null) => UseUserResult;
