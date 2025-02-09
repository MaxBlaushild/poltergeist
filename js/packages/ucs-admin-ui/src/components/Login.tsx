import { Logister } from '@poltergeist/components';
import { useAuth } from '@poltergeist/contexts';
import React, { useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

export const Login = () => {
  const {
    logister,
    getVerificationCode,
    isWaitingForVerificationCode,
    isRegister,
  } = useAuth();
  const [error, setError] = useState<string | undefined>(undefined);
  const navigate = useNavigate();
  const { search } = useLocation();
  const queryParams = new URLSearchParams(search);
  const from = queryParams.get('from');
  const unescapedFrom = from ? decodeURIComponent(from) : undefined;

  return (
    <div className="flex flex-col items-center gap-4 w-full">
      <h2 className="Login__title">Sign in</h2>
      <Logister
        logister={async (one, two, three, isRegister) => {
          try {
            await logister(one, two, three);
            if (isRegister) {
              setError('Something went wrong. Please try again later');
            } else {
              navigate(unescapedFrom || '/');
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
  );
};
