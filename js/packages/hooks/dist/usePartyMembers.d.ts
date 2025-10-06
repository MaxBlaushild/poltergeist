import { User } from '@poltergeist/types';
interface UseCityNameResult {
    partyMembers: User[];
    loading: boolean;
    error: Error | null;
}
export declare const usePartyMembers: () => UseCityNameResult;
export declare const joinParty: (inviterID: string) => Promise<void>;
export {};
