import './Home.css';
import './shared/Button.css';
import React, { useCallback, useState } from 'react';
import { Button, ButtonSize } from './shared/Button.tsx';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { Logister } from '@poltergeist/components';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useMediaContext } from '../contexts/MediaContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { User } from '@poltergeist/types';

type SubmitProfilePictureResponse = {
  message: string;
};

export function Home() {
  const navigate = useNavigate();
  const {
    logister,
    getVerificationCode,
    isWaitingForVerificationCode,
    isRegister,
  } = useAuth();
  
  const [isLogistering, setIsLogistering] = useState<boolean>(false);
  const [error, setError] = useState<string | undefined>(undefined);
  const [shouldSetProfilePicture, setShouldSetProfilePicture] = useState<boolean>(false);
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { search } = useLocation();
  const queryParams = new URLSearchParams(search);
  const from = queryParams.get('from');
  const unescapedFrom = from ? decodeURIComponent(from) : undefined;
  const { apiClient } = useAPI();
  

  const uploadProfilePicture = useCallback(async (image?: File | undefined): Promise<SubmitProfilePictureResponse | undefined> => {
    const user = await apiClient.get<User>('/sonar/whoami');
    let timestamp = new Date().getTime().toString();
    const getExtension = (filename: string): string => {
      return filename.split('.').pop()?.toLowerCase() || '';
    };
    const extension = image ? getExtension(image.name) : '';
    var imageUrl = '';

    if (image) {
      const presignedUrl = await getPresignedUploadURL("crew-profile-icons", `${user?.id}-${timestamp}.${extension}`);
      if (!presignedUrl) return;
      await uploadMedia(presignedUrl, image);
      imageUrl = presignedUrl.split("?")[0];
    }

    apiClient.post<SubmitProfilePictureResponse>(`/sonar/generateProfilePictureOptions`, {
      profilePictureUrl: imageUrl,
    });

    navigate(unescapedFrom || '/dashboard');

  }, [navigate]);

  return (
    <div className="Home__background">
      {!isLogistering && !unescapedFrom ? (
        <div className="Dashboard__gameOpenModal">
          <h2 className="text-2xl font-bold">
            Find your crew and set your sights on adventure
          </h2>
          <div>
            <Button
              title="Get started"
              buttonSize={ButtonSize.LARGE}
              onClick={() => {
                setIsLogistering(true);
              }}
            />
          </div>
        </div>
      ) : null}
      {isLogistering || unescapedFrom ? (
        <Modal size={ModalSize.FULLSCREEN}>
          {!shouldSetProfilePicture ? (
            <div className="flex flex-col items-center gap-4 w-full">
              <h2 className="Login__title">Sign in or sign up</h2>
              <Logister
              logister={async (one, two, three, isRegister) => {
                try {
                  await logister(one, two, three);
                  if (isRegister) {
                    setShouldSetProfilePicture(true);
                  } else {
                    navigate(unescapedFrom || '/dashboard');
                  }
                } catch (e) {
                  setError('Something went wrong. Please try again later');
                }
              }}
              error={error}
              getVerificationCode={getVerificationCode}
              isRegister={isRegister}
              isWaitingOnVerificationCode={isWaitingForVerificationCode}
              />
            </div>
          ) : (
            <div className="flex flex-col items-center gap-4 w-full">
              <h2 className="Login__title">Take a selfie and get piratified</h2>
              <input
              id="file"
              type="file"
              className="w-full"
              onChange={(e) => uploadProfilePicture(e.target.files?.[0])}
            />
            </div>
          )}
        </Modal>
      ) : null}
    </div>
  );
}
