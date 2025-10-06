import React from 'react';
import { useEffect, useState } from 'react';
import PhoneInput from 'react-phone-number-input/input';
import { isValidPhoneNumber } from 'react-phone-number-input';
import { User } from '@poltergeist/types';

export type LogisterProps = {
  logister: (
    phoneNumber: string,
    verificationCode: string,
    name: string,
    isRegister: boolean
  ) => void;
  getVerificationCode: (phoneNumber: string) => void;
  isRegister: boolean;
  isWaitingOnVerificationCode: boolean;
  error: string | undefined;
};
export function Logister(props: LogisterProps) {
  const {
    logister,
    getVerificationCode,
    isWaitingOnVerificationCode,
    isRegister,
    error,
  } = props;
  const [code, setCode] = useState('');
  const [phoneNumber, setPhoneNumber] = useState<string | undefined>(undefined);

  const validPhoneNumber =
    typeof phoneNumber === 'string' && isValidPhoneNumber(phoneNumber);

  return (
    <div className="Logister__container">
      <div className="Logister__inputs">
        <div>
          <PhoneInput
            value={phoneNumber}
            placeholder="Phone Number"
            country="US"
            onChange={setPhoneNumber}
          />
          {isWaitingOnVerificationCode && (
            <p className="Logister__disclaimer">
              We've just sent a 6-digit verification code. It may take a moment
              to arrive.
            </p>
          )}
        </div>
        {isWaitingOnVerificationCode && (
          <div>
            <input
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              placeholder="Verification code"
              value={code}
              autoComplete="one-time-code"
              onChange={(e) => {
                const inputValue = e.target.value;

                if (/^\d*$/.test(inputValue) && inputValue.length <= 6) {
                  setCode(inputValue);
                }
              }}
            />
            {error && <p className="Logister__error">{error}</p>}
          </div>
        )}
      </div>
      <div className="Logister__buttonBar">
        {!isWaitingOnVerificationCode && validPhoneNumber ? (
          <button
            className="Logister__button"
            onClick={() => getVerificationCode(phoneNumber!)}
          >
            Get code
          </button>
        ) : null}
        {isWaitingOnVerificationCode ? (
          <button
            onClick={() => logister(phoneNumber!, code, isRegister)}
            disabled={code.length !== 6}
            className="Logister__button"
          >
            {isRegister ? 'Register' : 'Login'}
          </button>
        ) : null}
      </div>
    </div>
  );
}
