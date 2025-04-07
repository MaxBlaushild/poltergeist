import React, { useState } from 'react';
import { AdjustmentsHorizontalIcon } from '@heroicons/react/24/solid';
import { Modal } from './shared/Modal.tsx';
import { useTagContext } from '../contexts/TagContext.tsx';

export const TagFilter = () => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <>
      <div
        className="absolute top-20 right-16 z-10 bg-white rounded-lg p-2 border-2 border-black opacity-80"
        onClick={() => setIsOpen(!isOpen)}
    >
      <AdjustmentsHorizontalIcon className="w-6 h-6" />
    </div>
    {isOpen && <TagFilterModal onClose={() => setIsOpen(!isOpen)} />}
    </>
  );
};

const TagFilterModal = ({ onClose }: { onClose: () => void }) => {
  const { tagGroups, selectedTags, setSelectedTags } = useTagContext();
  const [openGroups, setOpenGroups] = useState<{[key: string]: boolean}>({});

  const toggleGroup = (groupId: string) => {
    setOpenGroups(prev => ({
      ...prev,
      [groupId]: !prev[groupId]
    }));
  };

  const isGroupChecked = (tagGroup: any) => {
    return tagGroup.tags.every(tag => selectedTags.some(t => t.id === tag.id));
  };

  const toggleGroupTags = (tagGroup: any, checked: boolean) => {
    if (checked) {
      // Add all tags from group that aren't already selected
      const tagsToAdd = tagGroup.tags.filter(tag => 
        !selectedTags.some(t => t.id === tag.id)
      );
      setSelectedTags([...selectedTags, ...tagsToAdd]);
    } else {
      // Remove all tags from this group
      setSelectedTags(selectedTags.filter(tag => 
        !tagGroup.tags.some(t => t.id === tag.id)
      ));
    }
  };

  const handleTagChange = (tag: any, checked: boolean) => {
    if (checked) {
      setSelectedTags([...selectedTags, tag]);
    } else {
      setSelectedTags(selectedTags.filter(t => t.id !== tag.id));
    }
  };

  return (
    <Modal>
      <h1 className="font-bold">What are you in the mood for?</h1>
      <div className="flex flex-col gap-2 w-full mt-4">
        {tagGroups.map((tagGroup) => (
          <div key={tagGroup.id} className="border rounded p-2 w-full">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-2">
                <input 
                  type="checkbox" 
                  id={`group-${tagGroup.id}`}
                  checked={isGroupChecked(tagGroup)}
                  onChange={(e) => toggleGroupTags(tagGroup, e.target.checked)}
                />
                <span 
                  className="cursor-pointer"
                  onClick={() => toggleGroup(tagGroup.id)}
                >
                  {tagGroup.name.charAt(0).toUpperCase() + tagGroup.name.slice(1).toLowerCase()}
                </span>
              </div>
              <span 
                className="text-xl cursor-pointer" 
                onClick={() => toggleGroup(tagGroup.id)}
              >
                {openGroups[tagGroup.id] ? '−' : '+'}
              </span>
            </div>
            {openGroups[tagGroup.id] && (
              <div className="pl-4 mt-2 flex flex-col gap-1 w-full">
                {tagGroup.tags.map(tag => (
                  <div key={tag.id} className="flex items-center gap-2 w-full">
                    <input 
                      type="checkbox" 
                      id={tag.id} 
                      checked={selectedTags.some(t => t.id === tag.id)}
                      onChange={(e) => handleTagChange(tag, e.target.checked)}
                    />
                    <label htmlFor={tag.id}>{tag.name.charAt(0).toUpperCase() + tag.name.slice(1).toLowerCase()}</label>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </Modal>
  );
};
