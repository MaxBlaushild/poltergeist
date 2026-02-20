import React, { useEffect, useRef, useState, FC } from 'react';
import { useUserContext } from '../contexts/UserContext.tsx';
import { UserIcon, PhoneIcon, IdentificationIcon, UserGroupIcon, ArrowLeftIcon, CameraIcon } from '@heroicons/react/24/outline';
import { useAuth, useAPI, useMediaContext } from '@poltergeist/contexts';
import { useZoneContext } from '@poltergeist/contexts';
import { UserZoneReputation } from '@poltergeist/types';
import { useUserLevel } from '@poltergeist/hooks';
import { User } from '@poltergeist/types';
import { useNavigate } from 'react-router-dom';
import { useParty } from '../contexts/PartyContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

interface ProfileProps {
  isOwnProfile?: boolean;
  showBackButton?: boolean;
  showLevelProgress?: boolean;
  onBack?: () => void;
}

const Profile: FC<ProfileProps> = ({ isOwnProfile = false, showBackButton = false, showLevelProgress = true, onBack }) => {
  const userCtx = useUserContext();
  const { user: contextUser, loading: contextLoading, error: contextError, setUsername } = userCtx || { user: null, loading: false, error: null, setUsername: () => {} } as any;
  const { user: authUser, logout } = useAuth();
  const { currentUser, currentUserLoading, currentUserError, refreshUser } = useUserProfiles();
  const { userLevel } = useUserLevel();
  const navigate = useNavigate();
  const { apiClient } = useAPI();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { zones } = useZoneContext();
  const [isInviting, setIsInviting] = useState(false);
  const [inviteSuccess, setInviteSuccess] = useState(false);
  const { inviteToParty } = useParty();
  const [reputations, setReputations] = useState<UserZoneReputation[]>([]);
  const [isUpdatingProfilePicture, setIsUpdatingProfilePicture] = useState(false);
  const [profilePictureError, setProfilePictureError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const data = await apiClient.getUserReputations();
        if (mounted) setReputations(data || []);
      } catch (e) {
        console.error('Failed to load reputations', e);
      }
    })();
    return () => {
      mounted = false;
    };
  }, [apiClient]);

  // Use current user for own profile, context user for viewing others
  const ownUser = currentUser ?? authUser;
  const user = isOwnProfile ? ownUser : contextUser;
  const loading = isOwnProfile ? currentUserLoading : contextLoading;
  const error = isOwnProfile ? currentUserError : contextError;
  const canEditProfile = !!authUser && !!user && authUser.id === user.id;

  const handleProfilePictureUpload = async (file: File) => {
    if (!user) return;
    setProfilePictureError(null);
    setIsUpdatingProfilePicture(true);
    try {
      const timestamp = new Date().getTime().toString();
      const extension = file.name.split('.').pop()?.toLowerCase() || 'png';
      const presignedUrl = await getPresignedUploadURL(
        'crew-profile-icons',
        `${user.id}-${timestamp}.${extension}`
      );
      if (!presignedUrl) {
        throw new Error('Failed to get upload URL');
      }
      await uploadMedia(presignedUrl, file);
      const imageUrl = presignedUrl.split('?')[0];
      await apiClient.post('/sonar/generateProfilePictureOptions', {
        profilePictureUrl: imageUrl,
      });
      await refreshUser();
    } catch (e) {
      console.error('Failed to update profile picture', e);
      setProfilePictureError('Failed to update profile picture. Please try again.');
    } finally {
      setIsUpdatingProfilePicture(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-16 w-16 border-t-4 border-b-4 border-pink-400"></div>
          <p className="mt-4 text-gray-600 font-pixelify">Loading profile...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="bg-red-50 border-2 border-red-400 rounded-lg p-6 max-w-md">
          <h2 className="text-xl font-bold text-red-800 mb-2">Error Loading Profile</h2>
          <p className="text-red-600">{error.message}</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  const handleBack = () => {
    if (onBack) {
      onBack();
    } else if (setUsername) {
      setUsername(null);
    }
  };

  const handleInviteToParty = async () => {
    if (!user || isOwnProfile) return;
    
    setIsInviting(true);
    try {
      await inviteToParty(user);
      setInviteSuccess(true);
      setTimeout(() => setInviteSuccess(false), 3000);
    } catch (error) {
      console.error('Failed to send party invite:', error);
      alert('Failed to send party invite. Please try again.');
    } finally {
      setIsInviting(false);
    }
  };

  const formatPhoneNumber = (phoneNumber: string) => {
    const cleaned = phoneNumber.replace(/\D/g, '');
    const match = cleaned.match(/^(\d{1})(\d{3})(\d{3})(\d{4})$/);
    if (match) {
      return `+${match[1]} (${match[2]}) ${match[3]}-${match[4]}`;
    }
    return phoneNumber;
  };

  const earnedReputations = reputations.filter((rep) => rep.totalReputation > 0);

  return (
    <div className="bg-gradient-to-br from-pink-50 via-blue-50 to-purple-50 py-4 px-4 h-screen">
      <div className="max-w-4xl mx-auto">
        {/* Back Button */}
        {showBackButton && (
          <button
            onClick={handleBack}
            className="flex items-center gap-2 mb-4 px-4 py-2 text-gray-700 hover:text-gray-900 transition-colors"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            <span className="font-medium">Back to menu</span>
          </button>
        )}

        {/* Header Card */}
        <div className="bg-white rounded-2xl shadow-lg border-3 border-gray-900 overflow-hidden mb-6">
          {/* Banner */}
          <div className="h-32 bg-gradient-to-r from-pink-400 via-purple-400 to-blue-400"></div>
          
          {/* Profile Picture and Basic Info */}
          <div className="relative px-6 pb-6">
            <div className="flex flex-col sm:flex-row sm:items-end gap-4">
              {/* Profile Picture */}
              <div className="-mt-16 relative w-32 h-32">
                <div className="w-32 h-32 rounded-full border-4 border-white shadow-xl overflow-hidden bg-gray-200">
                  {user.profilePictureUrl ? (
                    <img
                      src={user.profilePictureUrl}
                      alt={user.username}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-pink-300 to-blue-300">
                      <UserIcon className="h-16 w-16 text-white" />
                    </div>
                  )}
                </div>
                <div className={`absolute bottom-0 right-0 w-6 h-6 rounded-full border-4 border-white ${user.isActive ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                {canEditProfile ? (
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={isUpdatingProfilePicture}
                    className={`absolute -bottom-10 left-1/2 -translate-x-1/2 flex items-center gap-2 px-3 py-1.5 text-xs font-semibold rounded-full border-2 border-gray-900 shadow-md transition-colors ${
                      isUpdatingProfilePicture
                        ? 'bg-gray-200 text-gray-500 cursor-wait'
                        : 'bg-white text-gray-900 hover:bg-gray-100'
                    }`}
                  >
                    <CameraIcon className="h-4 w-4" />
                    {isUpdatingProfilePicture
                      ? 'Updating...'
                      : user.profilePictureUrl
                      ? 'Change Photo'
                      : 'Add Photo'}
                  </button>
                ) : null}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(event) => {
                    const selectedFile = event.target.files?.[0];
                    if (selectedFile) {
                      handleProfilePictureUpload(selectedFile);
                    }
                    event.target.value = '';
                  }}
                />
              </div>

              {/* Username and Level */}
              <div className="flex-1 sm:mb-4">
                <h1 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-1">
                  {user.username}
                </h1>
                {user.name && (
                  <p className="text-lg text-gray-600">{user.name}</p>
                )}
                {profilePictureError ? (
                  <p className="mt-2 text-sm text-red-600">{profilePictureError}</p>
                ) : null}
              </div>
            </div>

            {/* Experience Bar */}
            {showLevelProgress && userLevel && (
              <div className="mt-6">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-bold">Level {userLevel.level}</span>
                  <span className="text-sm text-gray-600">
                    {userLevel.experiencePointsOnLevel} / {userLevel.experienceToNextLevel} XP
                  </span>
                </div>
                <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div 
                    className="h-full bg-blue-500 transition-all duration-300"
                    style={{
                      width: `${userLevel ? (userLevel.experiencePointsOnLevel / userLevel.experienceToNextLevel) * 100 : 0}%`
                    }}
                  />
                </div>
              </div>
            )}

            {/* Gold Display */}
            <div className="mt-4">
              <div className="bg-white rounded-lg p-3 shadow-md border border-amber-400">
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 rounded-lg border border-amber-400 flex items-center justify-center">
                    <span className="text-sm font-bold text-amber-600">GOLD</span>
                  </div>
                  <span className="text-lg font-bold text-gray-900">{user ? user.gold ? user.gold : '0' : '0'}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Details Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Zone Reputation */}
          <div className="bg-white rounded-2xl shadow-lg border-3 border-gray-900 overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-xl font-bold text-gray-900">Zone Reputation</h2>
            </div>
            <div className="p-6">
              {earnedReputations.length === 0 ? (
                <p className="text-gray-600">No reputation yet.</p>
              ) : (
                <ul className="divide-y divide-gray-100">
                  {earnedReputations.map((rep) => {
                    const zone = zones.find(z => z.id === rep.zoneId);
                    const zoneName = zone?.name || rep.zoneId.substring(0, 8);
                    return (
                      <li key={rep.zoneId} className="py-3 flex items-center justify-between">
                        <span className="font-medium text-gray-900">{zoneName}</span>
                        <span className="text-gray-700">Level {rep.level}</span>
                      </li>
                    );
                  })}
                </ul>
              )}
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="mt-6 flex justify-center">
          {isOwnProfile ? (
            <button
              onClick={() => {
                if (window.confirm('Are you sure you want to log out?')) {
                  logout();
                  navigate('/');
                }
              }}
              className="px-8 py-3 bg-red-500 hover:bg-red-600 text-white font-bold rounded-lg shadow-md transition-colors border-2 border-gray-900"
            >
              Log Out
            </button>
          ) : (
            <button
              onClick={handleInviteToParty}
              disabled={isInviting || inviteSuccess}
              className={`px-8 py-3 font-bold rounded-lg shadow-md transition-colors border-2 border-gray-900 ${
                inviteSuccess
                  ? 'bg-green-500 text-white cursor-default'
                  : isInviting
                  ? 'bg-gray-400 text-white cursor-wait'
                  : 'bg-purple-500 hover:bg-purple-600 text-white'
              }`}
            >
              {inviteSuccess ? (
                <span className="flex items-center gap-2">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  Invite Sent!
                </span>
              ) : isInviting ? (
                'Sending...'
              ) : (
                'Invite to Party'
              )}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default Profile;
