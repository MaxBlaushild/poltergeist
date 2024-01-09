/// <reference types="node" />

import React from 'react';
import './Subscribe.css';
import 'react-phone-number-input/style.css';
import PhoneInput from 'react-phone-number-input/input';
import { isValidPhoneNumber } from 'react-phone-number-input';
import cx from 'classnames';
import toast from 'react-hot-toast';
import axios from 'axios';
import { getUserID } from '../util';

type Subscription = {
  subscribed: boolean;
  numFreeQuestions: number;
};

function Subscribe() {
  const [phoneNumber, setValue] = React.useState('');
  const [existingPhoneNumber, setExistingPhoneNumber] = React.useState('');
  const [waitingOnVerificationCode, setWaitingOnVerificationCode] =
    React.useState(false);
  const [code, setCode] = React.useState('');
  const [subscription, setSubscription] = React.useState<Subscription>({
    subscribed: false,
    numFreeQuestions: 0,
  });
  const [hasSubscription, setHasSubscription] = React.useState(false);

  const validPhoneNumber =
    typeof phoneNumber === 'string' && isValidPhoneNumber(phoneNumber);
  const buttonClasses = ['Subscribe__button'];
  const { userId, ephemeralUserId } = getUserID();
  const { subscribed, numFreeQuestions } = subscription;

  const fetchUser = async () => {
    if (userId) {
      const res = await axios.get(
        `${process.env.REACT_APP_API_URL}/authenticator/users?id=${userId}`
      );
      const {
        data: { phoneNumber: existingPhoneNumber },
      } = res;
      setExistingPhoneNumber(existingPhoneNumber);
    }
  };

  const fetchSubscription = async () => {
    if (userId) {
      try {
        const res = await axios.get(
          `${process.env.REACT_APP_API_URL}/trivai/subscriptions/${userId}`
        );
        const { data } = res;
        setSubscription(data || {});
        setHasSubscription(true);
      } catch (e) {}
    }
  };

  React.useEffect(() => {
    fetchUser();
    fetchSubscription();
  }, []);

  if (validPhoneNumber) {
    buttonClasses.push('Button__enabled');
  } else {
    buttonClasses.push('Button__disabled');
  }

  const getVerificationCode = React.useCallback(async () => {
    try {
      await axios.post(
        `${process.env.REACT_APP_API_URL}/authenticator/text/verification-code`,
        { phoneNumber, appName: 'Guess How Many' }
      );
      toast('Verification code sent!');
      setWaitingOnVerificationCode(true);
    } catch (e) {
      toast('Something went wrong!');
    }
  }, [setWaitingOnVerificationCode, toast]);

  const unsubscribe = React.useCallback(async () => {
    try {
      await axios.post(
        `${process.env.REACT_APP_API_URL}/trivai/subscriptions/cancel`,
        { userId }
      );

      localStorage.removeItem('user-id');
      window.location.reload();
    } catch (e) {
      toast('Something went wrong. Please try again later.');
    }
  }, []);

  const logister = React.useCallback(async () => {
    try {
      const {
        data: {
          user: { ID: id },
          subscription,
        },
      } = await axios.post(`${process.env.REACT_APP_API_URL}/trivai/login`, {
        phoneNumber,
        code,
      });
      setExistingPhoneNumber(phoneNumber);
      setWaitingOnVerificationCode(false);
      localStorage.setItem('user-id', id);
      localStorage.removeItem('ephemeral-user-id');
      setSubscription(subscription || {});
      setHasSubscription(!!subscription);
      toast('Successfully logged in!');
    } catch (e) {
      try {
        const {
          data: {
            user: { ID: id },
            subscription,
          },
        } = await axios.post(
          `${process.env.REACT_APP_API_URL}/trivai/register`,
          { phoneNumber, code, name: '', userId: ephemeralUserId }
        );
        setExistingPhoneNumber(phoneNumber);
        false;
        setSubscription(subscription || {});
        setHasSubscription(!!subscription);
        localStorage.setItem('user-id', id);
        localStorage.removeItem('ephemeral-user-id');
        toast('Successfully registered!');
      } catch (e) {
        toast('Something went wrong!');
      }
    }
  }, [
    setExistingPhoneNumber,
    setWaitingOnVerificationCode,
    setHasSubscription,
    toast,
  ]);

  return (
    <div className="Subscribe">
      {!hasSubscription ? (
        <div>
          <p className="Subscribe__Script">
            Want questions texted to you daily?
          </p>
          <div className="Subscribe__form">
            <div className="Subscribe__inputGroup">
              <PhoneInput
                value={phoneNumber}
                placeholder="+1 234 567 8900"
                onChange={setValue}
              />
              {!waitingOnVerificationCode && (
                <button
                  className={cx(buttonClasses)}
                  onClick={getVerificationCode}
                  disabled={!validPhoneNumber}
                >
                  Login
                </button>
              )}
            </div>
            {waitingOnVerificationCode && (
              <div className="Subscribe__inputGroup">
                <input
                  type="text"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  value={code}
                  autoComplete="one-time-code"
                  onChange={(e) => {
                    const inputValue = e.target.value;

                    if (/^\d*$/.test(inputValue) && inputValue.length <= 6) {
                      setCode(inputValue);
                    }
                  }}
                />
                <button
                  className={cx(buttonClasses)}
                  onClick={logister}
                  disabled={code.length !== 6}
                >
                  Enter code
                </button>
              </div>
            )}
          </div>
        </div>
      ) : (
        <div>
          <p className="Subscribe__Script">
            You are {subscribed ? 'subscribed to receive' : 'receiving trial'}{' '}
            daily questions at: {existingPhoneNumber}
          </p>
          {subscribed ? (
            <button
              className="Subscribe__button Button__enabled"
              onClick={unsubscribe}
            >
              Unsubscribe
            </button>
          ) : (
            <div className="Subscribe__upgradeSubscription">
              <form
                action={`${process.env.REACT_APP_API_URL}/trivai/begin-checkout`}
                method="POST"
              >
                <input type="hidden" name="userId" value={userId} />
                <button
                  className="Subscribe__button Button__enabled"
                  type="submit"
                >
                  Upgrade subscription
                </button>
              </form>
              <p>You have {7 - numFreeQuestions} free questions left.</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default Subscribe;
