import React, { useState } from 'react';
import './Layout.css';
import { Outlet, useNavigate } from 'react-router-dom';
import { ChevronDownIcon, PencilIcon, XMarkIcon } from '@heroicons/react/20/solid';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { Button, ButtonSize } from './shared/Button.tsx';
import Divider from './shared/Divider.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { Scoreboard } from './Scoreboard.tsx';
import useImageGenerations from '../hooks/useImageGenerations.ts';

const ProfilePictureModal = ({ onExit }: { onExit: () => void }) => {
  const { imageGenerations } = useImageGenerations();
  const { currentUser, refreshUser } = useUserProfiles();
  const { apiClient } = useAPI();
  const [selectedProfilePicture, setSelectedProfilePicture] = useState<string>(currentUser?.profilePictureUrl || '/blank-avatar.webp');

  const setProfilePicture = async () => {
    await apiClient.post('/sonar/profilePicture', {
      profilePictureUrl: selectedProfilePicture,
    });
    refreshUser();
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
      className="fixed inset-0 bg-black z-[100] flex flex-col items-center justify-center"
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
  );
};

export function Layout() {
  const [isNavOpen, setIsNavOpen] = useState(false);
  const [showProfilePicture, setShowProfilePicture] = useState(false);
  const { user, loading, logout } = useAuth();
  const navigate = useNavigate();
  const { currentUser } = useUserProfiles();
  const { match, isLeavingMatch, leaveMatch, leaveMatchError } =
    useMatchContext();

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
                <div
                  className="flex justify-center items-center w-10 h-10 rounded-full overflow-hidden"
                  onClick={toggleNav}
                >
                  <img
                    src={
                      currentUser?.profilePictureUrl ||
                      '/blank-avatar.webp'
                    }
                    alt="Profile Icon"
                    className="object-cover w-full h-full"
                  />
                </div>
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
        {match ? (
          <div className="m-4 mb-6">
            <Button
              title="Leave match"
              onClick={() => {
                leaveMatch();
                navigate('/dashboard');
                setIsNavOpen(false);
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
