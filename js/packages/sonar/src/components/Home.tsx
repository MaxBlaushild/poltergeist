import './Home.css';
import './shared/Button.css';
import React, { useState } from 'react';
import { Button, ButtonSize } from './shared/Button.tsx';
import { useNavigate } from 'react-router-dom';
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

  return (
    <div className="Home__background">
      <div className="Home__splash">
        <p className="Home__splashContent">
          Navigate the Waves of Connection with Echolocation!
        </p>
      </div>
      <div className="Home__footer">
        <div className="Home__buttonBar">
          <div className="Home__buttonWrapper">
            <Button
              title="Find adventure"
              buttonSize={ButtonSize.LARGE}
              onClick={() => {
                setIsLogistering(true);
              }}
            />
          </div>
        </div>
      </div>
      {isLogistering ? (
        <Modal size={ModalSize.FULLSCREEN}>
          <h2 className="Login__title">Sign in or sign up</h2>
          <Logister
            logister={async (one, two, three) => {
              try {
                await logister(one, two, three);
                navigate('/dashboard');
              } catch (e) {
                setError('Something went wrong. Please try again later');
              }
            }}
            error={error}
            getVerificationCode={getVerificationCode}
            isRegister={isRegister}
            isWaitingOnVerificationCode={isWaitingForVerificationCode}
          />
        </Modal>
      ) : null}
    </div>
  );
}
