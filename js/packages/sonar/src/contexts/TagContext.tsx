import React, { createContext, useState, useEffect, useContext } from 'react';
import { Tag, TagGroup } from '@poltergeist/types';
import { useAPI } from '@poltergeist/contexts';

type TagContextType = {
  tagGroups: TagGroup[];
  selectedTags: Tag[];
  setSelectedTags: (tags: Tag[]) => void;
};

export const TagContext = createContext<TagContextType>({
  tagGroups: [],
  selectedTags: [],
  setSelectedTags: () => {},
});

export const TagProvider = ({ children }: { children: React.ReactNode }) => {
  const { apiClient } = useAPI();
  const [tags, setTags] = useState<Tag[]>([]);
  const [tagGroups, setTagGroups] = useState<TagGroup[]>([]);
  const [selectedTags, setSelectedTags] = useState<Tag[]>([]);

  const fetchTagGroups = async () => {
    const response = await apiClient.get<TagGroup[]>('/sonar/tagGroups');
    setTagGroups(response);
    setSelectedTags(response.flatMap(group => group.tags));
  };

  useEffect(() => {
    fetchTagGroups();
  }, []);

  return (
    <TagContext.Provider value={{ tagGroups, selectedTags, setSelectedTags }}>
      {children}
    </TagContext.Provider>
  );
};

export const useTagContext = () => {
  return useContext(TagContext);
};
