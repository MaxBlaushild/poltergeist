import './Subscribe.css';
import 'react-phone-number-input/style.css';
import PhoneInput from 'react-phone-number-input/input';
import { isValidPhoneNumber } from 'react-phone-number-input';
import cx from 'classnames';
import toast from 'react-hot-toast';
import axios from 'axios';
import { getUserID } from './../util';
import * as React from 'react';

function Subscribe() {
  const [phoneNumber, setValue] = React.useState('');
  const [existingPhoneNumber, setExistingPhoneNumber] = React.useState('');
  const [waitingOnVerificationCode, setWaitingOnVerificationCode] =
    React.useState(false);
  const [code, setCode] = React.useState('');

  const validPhoneNumber =
    typeof phoneNumber === 'string' && isValidPhoneNumber(phoneNumber);
  const buttonClasses = ['Subscribe__button'];
  const { userId } = getUserID();

  const fetchUser = async () => {
    if (userId) {
      const res = await axios.get(
        `${process.env.REACT_APP_API_URL}/authenticator/users?id=${userId}`,
      );
      const {
        data: { phoneNumber: existingPhoneNumber },
      } = res;
      setExistingPhoneNumber(existingPhoneNumber);
    }
  };

  React.useEffect(() => {
    fetchUser();
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
        { phoneNumber, appName: 'Guess How Many' },
      );
      toast('Verification code sent!');
      setWaitingOnVerificationCode(true);
    } catch (e) {
      toast('Something went wrong!');
    }
  });

  const logister = React.useCallback(async () => {
    try {
      // get the user
      await axios.get(
        `${process.env.REACT_APP_API_URL}/authenticator/users?phoneNumber=` +
          encodeURIComponent(phoneNumber),
      );
      const {
        data: { ID: id },
      } = await axios.post(
        `${process.env.REACT_APP_API_URL}/authenticator/text/login`,
        { phoneNumber, code },
      );
      setExistingPhoneNumber(phoneNumber);
      setWaitingOnVerificationCode(false);
      localStorage.setItem('user-id', id);
      toast('Successfully logged in!');
    } catch (e) {
      try {
        const {
          data: { ID: id },
        } = await axios.post(
          `${process.env.REACT_APP_API_URL}/authenticator/text/register`,
          { phoneNumber, code, name: '' },
        );
        setExistingPhoneNumber(phoneNumber);
        setWaitingOnVerificationCode(false);
        localStorage.setItem('user-id', id);
        toast('Successfully registered!');
      } catch (e) {
        toast('Something went wrong!');
      }
    }
  });

  return (
    <div className="Subscribe">
      {!existingPhoneNumber ? (
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
            You are subscribed to receive daily questions at:{' '}
            {existingPhoneNumber}
          </p>
          <button
            className="Subscribe__button Button__enabled"
            onClick={() => toast("Ha! You're stuck with us.")}
          >
            Unsubscribe
          </button>
        </div>
      )}
    </div>
  );
}

export default Subscribe;
