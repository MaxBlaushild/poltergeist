import React, { createContext, useState, useEffect, useContext } from 'react';
import { Tag, TagGroup } from '@poltergeist/types';
import { useAPI, useAuth } from '@poltergeist/contexts';

type TagContextType = {
  tagGroups: TagGroup[];
  selectedTags: Tag[];
  setSelectedTags: (tags: Tag[]) => void;
  createTagGroup: (tagGroup: TagGroup) => void;
  moveTagToTagGroup: (tagID: string, tagGroupID: string) => void;
};

export const TagContext = createContext<TagContextType>({
  tagGroups: [],
  selectedTags: [],
  setSelectedTags: () => {},
  createTagGroup: () => {},
  moveTagToTagGroup: () => {},
});

export const TagProvider = ({ children }: { children: React.ReactNode }) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [tags, setTags] = useState<Tag[]>([]);
  const [tagGroups, setTagGroups] = useState<TagGroup[]>([]);
  const [selectedTags, setSelectedTags] = useState<Tag[]>([]);

  const fetchTagGroups = async () => {
    const response = await apiClient.get<TagGroup[]>('/sonar/tagGroups');
    setTagGroups(response);
  };

  const createTagGroup = async (tagGroup: TagGroup) => {
    const response = await apiClient.post<TagGroup>('/sonar/tagGroups', tagGroup);
    setTagGroups([...tagGroups, response]);
  };

  const moveTagToTagGroup = async (tagID: string, tagGroupID: string) => {
    await apiClient.post(`/sonar/tags/move`, { tagID, tagGroupID });
    fetchTagGroups();
  };

  useEffect(() => {
    if (!user) {
      setTagGroups([]);
      return;
    }
    fetchTagGroups();
  }, [user]);

  return (
    <TagContext.Provider value={{ tagGroups, selectedTags, setSelectedTags, createTagGroup, moveTagToTagGroup }}>
      {children}
    </TagContext.Provider>
  );
};

export const useTagContext = () => {
  return useContext(TagContext);
};
