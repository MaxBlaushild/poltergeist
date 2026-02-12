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
import { useUserLevel } from '@poltergeist/hooks';
import { SideNavTabs } from './SideNavTabs.tsx';
import { UserContextProvider } from '../contexts/UserContext.tsx';
import Profile from './Profile.tsx';

export function Layout() {
  const [isNavOpen, setIsNavOpen] = useState(false);
  const [showProfilePicture, setShowProfilePicture] = useState(false);
  const { user, loading, logout } = useAuth();
  const navigate = useNavigate();
  const { currentUser } = useUserProfiles();
  const { hasCurrentMatch, matchID } = useHasCurrentMatch();
  const { leaveMatch } = useLeaveMatch();
  const { userLevel } = useUserLevel();
  const [username, setUsername] = useState<string | null>(null);

  const toggleNav = () => {
    setIsNavOpen(!isNavOpen);
    setShowProfilePicture(false);
    setUsername(null);
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
                onClick={() => navigate('/?from=/single-player')}
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
        className={`Layout__sideNav fixed top-0 right-0 w-full h-full transform ${isNavOpen ? 'translate-x-0' : 'translate-x-full'} transition-transform duration-300 ease-in-out z-60 flex flex-col overflow-y-auto`}
      >
        <button onClick={toggleNav}>
          <XMarkIcon className="h-8 w-8 mt-3 ml-3" />
        </button>

        <UserContextProvider>
            <Profile showBackButton={true} isOwnProfile={true} />
          <SideNavTabs />
        </UserContextProvider>

        {hasCurrentMatch ? (
          <div className="m-4 mb-6">
            <Button
              title="Leave match"
              onClick={async () => {
                await leaveMatch(matchID);
                navigate('/single-player');
                window.location.reload();
              }}
            />
          </div>
        ) : null}
      </div>
    </div>
  );
}
