import React, { useEffect, useState } from 'react';
import './Layout.css';
import { Outlet, useNavigate } from 'react-router-dom';
import { ChevronDownIcon, PencilIcon, XMarkIcon } from '@heroicons/react/20/solid';
import { useAPI, useAuth, useMediaContext } from '@poltergeist/contexts';
import { Button, ButtonSize } from './shared/Button.tsx';
import Divider from './shared/Divider.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { Scoreboard } from './Scoreboard.tsx';
import useImageGenerations from '../hooks/useImageGenerations.ts';
import { ImageBadge } from './shared/ImageBadge.tsx';
import { Modal, ModalSize } from './shared/Modal.tsx';
import useHasCurrentMatch from '../hooks/useHasCurrentMatch.ts';
import useLeaveMatch from '../hooks/useLeaveMatch.ts';
const ProfilePictureModal = ({ onExit }: { onExit: () => void }) => {
  const { imageGenerations } = useImageGenerations();
  const { getPresignedUploadURL, uploadMedia } = useMediaContext();
  const { currentUser, refreshUser } = useUserProfiles();
  const { apiClient } = useAPI();
  const [selectedProfilePicture, setSelectedProfilePicture] = useState<string>(currentUser?.profilePictureUrl || '/blank-avatar.webp');
  const [showGenderInput, setShowGenderInput] = useState(false);
  const [selectedGender, setSelectedGender] = useState('male');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [toastText, setToastText] = useState<string | null>(null);

  const setProfilePicture = async () => {
    await apiClient.post('/sonar/profilePicture', {
      profilePictureUrl: selectedProfilePicture,
    });
    refreshUser();
  };

  useEffect(() => {
    if (toastText) {
      setTimeout(() => {
        setToastText(null);
      }, 1500);
    }
  }, [toastText]);

  const generateProfilePictureOptions = async () => {
    if (!selectedFile) {
      return;
    }
    const presignedUrl = await getPresignedUploadURL('crew-profile-icons', `${currentUser?.id}-${new Date().getTime().toString()}.${selectedFile.name.split('.').pop()?.toLowerCase() || ''}`);
    if (!presignedUrl) {
      return;
    }
    await uploadMedia(presignedUrl, selectedFile);
    await apiClient.post('/sonar/generateProfilePictureOptions', {
      profilePictureUrl: presignedUrl.split('?')[0],
      gender: selectedGender,
    });
    setShowGenderInput(false);
    setToastText('Successfully started generating profile picture options! Check back in a few minutes.');
  };

  const profilePictures: string[] = [];
  imageGenerations?.forEach((gen) => {
    if (gen.optionOne) {
      profilePictures.push(gen.optionOne);
    }
    if (gen.optionTwo) {
      profilePictures.push(gen.optionTwo);
    }
    if (gen.optionThree) {
      profilePictures.push(gen.optionThree);
    }
    if (gen.optionFour) {
      profilePictures.push(gen.optionFour);
    }
  });

  return (
    <div 
      className="fixed inset-0 bg-black z-[100] flex flex-col items-center p-4 pt-10"
      onClick={() => {
        setProfilePicture();
        onExit();
      }}
    >
      <img
        src={selectedProfilePicture}
        alt="Profile Picture"
        className="max-w-[90%] max-h-[80%] object-contain mb-4"
      />
      <div className="flex gap-2 overflow-x-auto p-2">
        <div className="grid grid-cols-4 gap-2">
          {profilePictures.map((url, index) => (
            <div 
              key={index} 
              className={`w-12 h-12 rounded-full overflow-hidden flex-shrink-0 ${
                url === selectedProfilePicture ? 'border-2 border-[#fa9eb5]' : ''
              }`}
            >
              <img
                src={url}
                alt={`Profile Picture Option ${index + 1}`}
                className="w-full h-full object-cover"
                onClick={(e) => {
                  e.stopPropagation();
                  setSelectedProfilePicture(url);
                }}
              />
            </div>
          ))}
        </div>
      </div>
      <div className="flex flex-col items-center gap- mt-4 w-full" onClick={(e) => e.stopPropagation()}>
        <Button
          buttonSize={ButtonSize.SMALL}
          title={showGenderInput ? "Hide Options" : "Generate New Options"}
          onClick={() => setShowGenderInput(!showGenderInput)}
        />
        {showGenderInput && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[200]" onClick={() => setShowGenderInput(false)}>
            <div className="flex flex-col items-center gap-4 bg-gray-800 p-6 rounded-lg w-[90%] max-w-md" onClick={e => e.stopPropagation()}>
              <div className="flex flex-col gap-1 w-full">
                <label className="text-white text-sm">Gender</label>
                <input
                  type="text"
                  value={selectedGender}
                  onChange={(e) => setSelectedGender(e.target.value)}
                  placeholder="Enter gender"
                  className="p-2 rounded bg-gray-700 text-white w-full"
                />
              </div>
              <div className="flex justify-center w-full">
                <input
                  type="file"
                  onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
                  className="text-white w-full"
                />
              </div>
              <Button
                buttonSize={ButtonSize.SMALL}
                title="Generate"
                onClick={generateProfilePictureOptions}
              />
            </div>
          </div>
        )}
      </div>
      {toastText ? <Modal size={ModalSize.TOAST}>{toastText}</Modal> : null}
    </div>
  );
};

