import React from 'react';
import { useTagContext } from '@poltergeist/contexts';

export const Tags = () => {
  const { tagGroups, moveTagToTagGroup, createTagGroup } = useTagContext();

  const handleDragStart = (e: React.DragEvent, tagId: string) => {
    e.dataTransfer.setData('tagId', tagId);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = async (e: React.DragEvent, targetGroupId: string) => {
    e.preventDefault();
    const tagId = e.dataTransfer.getData('tagId');
    await moveTagToTagGroup(tagId, targetGroupId);
  };

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-6">Tag Groups</h1>
      <div className="space-y-8">
        {tagGroups.map(group => (
          <div 
            key={group.id} 
            className="bg-white rounded-lg shadow-md p-6"
            onDragOver={handleDragOver}
            onDrop={(e) => handleDrop(e, group.id)}
          >
            <h2 className="text-xl font-semibold mb-4">{group.name}</h2>
            <div className="flex flex-wrap gap-2">
              {group.tags.map(tag => (
                <span 
                  key={tag.id}
                  draggable
                  onDragStart={(e) => handleDragStart(e, tag.id)}
                  className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm cursor-move hover:bg-blue-200 transition-colors"
                >
                  {tag.name}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
