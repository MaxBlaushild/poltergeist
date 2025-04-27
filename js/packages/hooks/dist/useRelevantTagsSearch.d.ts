import { Tag } from '@poltergeist/types';
export interface UseRelevantTagsSearchResult {
    relevantTags: Tag[] | null;
    loading: boolean;
    error: Error | null;
}
export declare const useRelevantTagsSearch: (query: string) => UseRelevantTagsSearchResult;