export function Layout() {
  const [isNavOpen, setIsNavOpen] = useState(false);
  const [showProfilePicture, setShowProfilePicture] = useState(false);
  const { user, loading, logout } = useAuth();
  const navigate = useNavigate();
  const { currentUser } = useUserProfiles();
  const { hasCurrentMatch, matchID } = useHasCurrentMatch();
  const { leaveMatch } = useLeaveMatch();
  const toggleNav = () => {
    setIsNavOpen(!isNavOpen);
    setShowProfilePicture(false);
  };

  return (
    <div className="Layout__background">
      {!isNavOpen ? (
        <div
          className={`Layout__header fixed top-0 left-0 w-full py-4 z-50 ${isNavOpen ? 'hidden' : ''}`}
        >
          <div
            className="flex flex-row justify-between gap-4 cursor-pointer"
            onClick={() => {
              navigate('/');
              setIsNavOpen(false);
            }}
          >
            <img
              src="/pirate-ship.png"
              alt="Pirate Ship"
              className="Layout__icon"
            />
            <h1 className="font-bold text-3xl">crew</h1>
          </div>
          <div>
            {user ? (
              <div className="relative">
                <ImageBadge
                  imageUrl={currentUser?.profilePictureUrl || '/blank-avatar.webp'}
                  onClick={toggleNav}
                />
                <div className="absolute bottom-0 right-0 flex justify-center items-center w-4 h-4 rounded-full overflow-hidden Header__circleThing">
                  <ChevronDownIcon className="w-4 h-4" />
                </div>
              </div>
            ) : !loading ? (
              <Button
                buttonSize={ButtonSize.SMALL}
                title="Log in"
                onClick={() => navigate('/?from=/dashboard')}
              />
            ) : null}
          </div>
        </div>
      ) : null}
      <div className="Layout__content">
        <div style={{ display: isNavOpen ? 'none' : 'block' }}>
          <Outlet />
        </div>
      </div>
      <div
        className={`Layout__sideNav fixed top-0 right-0 w-full h-full transform ${isNavOpen ? 'translate-x-0' : 'translate-x-full'} transition-transform duration-300 ease-in-out z-60`}
      >
        <button onClick={toggleNav}>
          <XMarkIcon className="h-8 w-8 mt-3 ml-3" />
        </button>
        <div className="flex items-center justify-start p-4 gap-4">
          <div 
            className="flex justify-center items-center w-16 h-16 rounded-full overflow-hidden cursor-pointer"
            onClick={() => setShowProfilePicture(true)}
          >
            <img
              src={
                currentUser?.profilePictureUrl || '/blank-avatar.webp'
              }
              alt="Profile Icon"
              className="object-cover w-full h-full"
            />
          </div>
          <h2 className="text-center flex-1 text-xl font-bold text-left">
            {user?.name}
          </h2>
        </div>
        {showProfilePicture && (
          <ProfilePictureModal onExit={() => setShowProfilePicture(false)} />
        )}
        {hasCurrentMatch ? (
          <div className="m-4 mb-6">
            <Button
              title="Leave match"
              onClick={async () => {
                await leaveMatch(matchID);
                navigate('/dashboard');
                window.location.reload();
              }}
            />
          </div>
        ) : null}
        <Divider />
        <div className="m-2 m-4">
          <p
            className="cursor-pointer text-center"
            onClick={() => {
              logout();
              setIsNavOpen(false);
              navigate('/');
            }}
          >
            Log out
          </p>
        </div>
      </div>
    </div>
  );
}
