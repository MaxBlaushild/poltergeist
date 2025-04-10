import React from 'react';
import { Tag, TagGroup } from '@poltergeist/types';
type TagContextType = {
    tagGroups: TagGroup[];
    selectedTags: Tag[];
    setSelectedTags: (tags: Tag[]) => void;
};
export declare const TagContext: React.Context<TagContextType>;
export declare const TagProvider: ({ children }: {
    children: React.ReactNode;
}) => import("react/jsx-runtime").JSX.Element;
export declare const useTagContext: () => TagContextType;
export {};
