import { Candidate } from '@poltergeist/types';
interface UsePlacesResult {
    candidates: Candidate[];
    loading: boolean;
    error: Error | null;
}
export declare const useCandidates: (query: string) => UsePlacesResult;
export {};
