import React from 'react';
import { useEffect, useState } from 'react';
import PhoneInput from 'react-phone-number-input/input';
import { isValidPhoneNumber } from 'react-phone-number-input';
import { User } from '@poltergeist/types';

export type LogisterProps = {
  logister: (phoneNumber: string, verificationCode: string, name: string) => void
  getVerificationCode: (phoneNumber: string) => void
  isRegister: boolean
  isWaitingOnVerificationCode: boolean
};
export function Logister(props: LogisterProps) {
  const { logister, getVerificationCode, isWaitingOnVerificationCode, isRegister } = props;
  const [code, setCode] = useState('');
  const [phoneNumber, setPhoneNumber] = useState<string | undefined>(undefined);
  const [name, setName] = useState<string>('');

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
          {isRegister && isWaitingOnVerificationCode && (
            <div>
              <input
                placeholder='Lebron James'
                type="text"
                value={name}
                onChange={(e) => {
                  const inputValue = e.target.value;
                  setName(inputValue);
                }}
              />
            </div>
          )}
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
              <button onClick={() => logister(phoneNumber!, code, name)} disabled={code.length !== 6}>
                {isRegister ? 'Register' : 'Enter code'}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
