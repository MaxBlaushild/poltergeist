import './Home.css';
import './shared/Button.css';
import React, { useState } from 'react';
import { Button, ButtonSize } from './shared/Button.tsx';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '@poltergeist/contexts';
import { Logister } from '@poltergeist/components';
import { Modal, ModalSize } from './shared/Modal.tsx';

export function Home() {
  const navigate = useNavigate();
  const {
    user,
    logister,
    getVerificationCode,
    isWaitingForVerificationCode,
    isRegister,
  } = useAuth();
  const [isLogistering, setIsLogistering] = useState<boolean>(false);
  const [error, setError] = useState<string | undefined>(undefined);
  const { search } = useLocation();
  const queryParams = new URLSearchParams(search);
  const from = queryParams.get('from');
  const unescapedFrom = from ? decodeURIComponent(from) : undefined;

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
          <div className="flex flex-col items-center gap-4 w-full">
            <h2 className="Login__title">Sign in or sign up</h2>
            <Logister
              logister={async (one, two, three) => {
                try {
                  await logister(one, two, three);
                  navigate(unescapedFrom || '/dashboard');
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
        </Modal>
      ) : null}
    </div>
  );
}
