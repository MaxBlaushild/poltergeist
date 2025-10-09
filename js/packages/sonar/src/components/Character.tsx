import { useUserLevel } from '@poltergeist/hooks';
import React from 'react';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useAuth } from '@poltergeist/contexts';
import { useNavigate } from 'react-router-dom';

export const Character: React.FC = () => {
  const { userLevel } = useUserLevel();
  const { currentUser } = useUserProfiles();
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  return <div>
    <div className="flex items-center justify-start p-4 gap-4">
      <div 
        className="flex justify-center items-center w-16 h-16 rounded-full overflow-hidden cursor-pointer"
      >
        <img
          src={
            currentUser?.profilePictureUrl || '/blank-avatar.webp'
          }
          alt="Profile Icon"
          className="object-cover w-full h-full"
        />
      </div>
      <div className="flex-1">
        <h2 className="text-xl font-bold">
          {user?.username}
        </h2>
        <p
          className="text-sm text-gray-600 cursor-pointer hover:text-gray-800 mt-1"
          onClick={() => {
            logout();
            navigate('/');
          }}
        >
          Log out
        </p>
      </div>
    </div>
    <div className="px-4 mb-6">
      <div className="flex items-center justify-between mb-2">
        <span className="font-bold">Level {userLevel?.level}</span>
        <span className="text-sm text-gray-600">
          {userLevel?.experiencePointsOnLevel} / {userLevel?.experienceToNextLevel} XP
        </span>
      </div>
      <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
        <div 
          className="h-full bg-blue-500 transition-all duration-300"
          style={{
            width: `${ userLevel ? (userLevel.experiencePointsOnLevel / userLevel.experienceToNextLevel) * 100 : 0}%`
          }}
        />
      </div>
    </div>
  </div>;
};