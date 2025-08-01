import React, { useState, useEffect } from 'react';
import { PencilSquareIcon } from '@heroicons/react/24/solid';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useTagContext } from '@poltergeist/contexts';
import { AdjustmentsHorizontalIcon, XMarkIcon, MagnifyingGlassIcon } from '@heroicons/react/20/solid';
import { useRelevantTagsSearch } from '@poltergeist/hooks';
import { Button } from './shared/Button.tsx';
import { Tag } from '@poltergeist/types';

export const ActivityQuestionnaire = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [successToastOpen, setSuccessToastOpen] = useState(false);

  const openSuccessToast = () => {
    setSuccessToastOpen(true);
    setTimeout(() => {
      setSuccessToastOpen(false);
    }, 3000);
  }

  return (
    <>
      <div
        className="absolute top-44 right-4 z-10 bg-white rounded-lg p-2 border-2 border-black opacity-80"
        onClick={() => setIsOpen(!isOpen)}
    >
      <MagnifyingGlassIcon className="w-6 h-6" />
    </div>
    {isOpen && <ActivityQuestionnaireModal 
      onClose={() => setIsOpen(!isOpen)} 
      onSuccess={openSuccessToast}
    />}
    {successToastOpen && (
      <Modal size={ModalSize.TOAST}>
        Quests were successfully filtered!
      </Modal>
    )}
    </>
  );
};

const ActivityQuestionnaireModal = ({ onClose, onSuccess }: { onClose: () => void, onSuccess: () => void }) => {
  const [openGroups, setOpenGroups] = useState<{[key: string]: boolean}>({});
  const [selectedTags, setSelectedTags] = useState<Tag[]>([]);
  const [searchString, setSearchString] = useState('');
  const [query, setQuery] = useState('');
  const { relevantTags, loading } = useRelevantTagsSearch(query);
  const { setSelectedTags: setSelectedTagsFromContext } = useTagContext();

  useEffect(() => {
    if (relevantTags) {
      setSelectedTags(relevantTags);
    }
  }, [relevantTags]);

  return (
    <Modal size={ModalSize.FORM}>
      <div className="flex justify-between w-full">
        <h1 className="font-bold float-left text-xl">{selectedTags && selectedTags.length > 0 ? 'Do these types of places look right?' : 'What do you want to do?'}</h1>
        <XMarkIcon className='w-6 h-6 float-right' onClick={onClose} />
      </div>
      {loading ? (
        <div className="flex justify-center items-center mt-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black"></div>
        </div>
      ) : (
        <>
          {selectedTags && selectedTags.length === 0 && (
            <div className="flex flex-col gap-4 w-full mt-4">
              <textarea
                placeholder="I want to grocery shop, have a night on the town, etc"
                className="w-full px-4 py-2 border-2 border-gray-300 rounded-lg focus:outline-none focus:border-black"
                onChange={(e) => setSearchString(e.target.value)}
            />
            <Button
              onClick={() => setQuery(searchString)}
              title="Submit"
              />
              <p className="text-xs text-black-500">
                Enter a description of what you want to do. We'll provide you with a list of tags that match. We will filter the quests you see by your selected tags after you confirm them.
              </p>
            </div>
          )}
          {selectedTags && selectedTags.length > 0 && (
            <div className="flex flex-col gap-2 w-full mt-4">
              <div className="flex flex-wrap gap-2 mt-4">
                {selectedTags.map((tag) => (
                  <div
                  key={tag.id}
                  className="flex items-center gap-1 px-3 py-1 bg-gray-200 rounded-full"
                >
                  <span>{tag.name.replace(/_/g, ' ').charAt(0).toUpperCase() + tag.name.replace(/_/g, ' ').slice(1).toLowerCase()}</span>
                  <XMarkIcon
                    className="w-4 h-4 cursor-pointer"
                    onClick={() => setSelectedTags(selectedTags.filter(t => t.id !== tag.id))}
                  />
                </div>
              ))}
            </div>
            <Button
              onClick={() => setSelectedTags([])}
              title="Clear"
            />
            <Button
              onClick={() => {
                setSelectedTagsFromContext(selectedTags);
                onSuccess();
                onClose();
              }}
              title="Submit"
            />
            </div>
          )}
        </>
      )}
    </Modal>
  );
};
