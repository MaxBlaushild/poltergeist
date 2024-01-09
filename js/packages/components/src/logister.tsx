import React from 'react';
import { useEffect, useState } from 'react';
import PhoneInput from 'react-phone-number-input/input';
import { isValidPhoneNumber } from 'react-phone-number-input';
import { User } from '@poltergeist/types';

export type LogisterProps = {
  logister: (phoneNumber: string, verificationCode: string) => void
  getVerificationCode: (phoneNumber: string) => void
  isWaitingOnVerificationCode: boolean
};
export function Logister(props: LogisterProps) {
  const { logister, getVerificationCode, isWaitingOnVerificationCode } = props;
  const [code, setCode] = useState('');
  const [phoneNumber, setPhoneNumber] = useState<string | undefined>(undefined);
  const validPhoneNumber =
    typeof phoneNumber === 'string' && isValidPhoneNumber(phoneNumber);

  return (
    <div>
      <div>
        <p>Sign in or sign up</p>
        <div>
          <div>
            <PhoneInput
              value={phoneNumber}
              placeholder="+1 234 567 8900"
              onChange={setPhoneNumber}
            />
            {!isWaitingOnVerificationCode && (
              <button
                onClick={() => getVerificationCode(phoneNumber!)}
                disabled={!validPhoneNumber}
              >
                Login
              </button>
            )}
          </div>
          {isWaitingOnVerificationCode && (
            <div>
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
              <button onClick={() => logister(phoneNumber!, code)} disabled={code.length !== 6}>
                Enter code
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
