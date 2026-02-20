import React from 'react';
import Profile from './Profile.tsx';
import { useUserLevel } from '@poltergeist/hooks';

export const Character: React.FC = () => {
  const { userLevel } = useUserLevel();
  const progressPercent = userLevel && userLevel.experienceToNextLevel > 0
    ? Math.min(
        100,
        (userLevel.experiencePointsOnLevel / userLevel.experienceToNextLevel) * 100
      )
    : 0;

  return (
    <div className="space-y-4">
      {userLevel ? (
        <div className="rounded-xl border-2 border-gray-900 bg-white p-4 shadow-md">
          <div className="flex items-center justify-between">
            <span className="text-sm font-bold text-gray-900">Level {userLevel.level}</span>
            <span className="text-xs text-gray-600">
              {userLevel.experiencePointsOnLevel} / {userLevel.experienceToNextLevel} XP
            </span>
          </div>
          <div className="mt-2 h-2 w-full rounded-full bg-gray-200 overflow-hidden">
            <div
              className="h-full bg-blue-500 transition-all duration-300"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
        </div>
      ) : null}
      <Profile isOwnProfile={true} showLevelProgress={false} />
    </div>
  );
};
